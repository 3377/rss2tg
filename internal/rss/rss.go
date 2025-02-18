package rss

import (
    "log"
    "net/http"
    "strings"
    "sync"
    "time"
    "unicode"

    "github.com/mmcdole/gofeed"
    "rss2tg/internal/storage"
)

type MessageHandler func(title, url, group string, pubDate time.Time, matchedKeywords []string) error

type Manager struct {
    feeds          []*Feed
    db             *storage.Storage
    messageHandler MessageHandler
    mu             sync.Mutex
}

type Feed struct {
    URLs            []string
    Interval        time.Duration
    Keywords        []string
    Group           string
    AllowPartMatch  bool      // 是否允许部分匹配
    ticker          *time.Ticker
    stopChan        chan struct{}
}

type Config struct {
    URLs            []string
    Interval        int
    Keywords        []string
    Group           string
    AllowPartMatch  bool      // 是否允许部分匹配
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
            URLs:           config.URLs,
            Interval:       time.Duration(config.Interval) * time.Second,
            Keywords:       config.Keywords,
            Group:          config.Group,
            AllowPartMatch: config.AllowPartMatch,  // 添加部分匹配配置
            stopChan:       make(chan struct{}),
        }
    }

    // 启动新的feed轮询器
    for _, feed := range m.feeds {
        go m.pollFeed(feed)
    }
}

func (m *Manager) Start() {
    log.Println("RSS管理器已启动")
}

func (m *Manager) pollFeed(feed *Feed) {
    feed.ticker = time.NewTicker(feed.Interval)
    defer feed.ticker.Stop()

    for {
        select {
        case <-feed.ticker.C:
            for _, url := range feed.URLs {
                log.Printf("检查feed: %s", url)
                m.checkFeed(feed, url)
            }
        case <-feed.stopChan:
            log.Printf("停止feed轮询器: %v", feed.URLs)
            return
        }
    }
}

func (m *Manager) checkFeed(feed *Feed, url string) {
    fp := gofeed.NewParser()
    
    // 创建自定义的 HTTP 客户端
    client := &http.Client{
        Timeout: 30 * time.Second,
    }
    
    // 创建自定义的请求
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        log.Printf("创建请求失败 %s: %v", url, err)
        return
    }
    
    // 添加浏览器标识和其他必要的头信息
    req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
    req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
    req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
    req.Header.Set("Connection", "keep-alive")
    req.Header.Set("Upgrade-Insecure-Requests", "1")
    
    // 使用自定义客户端解析 Feed
    fp.Client = client
    parsedFeed, err := fp.ParseURL(url)
    if err != nil {
        log.Printf("解析Feed %s失败: %v", url, err)
        return
    }

    for _, item := range parsedFeed.Items {
        matchedKeywords := m.matchKeywords(item, feed)
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
}

// normalizeText 标准化文本，处理特殊字符和空白
func normalizeText(text string) string {
    // 1. 转换为小写
    text = strings.ToLower(text)
    
    // 2. 替换常见的特殊字符组合
    replacements := map[string]string{
        "c++": "cpp",
        "c#": "csharp",
        ".net": "dotnet",
    }
    
    for old, new := range replacements {
        text = strings.ReplaceAll(text, old, new)
    }
    
    // 3. 清理特殊字符，保留字母、数字和空格
    var result strings.Builder
    for _, ch := range text {
        if unicode.IsLetter(ch) || unicode.IsNumber(ch) || unicode.IsSpace(ch) {
            result.WriteRune(ch)
        } else {
            // 用空格替换特殊字符
            result.WriteRune(' ')
        }
    }
    
    // 4. 规范化空白字符
    return strings.Join(strings.Fields(result.String()), " ")
}

// isWordMatch 检查单词是否完全匹配
func isWordMatch(text, keyword string) bool {
    words := strings.Fields(text)
    for _, word := range words {
        if word == keyword {
            return true
        }
    }
    return false
}

// contains 检查字符串切片是否包含特定字符串
func contains(slice []string, str string) bool {
    for _, v := range slice {
        if v == str {
            return true
        }
    }
    return false
}

func (m *Manager) matchKeywords(item *gofeed.Item, feed *Feed) []string {
    if m.db.WasSent(item.Link) {
        return nil
    }

    if len(feed.Keywords) == 0 {
        return []string{"无关键词"}
    }

    // 标准化文本
    normalizedTitle := normalizeText(item.Title)
    normalizedDesc := normalizeText(item.Description)
    
    var matched []string
    
    // 检查每个关键词
    for _, keyword := range feed.Keywords {
        // 标准化关键词
        normalizedKeyword := normalizeText(keyword)
        
        // 首先尝试完整词匹配
        if isWordMatch(normalizedTitle, normalizedKeyword) {
            if !contains(matched, keyword) {
                matched = append(matched, keyword)
            }
            continue
        }
        
        if isWordMatch(normalizedDesc, normalizedKeyword) {
            if !contains(matched, keyword) {
                matched = append(matched, keyword)
            }
            continue
        }
        
        // 如果允许部分匹配且没有找到完整匹配，尝试部分匹配
        if feed.AllowPartMatch {
            if strings.Contains(normalizedTitle, normalizedKeyword) {
                if !contains(matched, keyword) {
                    matched = append(matched, keyword)
                }
            } else if strings.Contains(normalizedDesc, normalizedKeyword) {
                if !contains(matched, keyword) {
                    matched = append(matched, keyword)
                }
            }
        }
    }

    // 根据是否匹配到关键词来决定日志输出级别
    if len(matched) > 0 {
        // 如果匹配到关键词，输出详细日志
        log.Printf("📝 发现匹配文章:\n"+
            "   标题: %s\n"+
            "   描述: %s\n"+
            "   链接: %s\n"+
            "   部分匹配: %s\n"+
            "✨ 匹配关键词: %v",
            item.Title,
            item.Description,
            item.Link,
            map[bool]string{true: "允许", false: "禁用"}[feed.AllowPartMatch],
            matched)
    } else {
        // 如果未匹配到关键词，只输出简单的监听状态
        log.Printf("👀 监听RSS: %s, 标题: %s", feed.URLs[0], item.Title)
    }

    return matched
}
