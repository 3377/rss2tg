package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// Client webhook 客户端
type Client struct {
	URL        string
	Timeout    time.Duration
	RetryCount int
	Enabled    bool
}

// MultiClient 多 webhook 客户端
type MultiClient struct {
	Clients []WebhookClient
}

// WebhookClient 单个 webhook 客户端配置
type WebhookClient struct {
	Name       string
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

// SendResult 发送结果
type SendResult struct {
	Name    string
	Success bool
	Error   error
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

// NewMultiClient 创建新的多 webhook 客户端
func NewMultiClient(clients []WebhookClient) *MultiClient {
	return &MultiClient{
		Clients: clients,
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

// Send 并发发送消息到多个 webhook
func (mc *MultiClient) Send(msg Message) []SendResult {
	if len(mc.Clients) == 0 {
		return []SendResult{}
	}

	results := make([]SendResult, len(mc.Clients))
	var wg sync.WaitGroup

	for i, client := range mc.Clients {
		if !client.Enabled {
			results[i] = SendResult{
				Name:    client.Name,
				Success: false,
				Error:   fmt.Errorf("webhook 未启用"),
			}
			continue
		}

		wg.Add(1)
		go func(index int, webhookClient WebhookClient) {
			defer wg.Done()

			// 创建单个客户端
			singleClient := &Client{
				URL:        webhookClient.URL,
				Timeout:    webhookClient.Timeout,
				RetryCount: webhookClient.RetryCount,
				Enabled:    webhookClient.Enabled,
			}

			err := singleClient.Send(msg)
			results[index] = SendResult{
				Name:    webhookClient.Name,
				Success: err == nil,
				Error:   err,
			}
		}(i, client)
	}

	wg.Wait()
	return results
} 