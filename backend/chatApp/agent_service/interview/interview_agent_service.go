package interview

import (
	"ai-eino-interview-agent/chatApp/agent/interview/comprehensive"
	"ai-eino-interview-agent/chatApp/agent/interview/specialized"
	"context"
	"fmt"
	"log"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
)

// InterviewAgentType 面试智能体类型
type InterviewAgentType string

const (
	// Comprehensive 综合面试类型
	ComprehensiveSchool InterviewAgentType = "comprehensive_school" // 综合校招
	ComprehensiveSocial InterviewAgentType = "comprehensive_social" // 综合社招

	// Specialized 专项面试类型
	SpecializedGo    InterviewAgentType = "specialized_go"    // Go 专项
	SpecializedJava  InterviewAgentType = "specialized_java"  // Java 专项
	SpecializedMQ    InterviewAgentType = "specialized_mq"    // MQ 专项
	SpecializedMySQL InterviewAgentType = "specialized_mysql" // MySQL 专项
	SpecializedRedis InterviewAgentType = "specialized_redis" // Redis 专项
)

// InterviewAgentService 面试智能体服务
type InterviewAgentService struct {
	userId uint
}

// NewInterviewAgentService 创建面试智能体服务
func NewInterviewAgentService(userId uint) *InterviewAgentService {
	return &InterviewAgentService{
		userId: userId,
	}
}

// GetInterviewAgent 根据类型获取对应的面试智能体
// 参数:
//   - agentType: 智能体类型
//   - needResumeTool: 是否需要简历工具
//
// 返回:
//   - adk.Agent: 面试智能体
//   - error: 错误信息
func (s *InterviewAgentService) GetInterviewAgent(agentType InterviewAgentType, needResumeTool bool) (adk.Agent, error) {
	switch agentType {
	// Comprehensive 综合面试智能体
	case ComprehensiveSchool:
		return comprehensive.NewSchoolComprehensiveAgent(s.userId, needResumeTool)
	case ComprehensiveSocial:
		return comprehensive.NewSocialComprehensiveAgent(s.userId, needResumeTool)

	// Specialized 专项面试智能体
	case SpecializedGo:
		return specialized.NewGoSpecializedAgent(s.userId, needResumeTool)
	case SpecializedJava:
		return specialized.NewJavaSpecializedAgent(s.userId, needResumeTool)
	case SpecializedMQ:
		return specialized.NewMQSpecializedAgent(s.userId, needResumeTool)
	case SpecializedMySQL:
		return specialized.NewMySQLSpecializedAgent(s.userId, needResumeTool)
	case SpecializedRedis:
		return specialized.NewRedisSpecializedAgent(s.userId, needResumeTool)

	default:
		return nil, fmt.Errorf("unknown interview agent type: %s", agentType)
	}
}

// RunInterviewWithCallback 运行面试并通过回调返回结果
// 参数:
//   - ctx: 上下文
//   - agentType: 智能体类型
//   - needResumeTool: 是否需要简历工具
//   - prompt: 提示词/问题
//   - callback: 回调函数，每次接收到消息时调用
//
// 返回:
//   - error: 错误信息
func (s *InterviewAgentService) RunInterviewWithCallback(ctx context.Context, agentType InterviewAgentType, needResumeTool bool, prompt string, callback func(message string) error) error {
	// 获取对应的智能体
	agent, err := s.GetInterviewAgent(agentType, needResumeTool)
	if err != nil {
		log.Printf("[RunInterviewWithCallback] 获取智能体失败: %v", err)
		return err
	}

	// 运行智能体
	_, err = runAgentWithIterator(ctx, agent, prompt, callback)
	if err != nil {
		log.Printf("[RunInterviewWithCallback] 智能体执行出错: %v", err)
		return err
	}

	return nil
}

// runAgentWithIterator 运行智能体的通用方法
// 参数:
//   - ctx: 上下文
//   - agent: 智能体
//   - prompt: 提示词/问题
//   - callback: 可选的回调函数，为 nil 时只收集最后一条消息
//
// 返回:
//   - string: 最后一条消息（仅当 callback 为 nil 时有效）
//   - error: 错误信息
func runAgentWithIterator(ctx context.Context, agent adk.Agent, prompt string, callback func(string) error) (string, error) {
	// 创建 Runner
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent: agent,
	})

	// 构建消息
	messages := []adk.Message{
		schema.UserMessage(prompt),
	}

	// 运行智能体
	iter := runner.Run(ctx, messages)

	var lastMessage string
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}

		if event.Err != nil {
			return "", event.Err
		}

		// 处理消息事件
		if event.Output != nil && event.Output.MessageOutput != nil {
			message := event.Output.MessageOutput.Message.Content
			if message != "" {
				lastMessage = message
				// 如果有回调函数，调用它
				if callback != nil {
					if err := callback(message); err != nil {
						return "", err
					}
				}
			}
		}
	}

	return lastMessage, nil
}
