package mq

import (
	"ai-eino-interview-agent/chatApp/agent_service/evaluation"
	"context"
	"fmt"
	"log"
)

// ConsumerHandler 消费者处理器
type ConsumerHandler struct {
}

// NewConsumerHandler 创建消费者处理器
func NewConsumerHandler() *ConsumerHandler {
	return &ConsumerHandler{}
}

// HandleMessage 处理消息
func (h *ConsumerHandler) HandleMessage(ctx context.Context, message *Message) error {
	switch message.Type {
	case MessageTypeEvaluationReport:
		return h.handleEvaluationReport(ctx, message)
	case MessageTypeTopicEvaluation:
		return h.handleTopicEvaluation(ctx, message)
	default:
		return fmt.Errorf("unknown message type: %s", message.Type)
	}
}

// handleEvaluationReport 处理评估报告消息
func (h *ConsumerHandler) handleEvaluationReport(ctx context.Context, message *Message) error {
	log.Printf("[Consumer] Processing evaluation report message")

	// 提取负载
	userID, ok := message.Payload["user_id"].(float64)
	if !ok {
		return fmt.Errorf("invalid user_id in payload")
	}

	reportID, ok := message.Payload["report_id"].(float64)
	if !ok {
		return fmt.Errorf("invalid report_id in payload")
	}

	log.Printf("[Consumer] Generating evaluation report: userID=%d, reportID=%d", uint(userID), uint64(reportID))

	// 调用评估服务生成报告
	// 这里使用 evaluation.GenerateRecordEvaluation 生成整体评估
	_, err := evaluation.GenerateRecordEvaluation(ctx, uint(userID), uint64(reportID))
	if err != nil {
		log.Printf("[Consumer] Failed to generate evaluation report: %v", err)
		return err
	}

	log.Printf("[Consumer] Evaluation report generated successfully: userID=%d, reportID=%d", uint(userID), uint64(reportID))
	return nil
}

// handleTopicEvaluation 处理主题评估消息
func (h *ConsumerHandler) handleTopicEvaluation(ctx context.Context, message *Message) error {
	log.Printf("[Consumer] Processing topic evaluation message")

	// 提取负载
	userID, ok := message.Payload["user_id"].(float64)
	if !ok {
		return fmt.Errorf("invalid user_id in payload")
	}

	reportID, ok := message.Payload["report_id"].(float64)
	if !ok {
		return fmt.Errorf("invalid report_id in payload")
	}

	log.Printf("[Consumer] Generating topic evaluation: userID=%d, reportID=%d", uint(userID), uint64(reportID))

	// 调用评估服务生成主题评估
	// 这里使用 GenerateAnswerRecordEvaluation 生成答题记录的评估
	_, err := evaluation.GenerateAnswerRecordEvaluation(ctx, uint(userID), uint64(reportID))
	if err != nil {
		log.Printf("[Consumer] Failed to generate topic evaluation: %v", err)
		return err
	}

	log.Printf("[Consumer] Topic evaluation generated successfully: userID=%d, reportID=%d", uint(userID), uint64(reportID))
	return nil
}

// StartConsumer 启动消费者
func StartConsumer(ctx context.Context) error {
	mq := GetMessageQueue()
	handler := NewConsumerHandler()

	log.Printf("[Consumer] Starting message consumer")
	log.Printf("[Consumer] MQ type: %T", mq)

	err := mq.Subscribe(ctx, handler.HandleMessage)
	log.Printf("[Consumer] Consumer stopped: %v", err)
	return err
}
