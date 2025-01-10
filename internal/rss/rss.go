package rss

import (
    "log"
    "math/rand"
    "net/http"
    "strings"
    "sync"
    "time"
    "fmt"

    "github.com/mmcdole/gofeed"
    "rss2telegram/internal/storage"
)

var userAgents = []string{
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Edge/120.0.0.0 Safari/537.36",
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0",
    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Safari/605.1.15",
}

func getRandomUserAgent() string {
    return userAgents[rand.Intn(len(userAgents))]
}

type MessageHandler func(title, url, group string, pubDate time.Time, matchedKeywords []string) error

type Manager struct {
    feeds          []*Feed
    db             *storage.Storage
    messageHandler MessageHandler
    mu             sync.Mutex
}

type Feed struct {
    URL      string
    Interval time.Duration
    Keywords []string
    Group    string
    ticker   *time.Ticker
    stopChan chan struct{}
}

type Config struct {
    URL      string
    Interval int
    Keywords []string
    Group    string
}

func NewManager(configs []Config, db *storage.Storage) *Manager {
    manager := &Manager{
        db: db,
    }
    manager.UpdateFeeds(configs)
    return manager
}

func (m *Manager) SetMessageHandler(handler MessageHandler) {
    m.messageHandler = handler
}

func (m *Manager) UpdateFeeds(configs []Config) {
    m.mu.Lock()
    defer m.mu.Unlock()

    // 停止所有现有的feed轮询器
    for _, feed := range m.feeds {
        if feed.stopChan != nil {
            close(feed.stopChan)
        }
    }

    // 创建新的feeds
    m.feeds = make([]*Feed, len(configs))
    for i, config := range configs {
        m.feeds[i] = &Feed{
            URL:      config.URL,
            Interval: time.Duration(config.Interval) * time.Second,
            Keywords: config.Keywords,
            Group:    config.Group,
            stopChan: make(chan struct{}),
        }
    }

    // 启动新的feed轮询器
    for _, feed := range m.feeds {
        go m.pollFeed(feed)
    }
}

func (m *Manager) Start() {
    log.Println("RSS管理器已启动")
    // 实际的轮询现在在UpdateFeeds中处理
}

func (m *Manager) pollFeed(feed *Feed) {
    feed.ticker = time.NewTicker(feed.Interval)
    defer feed.ticker.Stop()

    for {
        select {
        case <-feed.ticker.C:
            log.Printf("检查feed: %s", feed.URL)
            m.checkFeed(feed)
        case <-feed.stopChan:
            log.Printf("停止feed轮询器: %s", feed.URL)
            return
        }
    }
}

func (m *Manager) checkFeed(feed *Feed) {
    fp := gofeed.NewParser()
    
    // 创建自定义的 HTTP 客户端
    client := &http.Client{
        Timeout: 30 * time.Second,
        Transport: &http.Transport{
            MaxIdleConns:        100,
            IdleConnTimeout:     90 * time.Second,
            DisableCompression:  true,
            TLSHandshakeTimeout: 10 * time.Second,
        },
    }

    // 最多重试3次
    maxRetries := 3
    var lastErr error
    
    for retry := 0; retry < maxRetries; retry++ {
        if retry > 0 {
            // 重试间隔随机化，避免固定间隔
            time.Sleep(time.Duration(rand.Intn(5)+1) * time.Second)
        }
        
        // 创建自定义的请求
        req, err := http.NewRequest("GET", feed.URL, nil)
        if err != nil {
            lastErr = fmt.Errorf("创建请求失败: %v", err)
            continue
        }
        
        // 添加更完整的请求头
        req.Header.Set("User-Agent", getRandomUserAgent())
        req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
        req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
        req.Header.Set("Accept-Encoding", "gzip, deflate, br")
        req.Header.Set("Connection", "keep-alive")
        req.Header.Set("Cache-Control", "max-age=0")
        req.Header.Set("Upgrade-Insecure-Requests", "1")
        req.Header.Set("Sec-Fetch-Dest", "document")
        req.Header.Set("Sec-Fetch-Mode", "navigate")
        req.Header.Set("Sec-Fetch-Site", "none")
        req.Header.Set("Sec-Fetch-User", "?1")
        req.Header.Set("DNT", "1")
        
        // 如果是 hostloc 域名，添加特殊处理
        if strings.Contains(feed.URL, "hostloc.com") {
            req.Header.Set("Referer", "https://hostloc.com/")
            req.Header.Set("Origin", "https://hostloc.com")
            // 添加一个随机的 Cookie
            req.Header.Set("Cookie", fmt.Sprintf("_ga=GA1.%d.%d.%d", 
                rand.Intn(999999999), 
                rand.Intn(999999999), 
                time.Now().Unix()))
        }
        
        // 使用自定义客户端解析 Feed
        fp.Client = client
        parsedFeed, err := fp.ParseURL(feed.URL)
        if err != nil {
            lastErr = fmt.Errorf("解析Feed失败: %v", err)
            log.Printf("第 %d 次尝试解析Feed %s失败: %v", retry+1, feed.URL, err)
            continue
        }
        
        // 成功获取数据，处理 Feed 内容
        for _, item := range parsedFeed.Items {
            matchedKeywords := m.matchKeywords(item, feed.Keywords)
            if len(matchedKeywords) > 0 {
                log.Printf("发现新项目: %s", item.Title)
                if err := m.messageHandler(item.Title, item.Link, feed.Group, *item.PublishedParsed, matchedKeywords); err != nil {
                    log.Printf("发送消息失败: %v", err)
                } else {
                    log.Printf("成功发送项目的消息: %s", item.Title)
                    m.db.MarkAsSent(item.Link)
                }
            }
        }
        
        // 如果成功，直接返回
        return
    }
    
    // 所有重试都失败后记录最后的错误
    if lastErr != nil {
        log.Printf("在 %d 次尝试后仍然无法解析Feed %s: %v", maxRetries, feed.URL, lastErr)
    }
}

func (m *Manager) matchKeywords(item *gofeed.Item, keywords []string) []string {
    if m.db.WasSent(item.Link) {
        return nil
    }

    if len(keywords) == 0 {
        return []string{"无关键词"}
    }

    var matched []string
    for _, keyword := range keywords {
        if strings.Contains(strings.ToLower(item.Title), strings.ToLower(keyword)) || 
           strings.Contains(strings.ToLower(item.Description), strings.ToLower(keyword)) {
            matched = append(matched, keyword)
        }
    }

    return matched
}

func init() {
    // 初始化随机数生成器
    rand.Seed(time.Now().UnixNano())
}
