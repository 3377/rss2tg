package config

import (
    "fmt"
    "io/ioutil"
    "log"
    "net/url"
    "os"
    "strconv"
    "strings"
    "path/filepath"

    "gopkg.in/yaml.v2"
)

// Config 定义了整个应用的配置结构
type Config struct {
    Telegram struct {
        BotToken    string   `yaml:"bot_token"`
        Users       []string `yaml:"users"`
        Channels    []string `yaml:"channels"`
        AdminUsers  []string `yaml:"adminuser,omitempty"`  // 管理员用户ID列表
    } `yaml:"telegram"`
    Webhook struct {
        Enabled    bool   `yaml:"enabled"`      // 是否启用 webhook 推送（向后兼容）
        URL        string `yaml:"url"`          // webhook 地址（向后兼容）
        Timeout    int    `yaml:"timeout"`      // 请求超时时间（秒）（向后兼容）
        RetryCount int    `yaml:"retry_count"`  // 失败重试次数（向后兼容）
    } `yaml:"webhook"`
    Webhooks []WebhookEntry `yaml:"webhooks,omitempty"` // 多个 webhook 配置
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
    Enabled        bool     `yaml:"enabled"`            // 是否启用此订阅
}

// WebhookEntry 定义单个 webhook 配置项
type WebhookEntry struct {
    Name       string `yaml:"name"`         // webhook 名称
    Enabled    bool   `yaml:"enabled"`      // 是否启用此 webhook
    URL        string `yaml:"url"`          // webhook 地址
    Timeout    int    `yaml:"timeout"`      // 请求超时时间（秒）
    RetryCount int    `yaml:"retry_count"`  // 失败重试次数
}

