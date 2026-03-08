package alert

import (
	"ai-eino-interview-agent/internal/config"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// FeishuMessage 飞书消息结构
type FeishuMessage struct {
	MsgType string      `json:"msg_type"`
	Content interface{} `json:"content"`
}

// FeishuTextContent 飞书文本消息内容
type FeishuTextContent struct {
	Text string `json:"text"`
}

// SendFeishuAlert 发送飞书告警
func SendFeishuAlert(title, content string) error {
	if !config.Global.Feishu.Enabled {
		log.Printf("[FeishuAlert] 飞书告警未启用，跳过发送")
		return nil
	}

	webhookURL := config.Global.Feishu.WebhookURL
	if webhookURL == "" {
		log.Printf("[FeishuAlert] 飞书 Webhook URL 未配置")
		return fmt.Errorf("飞书 Webhook URL 未配置")
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	text := fmt.Sprintf("【%s】\n时间: %s\n%s", title, timestamp, content)

	message := FeishuMessage{
		MsgType: "text",
		Content: FeishuTextContent{
			Text: text,
		},
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		log.Printf("[FeishuAlert] 序列化消息失败: %v", err)
		return err
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("[FeishuAlert] 发送请求失败: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[FeishuAlert] 飞书返回非200状态码: %d", resp.StatusCode)
		return fmt.Errorf("飞书返回状态码: %d", resp.StatusCode)
	}

	log.Printf("[FeishuAlert] 告警发送成功: %s", title)
	return nil
}

// SendDatabaseErrorAlert 发送数据库错误告警
func SendDatabaseErrorAlert(operation string, err error, retryCount int) {
	title := "数据库操作失败告警"
	content := fmt.Sprintf("操作: %s\n错误: %v\n已尝试次数: %d", operation, err, retryCount)

	if alertErr := SendFeishuAlert(title, content); alertErr != nil {
		log.Printf("[FeishuAlert] 发送告警失败: %v", alertErr)
	}

}
