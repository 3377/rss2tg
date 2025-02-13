package config

import (
    "fmt"
    "io/ioutil"
    "log"
    "net/url"
    "os"
    "reflect"
    "strconv"
    "strings"

    "gopkg.in/yaml.v2"
)

type TelegramConfig struct {
    BotToken string   `yaml:"bot_token"`
    Users    []string `yaml:"users"`
    Channels []string `yaml:"channels"`
}

type RSSConfig struct {
    URL      string   `yaml:"url"`
    Interval int      `yaml:"interval"`
    Keywords []string `yaml:"keywords"`
    Group    string   `yaml:"group"`
}

type Config struct {
    Telegram struct {
        BotToken  string   `yaml:"bot_token"`
        Users     []string `yaml:"users"`
        Channels  []string `yaml:"channels"`
    } `yaml:"telegram"`
    RSS []struct {
        URLs     []string `yaml:"urls"`
        Interval int      `yaml:"interval"`
        Keywords []string `yaml:"keywords"`
        Group    string   `yaml:"group"`
    } `yaml:"rss"`
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
    
    data, err := ioutil.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("读取配置文件失败: %v", err)
    }

    var config Config
    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, fmt.Errorf("解析配置文件失败: %v", err)
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
    for i, rss := range config.RSS {
        // 验证URLs
        if len(rss.URLs) == 0 {
            return fmt.Errorf("RSS #%d: URLs为空", i+1)
        }
        
        // 验证每个URL的格式
        for j, urlStr := range rss.URLs {
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
        if rss.Interval <= 0 {
            config.RSS[i].Interval = 300 // 默认5分钟
        }

        // 设置默认分组
        if rss.Group == "" {
            config.RSS[i].Group = "默认分组"
        }

        // 清理关键词列表
        cleanKeywords := make([]string, 0)
        for _, keyword := range rss.Keywords {
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
        config.RSS = make([]struct {
            URLs     []string `yaml:"urls"`
            Interval int      `yaml:"interval"`
            Keywords []string `yaml:"keywords"`
            Group    string   `yaml:"group"`
        }, len(urlGroups))

        for i, urlGroup := range urlGroups {
            config.RSS[i].URLs = strings.Split(strings.TrimSpace(urlGroup), ",")
            config.RSS[i].Interval = 300 // 默认5分钟
            config.RSS[i].Group = "默认分组"
            
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
        return err
    }

    return ioutil.WriteFile(filename, data, 0644)
}
