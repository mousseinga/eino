package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/redis/go-redis/v9"
)

// RedisQueue Redis 消息队列实现
type RedisQueue struct {
	client   *redis.Client
	mu       sync.RWMutex
	handlers []MessageHandler
	done     chan struct{}
}

// NewRedisQueue 创建 Redis 消息队列
func NewRedisQueue(client *redis.Client) *RedisQueue {
	if client == nil {
		log.Fatal("Redis client cannot be nil")
	}
	return &RedisQueue{
		client: client,
		done:   make(chan struct{}),
	}
}

// Publish 发布消息到 Redis
func (q *RedisQueue) Publish(ctx context.Context, message *Message) error {
	// 序列化消息
	messageJSON, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// 根据消息类型发布到不同的 channel
	channel := fmt.Sprintf("interview:messages:%s", message.Type)

	log.Printf("[RedisQueue] Publishing message to channel %s: %s", channel, string(messageJSON))

	// 发布消息
	result := q.client.Publish(ctx, channel, string(messageJSON))
	if err := result.Err(); err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	// 获取订阅者数量
	numSubscribers := result.Val()
	log.Printf("[RedisQueue] Message published to %d subscribers", numSubscribers)

	return nil
}

// Subscribe 订阅消息
func (q *RedisQueue) Subscribe(ctx context.Context, handler MessageHandler) error {
	q.mu.Lock()
	q.handlers = append(q.handlers, handler)
	q.mu.Unlock()

	// 订阅所有消息类型的 channel
	channels := []string{
		fmt.Sprintf("interview:messages:%s", MessageTypeEvaluationReport),
		fmt.Sprintf("interview:messages:%s", MessageTypeTopicEvaluation),
	}

	log.Printf("[RedisQueue] Subscribing to channels: %v", channels)

	pubsub := q.client.Subscribe(ctx, channels...)
	defer pubsub.Close()

	log.Printf("[RedisQueue] Successfully subscribed to channels, waiting for messages...")

	// 处理消息
	ch := pubsub.Channel()
	messageCount := 0
	for {
		select {
		case <-ctx.Done():
			log.Printf("[RedisQueue] Context cancelled, stopping subscription (received %d messages)", messageCount)
			return ctx.Err()
		case <-q.done:
			log.Printf("[RedisQueue] Queue closed, stopping subscription (received %d messages)", messageCount)
			return nil
		case msg := <-ch:
			if msg == nil {
				log.Printf("[RedisQueue] Received nil message")
				continue
			}

			messageCount++
			log.Printf("[RedisQueue] Received message #%d from channel %s (payload length: %d)",
				messageCount, msg.Channel, len(msg.Payload))

			// 反序列化消息
			var message Message
			if err := json.Unmarshal([]byte(msg.Payload), &message); err != nil {
				log.Printf("[RedisQueue] Failed to unmarshal message: %v", err)
				continue
			}

			// 异步调用所有处理器
			q.mu.RLock()
			handlers := q.handlers
			q.mu.RUnlock()

			for _, h := range handlers {
				go func(handler MessageHandler, m *Message) {
					if err := handler(ctx, m); err != nil {
						log.Printf("[RedisQueue] Error processing message: %v, type: %s", err, m.Type)
					}
				}(h, &message)
			}
		}
	}
}

// Close 关闭消息队列
func (q *RedisQueue) Close() error {
	close(q.done)
	// 注意：不关闭 Redis 客户端，因为它由外部管理
	return nil
}

// GetRedisClient 获取 Redis 客户端（用于测试和其他用途）
func (q *RedisQueue) GetRedisClient() *redis.Client {
	return q.client
}
