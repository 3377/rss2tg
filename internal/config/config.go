package config

import (
    "io/ioutil"
    "log"
    "os"
    "reflect"
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
    Telegram TelegramConfig `yaml:"telegram"`
    RSS      []RSSConfig    `yaml:"rss"`
}

func Load(filename string) (*Config, error) {
    data, err := ioutil.ReadFile(filename)
    if err != nil {
        return nil, err
    }

    var config Config
    err = yaml.Unmarshal(data, &config)
    if err != nil {
        return nil, err
    }

    return &config, nil
}

func LoadFromEnv() *Config {
    log.Println("从环境变量加载配置")
    return &Config{
        Telegram: TelegramConfig{
            BotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
            Users:    strings.Split(os.Getenv("TELEGRAM_USERS"), ","),
            Channels: strings.Split(os.Getenv("TELEGRAM_CHANNELS"), ","),
        },
        RSS: []RSSConfig{}, // RSS 配置仍然从配置文件加载
    }
}

func (c *Config) Save(filename string) error {
    data, err := yaml.Marshal(c)
    if err != nil {
        return err
    }

    return ioutil.WriteFile(filename, data, 0644)
}

func (c *Config) Equal(other *Config) bool {
    return reflect.DeepEqual(c, other)
}
