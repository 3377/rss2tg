package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// Client webhook 客户端
type Client struct {
	URL        string
	Timeout    time.Duration
	RetryCount int
	Enabled    bool
}

// Message webhook 消息结构
type Message struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Content     string `json:"content"`
	URL         string `json:"url"`
	Group       string `json:"group"`
	Keywords    string `json:"keywords"`
	Timestamp   string `json:"timestamp"`
}

// Response webhook 响应结构
type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// NewClient 创建新的 webhook 客户端
func NewClient(url string, timeout int, retryCount int, enabled bool) *Client {
	return &Client{
		URL:        url,
		Timeout:    time.Duration(timeout) * time.Second,
		RetryCount: retryCount,
		Enabled:    enabled,
	}
}

// Send 发送消息到 webhook
func (c *Client) Send(msg Message) error {
	if !c.Enabled {
		return nil
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %v", err)
	}

	client := &http.Client{
		Timeout: c.Timeout,
	}

	var lastErr error
	for i := 0; i <= c.RetryCount; i++ {
		if i > 0 {
			log.Printf("Webhook 推送重试第 %d 次", i)
			time.Sleep(time.Duration(i) * time.Second) // 指数退避
		}

		resp, err := client.Post(c.URL, "application/json", bytes.NewBuffer(data))
		if err != nil {
			lastErr = fmt.Errorf("发送请求失败: %v", err)
			log.Printf("Webhook 推送失败: %v", lastErr)
			continue
		}

		defer resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			log.Printf("Webhook 推送成功: %s", c.URL)
			return nil
		}

		var webhookResp Response
		if err := json.NewDecoder(resp.Body).Decode(&webhookResp); err == nil {
			lastErr = fmt.Errorf("webhook 返回错误: %s", webhookResp.Message)
		} else {
			lastErr = fmt.Errorf("webhook 返回状态码: %d", resp.StatusCode)
		}
		log.Printf("Webhook 推送失败: %v", lastErr)
	}

	return fmt.Errorf("webhook 推送最终失败: %v", lastErr)
} 