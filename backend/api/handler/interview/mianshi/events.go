package mianshi

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// SendSSEEvent 发送 SSE 事件
func SendSSEEvent(writer io.Writer, event map[string]interface{}) error {
	eventJSON, _ := json.Marshal(event)

	// 获取事件类型
	eventType := "message"
	if t, ok := event["type"]; ok {
		eventType = fmt.Sprintf("%v", t)
	}

	// 标准 SSE 格式：event: type\ndata: {...}\n\n
	message := fmt.Sprintf("event: %s\ndata: %s\n\n", eventType, string(eventJSON))
	n, err := fmt.Fprint(writer, message)
	if err != nil {
		log.Printf("[SSE] Failed to write event: %v (wrote %d bytes)", err, n)
		return err
	}

	// 立即 flush，确保数据发送到客户端
	if flusher, ok := writer.(http.Flusher); ok {
		flusher.Flush()
	}
	return nil
}

// SendErrorEvent 发送错误事件
func SendErrorEvent(writer io.Writer, message string) {
	err := SendSSEEvent(writer, map[string]interface{}{"type": "error", "message": message})
	if err != nil {
		return
	}
}

// SendCompleteEvent 发送完成事件
func SendCompleteEvent(writer io.Writer) {
	err := SendSSEEvent(writer, map[string]interface{}{"type": "complete", "message": "面试已结束"})
	if err != nil {
		return
	}
}

// SendReadyEventWithSession 发送就绪事件
func SendReadyEventWithSession(writer io.Writer, questionIndex int, sessionID string) {
	err := SendSSEEvent(writer, map[string]interface{}{
		"type":           "ready_for_answer",
		"message":        "请回答上述问题",
		"question_index": questionIndex,
		"session_id":     sessionID,
	})
	if err != nil {
		return
	}
}

// SendHeartbeatEvent 发送心跳事件保活连接
func SendHeartbeatEvent(writer io.Writer) error {
	return SendSSEEvent(writer, map[string]interface{}{
		"type":    "heartbeat",
		"message": "连接保活",
	})
}
