package config

import (
    "fmt"
    "io/ioutil"
    "log"
    "net/url"
    "os"
    "strconv"
    "strings"

    "gopkg.in/yaml.v2"
)

// Config 定义了整个应用的配置结构
type Config struct {
    Telegram struct {
        BotToken  string   `yaml:"bot_token"`
        Users     []string `yaml:"users"`
        Channels  []string `yaml:"channels"`
    } `yaml:"telegram"`
    RSS []RSSEntry `yaml:"rss"`
}

// RSSEntry 定义RSS配置项
type RSSEntry struct {
    URLs           []string `yaml:"urls,omitempty"`     // 新版本：支持多个URL
    URL            string   `yaml:"url,omitempty"`      // 旧版本：单个URL
    Interval       int      `yaml:"interval"`           // 更新间隔（秒）
    Keywords       []string `yaml:"keywords"`           // 关键词列表
    Group          string   `yaml:"group"`              // 分组名称
    AllowPartMatch bool     `yaml:"allow_part_match"`   // 是否允许部分匹配
}

// UnmarshalYAML 实现自定义的YAML解析逻辑，支持新旧两种格式
func (r *RSSEntry) UnmarshalYAML(unmarshal func(interface{}) error) error {
    // 定义一个临时结构体来解析YAML
    type tempRSSEntry RSSEntry
    if err := unmarshal((*tempRSSEntry)(r)); err != nil {
        return err
    }

    // 如果存在旧版本的单个URL，将其转换为URLs数组
    if r.URL != "" {
        if len(r.URLs) == 0 {
            r.URLs = []string{r.URL}
        }
        r.URL = "" // 清空旧字段
    }

    // 设置默认值：允许部分匹配
    r.AllowPartMatch = true

    return nil
}

// MarshalYAML 实现自定义的YAML序列化逻辑
func (r RSSEntry) MarshalYAML() (interface{}, error) {
    // 始终使用新格式序列化
    return struct {
        URLs           []string `yaml:"urls"`
        Interval       int      `yaml:"interval"`
        Keywords       []string `yaml:"keywords"`
        Group          string   `yaml:"group"`
        AllowPartMatch bool     `yaml:"allow_part_match"`
    }{
        URLs:           r.URLs,
        Interval:       r.Interval,
        Keywords:       r.Keywords,
        Group:          r.Group,
        AllowPartMatch: r.AllowPartMatch,
    }, nil
}

func (c *Config) Equal(other *Config) bool {
    if c.Telegram.BotToken != other.Telegram.BotToken {
        return false
    }
    if !stringSliceEqual(c.Telegram.Users, other.Telegram.Users) {
        return false
    }
    if !stringSliceEqual(c.Telegram.Channels, other.Telegram.Channels) {
        return false
    }
    if len(c.RSS) != len(other.RSS) {
        return false
    }
    for i := range c.RSS {
        if !stringSliceEqual(c.RSS[i].URLs, other.RSS[i].URLs) {
            return false
        }
        if c.RSS[i].Interval != other.RSS[i].Interval ||
           c.RSS[i].Group != other.RSS[i].Group ||
           !stringSliceEqual(c.RSS[i].Keywords, other.RSS[i].Keywords) {
            return false
        }
    }
    return true
}

func stringSliceEqual(a, b []string) bool {
    if len(a) != len(b) {
        return false
    }
    for i := range a {
        if a[i] != b[i] {
            return false
        }
    }
    return true
}

func Load(path string) (*Config, error) {
    log.Printf("正在加载配置文件: %s", path)
    
    var config Config
    
    // 尝试读取配置文件
    data, err := ioutil.ReadFile(path)
    if err != nil {
        if !os.IsNotExist(err) {
            return nil, fmt.Errorf("读取配置文件失败: %v", err)
        }
        // 如果文件不存在，创建一个空的配置
        config = Config{}
    } else {
        // 解析已存在的配置文件
        if err := yaml.Unmarshal(data, &config); err != nil {
            return nil, fmt.Errorf("解析配置文件失败: %v", err)
        }
    }

    // 从环境变量补充缺失的配置
    configChanged := false

    // 检查并补充 Telegram 配置
    if config.Telegram.BotToken == "" {
        if token := os.Getenv("TELEGRAM_BOT_TOKEN"); token != "" {
            config.Telegram.BotToken = token
            configChanged = true
        }
    }

    if len(config.Telegram.Users) == 0 {
        if users := os.Getenv("TELEGRAM_USERS"); users != "" {
            config.Telegram.Users = strings.Split(users, ",")
            configChanged = true
        }
    }

    if len(config.Telegram.Channels) == 0 {
        if channels := os.Getenv("TELEGRAM_CHANNELS"); channels != "" {
            config.Telegram.Channels = strings.Split(channels, ",")
            configChanged = true
        }
    }

    // 检查并补充 RSS 配置
    if len(config.RSS) == 0 {
        if rssURLs := os.Getenv("RSS_URLS"); rssURLs != "" {
            urlGroups := strings.Split(rssURLs, ";") // 使用分号分隔不同的RSS组
            for i, urlGroup := range urlGroups {
                entry := RSSEntry{
                    URLs:     strings.Split(strings.TrimSpace(urlGroup), ","),
                    Interval: 300, // 默认5分钟
                    Group:    "默认分组",
                }
                
                // 尝试加载对应的关键词
                if keywords := os.Getenv(fmt.Sprintf("RSS_KEYWORDS_%d", i)); keywords != "" {
                    entry.Keywords = strings.Split(keywords, ",")
                }
                
                // 尝试加载对应的间隔时间
                if interval := os.Getenv(fmt.Sprintf("RSS_INTERVAL_%d", i)); interval != "" {
                    if i, err := strconv.Atoi(interval); err == nil && i > 0 {
                        entry.Interval = i
                    }
                }
                
                // 尝试加载对应的分组
                if group := os.Getenv(fmt.Sprintf("RSS_GROUP_%d", i)); group != "" {
                    entry.Group = group
                }

                config.RSS = append(config.RSS, entry)
            }
            configChanged = true
        }
    }

    // 如果配置有变化，保存到文件
    if configChanged {
        log.Println("从环境变量补充了配置信息，正在保存到配置文件")
        if err := config.Save(path); err != nil {
            log.Printf("警告：保存补充的配置到文件失败: %v", err)
        }
    }

    // 验证和清理配置
    if err := validateAndCleanConfig(&config); err != nil {
        return nil, fmt.Errorf("配置验证失败: %v", err)
    }

    log.Printf("成功加载配置文件")
    return &config, nil
}