// UnmarshalYAML 实现自定义的YAML解析逻辑，支持新旧两种格式
func (r *RSSEntry) UnmarshalYAML(unmarshal func(interface{}) error) error {
    // 定义一个临时结构体来解析YAML
    type tempRSSEntry struct {
        URLs           []string `yaml:"urls,omitempty"`
        URL            string   `yaml:"url,omitempty"`
        Interval       int      `yaml:"interval"`
        Keywords       []string `yaml:"keywords"`
        Group          string   `yaml:"group"`
        AllowPartMatch *bool    `yaml:"allow_part_match,omitempty"`  // 使用指针类型
        Enabled        *bool    `yaml:"enabled,omitempty"`          // 使用指针类型
    }

    // 解析配置到临时结构体
    var temp tempRSSEntry
    if err := unmarshal(&temp); err != nil {
        return err
    }

    // 复制字段到实际结构体
    r.URLs = temp.URLs
    r.URL = temp.URL
    r.Interval = temp.Interval
    r.Keywords = temp.Keywords
    r.Group = temp.Group

    // 如果存在旧版本的单个URL，将其转换为URLs数组
    if r.URL != "" {
        if len(r.URLs) == 0 {
            r.URLs = []string{r.URL}
        }
        r.URL = "" // 清空旧字段
    }

    // 只在未设置值时使用默认值
    if temp.AllowPartMatch == nil {
        r.AllowPartMatch = true  // 默认值
    } else {
        r.AllowPartMatch = *temp.AllowPartMatch
    }

    if temp.Enabled == nil {
        r.Enabled = true // 默认值
    } else {
        r.Enabled = *temp.Enabled
    }

    return nil
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
    // 检查 webhook 配置
    if c.Webhook.Enabled != other.Webhook.Enabled ||
       c.Webhook.URL != other.Webhook.URL ||
       c.Webhook.Timeout != other.Webhook.Timeout ||
       c.Webhook.RetryCount != other.Webhook.RetryCount {
        return false
    }
    // 检查 webhooks 配置
    if len(c.Webhooks) != len(other.Webhooks) {
        return false
    }
    for i := range c.Webhooks {
        if c.Webhooks[i].Name != other.Webhooks[i].Name ||
           c.Webhooks[i].Enabled != other.Webhooks[i].Enabled ||
           c.Webhooks[i].URL != other.Webhooks[i].URL ||
           c.Webhooks[i].Timeout != other.Webhooks[i].Timeout ||
           c.Webhooks[i].RetryCount != other.Webhooks[i].RetryCount {
            return false
        }
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

    // 检查并补充管理员配置
    if len(config.Telegram.AdminUsers) == 0 {
        if adminUsers := os.Getenv("TELEGRAM_ADMIN_USERS"); adminUsers != "" {
            config.Telegram.AdminUsers = strings.Split(adminUsers, ",")
            configChanged = true
        }
    }

    // 检查并补充 Webhook 配置
    if config.Webhook.URL == "" {
        if webhookURL := os.Getenv("WEBHOOK_URL"); webhookURL != "" {
            config.Webhook.URL = webhookURL
            configChanged = true
        }
    }
    
    if !config.Webhook.Enabled {
        if webhookEnabled := os.Getenv("WEBHOOK_ENABLED"); webhookEnabled != "" {
            if webhookEnabled == "true" || webhookEnabled == "1" {
                config.Webhook.Enabled = true
                configChanged = true
            }
        }
    }
    
    if config.Webhook.Timeout == 0 {
        if webhookTimeout := os.Getenv("WEBHOOK_TIMEOUT"); webhookTimeout != "" {
            if timeout, err := strconv.Atoi(webhookTimeout); err == nil && timeout > 0 {
                config.Webhook.Timeout = timeout
                configChanged = true
            }
        }
        // 设置默认超时时间
        if config.Webhook.Timeout == 0 {
            config.Webhook.Timeout = 10
        }
    }
    
    if config.Webhook.RetryCount == 0 {
        if webhookRetryCount := os.Getenv("WEBHOOK_RETRY_COUNT"); webhookRetryCount != "" {
            if retryCount, err := strconv.Atoi(webhookRetryCount); err == nil && retryCount >= 0 {
                config.Webhook.RetryCount = retryCount
                configChanged = true
            }
        }
        // 设置默认重试次数
        if config.Webhook.RetryCount == 0 {
            config.Webhook.RetryCount = 3
        }
    }

    // 检查并补充多个 Webhooks 配置
    if len(config.Webhooks) == 0 {
        // 扫描环境变量，查找 WEBHOOK_URL_1, WEBHOOK_URL_2 等
        webhookEntries := make([]WebhookEntry, 0)
        
        for i := 1; i <= 10; i++ { // 最多支持10个webhook
            urlEnvKey := fmt.Sprintf("WEBHOOK_URL_%d", i)
            if webhookURL := os.Getenv(urlEnvKey); webhookURL != "" {
                entry := WebhookEntry{
                    Name:       fmt.Sprintf("webhook-%d", i),
                    Enabled:    true,
                    URL:        webhookURL,
                    Timeout:    10, // 默认10秒
                    RetryCount: 3,  // 默认重试3次
                }
                
                // 加载对应的名称
                if name := os.Getenv(fmt.Sprintf("WEBHOOK_NAME_%d", i)); name != "" {
                    entry.Name = name
                }
                
                // 加载对应的启用状态
                if enabled := os.Getenv(fmt.Sprintf("WEBHOOK_ENABLED_%d", i)); enabled != "" {
                    if enabled == "false" || enabled == "0" {
                        entry.Enabled = false
                    }
                }
                
                // 加载对应的超时时间
                if timeout := os.Getenv(fmt.Sprintf("WEBHOOK_TIMEOUT_%d", i)); timeout != "" {
                    if parsedTimeout, err := strconv.Atoi(timeout); err == nil && parsedTimeout > 0 {
                        entry.Timeout = parsedTimeout
                    }
                }
                
                // 加载对应的重试次数
                if retryCount := os.Getenv(fmt.Sprintf("WEBHOOK_RETRY_COUNT_%d", i)); retryCount != "" {
                    if parsedRetryCount, err := strconv.Atoi(retryCount); err == nil && parsedRetryCount >= 0 {
                        entry.RetryCount = parsedRetryCount
                    }
                }
                
                webhookEntries = append(webhookEntries, entry)
            }
        }
        
        if len(webhookEntries) > 0 {
            config.Webhooks = webhookEntries
            configChanged = true
        }
    }

    // 检查并补充 RSS 配置
    if len(config.RSS) == 0 {
        // 先尝试新的格式：RSS_URLS_1, RSS_URLS_2 等
        rssEntries := make([]RSSEntry, 0)
        
        // 扫描环境变量，查找RSS_URLS_1, RSS_URLS_2 等
        for i := 1; i <= 10; i++ { // 最多支持10个RSS源
            urlsEnvKey := fmt.Sprintf("RSS_URLS_%d", i)
            if urls := os.Getenv(urlsEnvKey); urls != "" {
                entry := RSSEntry{
                    URLs:           strings.Split(strings.TrimSpace(urls), ","),
                    Interval:       300, // 默认5分钟
                    Group:          "默认分组",
                    AllowPartMatch: true, // 默认允许部分匹配
                    Enabled:        true, // 默认启用
                }
                
                // 加载对应的间隔时间
                if interval := os.Getenv(fmt.Sprintf("RSS_INTERVAL_%d", i)); interval != "" {
                    if parsedInterval, err := strconv.Atoi(interval); err == nil && parsedInterval > 0 {
                        entry.Interval = parsedInterval
                    }
                }
                
                // 加载对应的关键词
                if keywords := os.Getenv(fmt.Sprintf("RSS_KEYWORDS_%d", i)); keywords != "" {
                    if strings.TrimSpace(keywords) != "" {
                        entry.Keywords = strings.Split(keywords, ",")
                        // 清理关键词列表
                        cleanKeywords := make([]string, 0)
                        for _, keyword := range entry.Keywords {
                            keyword = strings.TrimSpace(keyword)
                            if keyword != "" {
                                cleanKeywords = append(cleanKeywords, keyword)
                            }
                        }
                        entry.Keywords = cleanKeywords
                    }
                }
                
                // 加载对应的分组
                if group := os.Getenv(fmt.Sprintf("RSS_GROUP_%d", i)); group != "" {
                    entry.Group = group
                }
                
                // 加载allow_part_match设置
                if allowPartMatch := os.Getenv(fmt.Sprintf("RSS_ALLOW_PART_MATCH_%d", i)); allowPartMatch != "" {
                    if allowPartMatch == "false" || allowPartMatch == "0" {
                        entry.AllowPartMatch = false
                    } else {
                        entry.AllowPartMatch = true
                    }
                }
                
                // 加载启用状态
                if enabled := os.Getenv(fmt.Sprintf("RSS_ENABLED_%d", i)); enabled != "" {
                    if enabled == "false" || enabled == "0" {
                        entry.Enabled = false
                    }
                }
                
                rssEntries = append(rssEntries, entry)
            }
        }
        
        // 如果找到了新格式的RSS配置，使用它们
        if len(rssEntries) > 0 {
            config.RSS = rssEntries
            configChanged = true
        } else {
            // 如果没有找到新格式，尝试旧格式：RSS_URLS
            if rssURLs := os.Getenv("RSS_URLS"); rssURLs != "" {
                urlGroups := strings.Split(rssURLs, ";") // 使用分号分隔不同的RSS组
                for i, urlGroup := range urlGroups {
                    entry := RSSEntry{
                        URLs:           strings.Split(strings.TrimSpace(urlGroup), ","),
                        Interval:       300, // 默认5分钟
                        Group:          "默认分组",
                        AllowPartMatch: true,
                        Enabled:        true, // 默认启用
                    }
                    
                    // 尝试加载对应的关键词（注意：旧格式使用0开始的索引）
                    if keywords := os.Getenv(fmt.Sprintf("RSS_KEYWORDS_%d", i)); keywords != "" {
                        entry.Keywords = strings.Split(keywords, ",")
                    }
                    
                    // 尝试加载对应的间隔时间
                    if interval := os.Getenv(fmt.Sprintf("RSS_INTERVAL_%d", i)); interval != "" {
                        if parsedInterval, err := strconv.Atoi(interval); err == nil && parsedInterval > 0 {
                            entry.Interval = parsedInterval
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
    if adminUsers := os.Getenv("TELEGRAM_ADMIN_USERS"); adminUsers != "" {
        config.Telegram.AdminUsers = strings.Split(adminUsers, ",")
    }

    // 加载 Webhook 配置
    if webhookEnabled := os.Getenv("WEBHOOK_ENABLED"); webhookEnabled != "" {
        if webhookEnabled == "true" || webhookEnabled == "1" {
            config.Webhook.Enabled = true
        }
    }
    
    if webhookURL := os.Getenv("WEBHOOK_URL"); webhookURL != "" {
        config.Webhook.URL = webhookURL
    }
    
    if webhookTimeout := os.Getenv("WEBHOOK_TIMEOUT"); webhookTimeout != "" {
        if timeout, err := strconv.Atoi(webhookTimeout); err == nil && timeout > 0 {
            config.Webhook.Timeout = timeout
        }
    }
    // 设置默认超时时间
    if config.Webhook.Timeout == 0 {
        config.Webhook.Timeout = 10
    }
    
    if webhookRetryCount := os.Getenv("WEBHOOK_RETRY_COUNT"); webhookRetryCount != "" {
        if retryCount, err := strconv.Atoi(webhookRetryCount); err == nil && retryCount >= 0 {
            config.Webhook.RetryCount = retryCount
        }
    }
    // 设置默认重试次数
    if config.Webhook.RetryCount == 0 {
        config.Webhook.RetryCount = 3
    }

    // 加载多个 Webhooks 配置
    webhookEntries := make([]WebhookEntry, 0)
    
    for i := 1; i <= 10; i++ { // 最多支持10个webhook
        urlEnvKey := fmt.Sprintf("WEBHOOK_URL_%d", i)
        if webhookURL := os.Getenv(urlEnvKey); webhookURL != "" {
            entry := WebhookEntry{
                Name:       fmt.Sprintf("webhook-%d", i),
                Enabled:    true,
                URL:        webhookURL,
                Timeout:    10, // 默认10秒
                RetryCount: 3,  // 默认重试3次
            }
            
            // 加载对应的名称
            if name := os.Getenv(fmt.Sprintf("WEBHOOK_NAME_%d", i)); name != "" {
                entry.Name = name
            }
            
            // 加载对应的启用状态
            if enabled := os.Getenv(fmt.Sprintf("WEBHOOK_ENABLED_%d", i)); enabled != "" {
                if enabled == "false" || enabled == "0" {
                    entry.Enabled = false
                }
            }
            
            // 加载对应的超时时间
            if timeout := os.Getenv(fmt.Sprintf("WEBHOOK_TIMEOUT_%d", i)); timeout != "" {
                if parsedTimeout, err := strconv.Atoi(timeout); err == nil && parsedTimeout > 0 {
                    entry.Timeout = parsedTimeout
                }
            }
            
            // 加载对应的重试次数
            if retryCount := os.Getenv(fmt.Sprintf("WEBHOOK_RETRY_COUNT_%d", i)); retryCount != "" {
                if parsedRetryCount, err := strconv.Atoi(retryCount); err == nil && parsedRetryCount >= 0 {
                    entry.RetryCount = parsedRetryCount
                }
            }
            
            webhookEntries = append(webhookEntries, entry)
        }
    }
    
    if len(webhookEntries) > 0 {
        config.Webhooks = webhookEntries
    }

    // 加载RSS配置 - 优先使用新格式RSS_URLS_1, RSS_URLS_2等
    rssEntries := make([]RSSEntry, 0)
    
    // 扫描环境变量，查找RSS_URLS_1, RSS_URLS_2 等
    for i := 1; i <= 10; i++ { // 最多支持10个RSS源
        urlsEnvKey := fmt.Sprintf("RSS_URLS_%d", i)
        if urls := os.Getenv(urlsEnvKey); urls != "" {
            entry := RSSEntry{
                URLs:           strings.Split(strings.TrimSpace(urls), ","),
                Interval:       300, // 默认5分钟
                Group:          "默认分组",
                AllowPartMatch: true, // 默认允许部分匹配
                Enabled:        true, // 默认启用
            }
            
            // 加载对应的间隔时间
            if interval := os.Getenv(fmt.Sprintf("RSS_INTERVAL_%d", i)); interval != "" {
                if parsedInterval, err := strconv.Atoi(interval); err == nil && parsedInterval > 0 {
                    entry.Interval = parsedInterval
                }
            }
            
            // 加载对应的关键词
            if keywords := os.Getenv(fmt.Sprintf("RSS_KEYWORDS_%d", i)); keywords != "" {
                if strings.TrimSpace(keywords) != "" {
                    entry.Keywords = strings.Split(keywords, ",")
                    // 清理关键词列表
                    cleanKeywords := make([]string, 0)
                    for _, keyword := range entry.Keywords {
                        keyword = strings.TrimSpace(keyword)
                        if keyword != "" {
                            cleanKeywords = append(cleanKeywords, keyword)
                        }
                    }
                    entry.Keywords = cleanKeywords
                }
            }
            
            // 加载对应的分组
            if group := os.Getenv(fmt.Sprintf("RSS_GROUP_%d", i)); group != "" {
                entry.Group = group
            }
            
            // 加载allow_part_match设置
            if allowPartMatch := os.Getenv(fmt.Sprintf("RSS_ALLOW_PART_MATCH_%d", i)); allowPartMatch != "" {
                if allowPartMatch == "false" || allowPartMatch == "0" {
                    entry.AllowPartMatch = false
                } else {
                    entry.AllowPartMatch = true
                }
            }
            
            // 加载启用状态
            if enabled := os.Getenv(fmt.Sprintf("RSS_ENABLED_%d", i)); enabled != "" {
                if enabled == "false" || enabled == "0" {
                    entry.Enabled = false
                }
            }
            
            rssEntries = append(rssEntries, entry)
        }
    }
    
    // 如果找到了新格式的RSS配置，使用它们
    if len(rssEntries) > 0 {
        config.RSS = rssEntries
    } else {
        // 如果没有找到新格式，尝试旧格式：RSS_URLS
        if rssURLs := os.Getenv("RSS_URLS"); rssURLs != "" {
            urlGroups := strings.Split(rssURLs, ";") // 使用分号分隔不同的RSS组
            config.RSS = make([]RSSEntry, len(urlGroups))

            for i, urlGroup := range urlGroups {
                config.RSS[i] = RSSEntry{
                    URLs:           strings.Split(strings.TrimSpace(urlGroup), ","),
                    Interval:       300, // 默认5分钟
                    Group:          "默认分组",
                    AllowPartMatch: true,
                    Enabled:        true, // 默认启用
                }
                
                // 尝试加载对应的关键词（注意：旧格式使用0开始的索引）
                if keywords := os.Getenv(fmt.Sprintf("RSS_KEYWORDS_%d", i)); keywords != "" {
                    config.RSS[i].Keywords = strings.Split(keywords, ",")
                }
                
                // 尝试加载对应的间隔时间
                if interval := os.Getenv(fmt.Sprintf("RSS_INTERVAL_%d", i)); interval != "" {
                    if parsedInterval, err := strconv.Atoi(interval); err == nil && parsedInterval > 0 {
                        config.RSS[i].Interval = parsedInterval
                    }
                }
                
                // 尝试加载对应的分组
                if group := os.Getenv(fmt.Sprintf("RSS_GROUP_%d", i)); group != "" {
                    config.RSS[i].Group = group
                }
            }
        }
    }

    return config
}

func (c *Config) Save(filename string) error {
    // 确保目录存在
    dir := filepath.Dir(filename)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return fmt.Errorf("创建配置目录失败: %v", err)
    }
    
    data, err := yaml.Marshal(c)
    if err != nil {
        return fmt.Errorf("序列化配置失败: %v", err)
    }
    return ioutil.WriteFile(filename, data, 0644)
}
