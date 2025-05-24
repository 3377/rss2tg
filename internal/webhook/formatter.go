package webhook

import (
	"fmt"
	"strings"
	"time"
)

// Formatter 消息格式转换器
type Formatter struct{}

// NewFormatter 创建新的格式转换器
func NewFormatter() *Formatter {
	return &Formatter{}
}

// FormatMessage 将 rss2tg 消息转换为 webhook 格式
func (f *Formatter) FormatMessage(title, url, group string, pubDate time.Time, matchedKeywords []string) Message {
	// 格式化时间为中国时区
	chinaLoc, _ := time.LoadLocation("Asia/Shanghai")
	pubDateChina := pubDate.In(chinaLoc)
	timestamp := pubDateChina.Format("2006-01-02 15:04:05")

	// 处理关键词
	keywords := ""
	if len(matchedKeywords) > 0 {
		keywords = strings.Join(matchedKeywords, ", ")
	}

	// 生成简短描述
	description := fmt.Sprintf("分组: %s", group)
	if keywords != "" {
		description += fmt.Sprintf(" | 关键词: %s", keywords)
	}
	description += fmt.Sprintf(" | 时间: %s", timestamp)

	// 生成完整内容（Markdown格式）
	content := fmt.Sprintf("📰 **%s**\n\n", title)
	content += fmt.Sprintf("🌐 **链接:** %s\n\n", url)
	
	if keywords != "" {
		// 为关键词添加标签格式
		keywordTags := make([]string, len(matchedKeywords))
		for i, keyword := range matchedKeywords {
			keywordTags[i] = fmt.Sprintf("#%s", keyword)
		}
		content += fmt.Sprintf("🔍 **关键词:** %s\n\n", strings.Join(keywordTags, " "))
	}
	
	content += fmt.Sprintf("🏷️ **分组:** %s\n\n", group)
	content += fmt.Sprintf("🕒 **时间:** %s", timestamp)

	return Message{
		Title:       title,
		Description: description,
		Content:     content,
		URL:         url,
		Group:       group,
		Keywords:    keywords,
		Timestamp:   timestamp,
	}
} 