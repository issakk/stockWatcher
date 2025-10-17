package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Notifier 通知器接口
type Notifier interface {
	Send(message string) error
}

// WeChatNotifier 企业微信通知器
type WeChatNotifier struct {
	webhookURL string
	client     *http.Client
}

// WeChatMessage 企业微信消息格式
type WeChatMessage struct {
	MsgType  string      `json:"msgtype"`
	Text     TextContent `json:"text,omitempty"`
	Markdown Markdown    `json:"markdown,omitempty"`
}

// TextContent 文本内容
type TextContent struct {
	Content string `json:"content"`
}

// Markdown Markdown内容
type Markdown struct {
	Content string `json:"content"`
}

// NewWeChatNotifier 创建企业微信通知器
func NewWeChatNotifier(webhookURL string) (*WeChatNotifier, error) {
	if webhookURL == "" {
		return nil, fmt.Errorf("企业微信webhookURL不能为空")
	}

	return &WeChatNotifier{
		webhookURL: webhookURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

// Send 发送消息
func (w *WeChatNotifier) Send(message string) error {
	// 使用纯文本格式发送消息，避免编码问题
	msg := WeChatMessage{
		MsgType: "text",
		Text: TextContent{
			Content: message,
		},
	}

	return w.sendMessage(msg)
}

// sendMessage 发送HTTP请求
func (w *WeChatNotifier) sendMessage(msg WeChatMessage) error {
	jsonData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %v", err)
	}

	// 创建HTTP请求，明确设置UTF-8编码
	req, err := http.NewRequest("POST", w.webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头，明确指定UTF-8编码
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json")

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("企业微信API返回错误状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 读取响应以确保消息发送成功
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %v", err)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("解析响应失败: %v", err)
	}

	// 检查企业微信API返回状态
	if errCode, ok := response["errcode"].(float64); ok && errCode != 0 {
		errMsg := "未知错误"
		if msg, ok := response["errmsg"].(string); ok {
			errMsg = msg
		}
		return fmt.Errorf("企业微信API返回错误: %d - %s", int(errCode), errMsg)
	}

	return nil
}

// TestConnection 测试企业微信连接
func (w *WeChatNotifier) TestConnection() error {
	testMsg := WeChatMessage{
		MsgType: "text",
		Text: TextContent{
			Content: "🚀 股票监视器已启动，连接测试成功！✅",
		},
	}

	return w.sendMessage(testMsg)
}
