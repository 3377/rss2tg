package enhancer

import (
	"log"
	"time"

	"rss2tg/internal/webhook"
)

// MessageHandler 原始消息处理器类型
type MessageHandler func(title, url, group string, pubDate time.Time, matchedKeywords []string) error

// EnhancedMessageHandler 增强的消息处理器
type EnhancedMessageHandler struct {
	originalHandler func(title, url, group string, pubDate time.Time, matchedKeywords []string) error
	webhookClient   *webhook.Client
	multiWebhookClient *webhook.MultiClient
	formatter       *webhook.Formatter
}

// NewEnhancedMessageHandler 创建增强的消息处理器（单个 webhook）
func NewEnhancedMessageHandler(originalHandler func(title, url, group string, pubDate time.Time, matchedKeywords []string) error, webhookClient *webhook.Client) *EnhancedMessageHandler {
	return &EnhancedMessageHandler{
		originalHandler: originalHandler,
		webhookClient:   webhookClient,
		formatter:       webhook.NewFormatter(),
	}
}

// NewEnhancedMultiMessageHandler 创建增强的消息处理器（多个 webhook）
func NewEnhancedMultiMessageHandler(originalHandler func(title, url, group string, pubDate time.Time, matchedKeywords []string) error, multiWebhookClient *webhook.MultiClient) *EnhancedMessageHandler {
	return &EnhancedMessageHandler{
		originalHandler: originalHandler,
		multiWebhookClient: multiWebhookClient,
		formatter:       webhook.NewFormatter(),
	}
}

// HandleMessage 处理消息，同时发送到 Telegram 和 webhook
func (h *EnhancedMessageHandler) HandleMessage(title, url, group string, pubDate time.Time, matchedKeywords []string) error {
	// 首先发送到原有的 Telegram 推送
	err := h.originalHandler(title, url, group, pubDate, matchedKeywords)
	if err != nil {
		log.Printf("Telegram 推送失败: %v", err)
		// 注意：即使 Telegram 推送失败，我们仍然继续 webhook 推送
	}

	// 异步发送到 webhook（不影响 Telegram 推送）
	go func() {
		msg := h.formatter.FormatMessage(title, url, group, pubDate, matchedKeywords)
		
		if h.multiWebhookClient != nil {
			// 使用多 webhook 客户端
			results := h.multiWebhookClient.Send(msg)
			for _, result := range results {
				if result.Success {
					log.Printf("Webhook [%s] 推送成功", result.Name)
				} else {
					log.Printf("Webhook [%s] 推送失败: %v", result.Name, result.Error)
				}
			}
		} else if h.webhookClient != nil {
			// 使用单个 webhook 客户端（向后兼容）
			if err := h.webhookClient.Send(msg); err != nil {
				log.Printf("Webhook 推送失败: %v", err)
			}
		}
	}()

	return err // 返回 Telegram 推送的结果
} 