func validateAndCleanConfig(config *Config) error {
    // 验证Telegram配置
    if config.Telegram.BotToken == "" {
        return fmt.Errorf("未设置bot_token")
    }
    if len(config.Telegram.Users) == 0 {
        return fmt.Errorf("未设置用户列表")
    }

    // 验证和清理RSS配置
    for i := range config.RSS {
        // 验证URLs
        if len(config.RSS[i].URLs) == 0 {
            return fmt.Errorf("RSS #%d: URLs为空", i+1)
        }
        
        // 验证每个URL的格式
        for j, urlStr := range config.RSS[i].URLs {
            urlStr = strings.TrimSpace(urlStr)
            if urlStr == "" {
                return fmt.Errorf("RSS #%d: URL #%d 为空", i+1, j+1)
            }
            
            if _, err := url.Parse(urlStr); err != nil {
                return fmt.Errorf("RSS #%d: URL #%d 格式无效: %v", i+1, j+1, err)
            }
            config.RSS[i].URLs[j] = urlStr // 保存清理后的URL
        }

        // 设置默认间隔时间
        if config.RSS[i].Interval <= 0 {
            config.RSS[i].Interval = 300 // 默认5分钟
        }

        // 设置默认分组
        if config.RSS[i].Group == "" {
            config.RSS[i].Group = "默认分组"
        }

        // 清理关键词列表
        cleanKeywords := make([]string, 0)
        for _, keyword := range config.RSS[i].Keywords {
            keyword = strings.TrimSpace(keyword)
            if keyword != "" {
                cleanKeywords = append(cleanKeywords, keyword)
            }
        }
        config.RSS[i].Keywords = cleanKeywords
    }

    return nil
}

func LoadFromEnv() *Config {
    config := &Config{}
    
    // 加载Telegram配置
    config.Telegram.BotToken = os.Getenv("TELEGRAM_BOT_TOKEN")
    if users := os.Getenv("TELEGRAM_USERS"); users != "" {
        config.Telegram.Users = strings.Split(users, ",")
    }
    if channels := os.Getenv("TELEGRAM_CHANNELS"); channels != "" {
        config.Telegram.Channels = strings.Split(channels, ",")
    }

    // 加载RSS配置
    if rssURLs := os.Getenv("RSS_URLS"); rssURLs != "" {
        urlGroups := strings.Split(rssURLs, ";") // 使用分号分隔不同的RSS组
        config.RSS = make([]RSSEntry, len(urlGroups))

        for i, urlGroup := range urlGroups {
            config.RSS[i] = RSSEntry{
                URLs:     strings.Split(strings.TrimSpace(urlGroup), ","),
                Interval: 300, // 默认5分钟
                Group:    "默认分组",
            }
            
            // 尝试加载对应的关键词
            if keywords := os.Getenv(fmt.Sprintf("RSS_KEYWORDS_%d", i)); keywords != "" {
                config.RSS[i].Keywords = strings.Split(keywords, ",")
            }
            
            // 尝试加载对应的间隔时间
            if interval := os.Getenv(fmt.Sprintf("RSS_INTERVAL_%d", i)); interval != "" {
                if i, err := strconv.Atoi(interval); err == nil && i > 0 {
                    config.RSS[i].Interval = i
                }
            }
            
            // 尝试加载对应的分组
            if group := os.Getenv(fmt.Sprintf("RSS_GROUP_%d", i)); group != "" {
                config.RSS[i].Group = group
            }
        }
    }

    return config
}

func (c *Config) Save(filename string) error {
    data, err := yaml.Marshal(c)
    if err != nil {
        return fmt.Errorf("序列化配置失败: %v", err)
    }
    return ioutil.WriteFile(filename, data, 0644)
}
