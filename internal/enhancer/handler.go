package enhancer

import (
	"log"
	"time"

	"rss2tg/internal/webhook"
)

// MessageHandler 原始消息处理器类型
type MessageHandler func(title, url, group string, pubDate time.Time, matchedKeywords []string) error

// EnhancedMessageHandler 增强消息处理器
type EnhancedMessageHandler struct {
	originalHandler MessageHandler
	webhookClient   *webhook.Client
	formatter       *webhook.Formatter
}

// NewEnhancedMessageHandler 创建新的增强消息处理器
func NewEnhancedMessageHandler(originalHandler MessageHandler, webhookClient *webhook.Client) *EnhancedMessageHandler {
	return &EnhancedMessageHandler{
		originalHandler: originalHandler,
		webhookClient:   webhookClient,
		formatter:       webhook.NewFormatter(),
	}
}

// HandleMessage 处理消息，先执行原有的 Telegram 推送，再异步执行 webhook 推送
func (h *EnhancedMessageHandler) HandleMessage(title, url, group string, pubDate time.Time, matchedKeywords []string) error {
	// 1. 首先执行原有的 Telegram 推送（保证原功能不变）
	if err := h.originalHandler(title, url, group, pubDate, matchedKeywords); err != nil {
		return err // 如果 Telegram 推送失败，直接返回错误
	}

	// 2. 如果启用了 webhook，则异步执行 webhook 推送
	if h.webhookClient != nil && h.webhookClient.Enabled {
		go h.sendToWebhook(title, url, group, pubDate, matchedKeywords)
	}

	return nil
}

// sendToWebhook 异步发送到 webhook
func (h *EnhancedMessageHandler) sendToWebhook(title, url, group string, pubDate time.Time, matchedKeywords []string) {
	// 格式化消息
	webhookMsg := h.formatter.FormatMessage(title, url, group, pubDate, matchedKeywords)

	// 发送到 webhook
	if err := h.webhookClient.Send(webhookMsg); err != nil {
		log.Printf("Webhook 推送失败: %v", err)
	}
} 