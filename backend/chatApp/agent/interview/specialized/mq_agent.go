package specialized

import (
	"ai-eino-interview-agent/chatApp/chat"
	tool2 "ai-eino-interview-agent/chatApp/tool"
	"context"
	"fmt"

	"github.com/cloudwego/eino/adk"
	componenttool "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
)

// NewMQSpecializedAgent 创建 MQ 专项面试官智能体
// 专注于评估候选人在消息队列技术方面的专业能力和深度
func NewMQSpecializedAgent(userId uint, needResumeTool bool) (adk.Agent, error) {
	ctx := context.Background()

	var toolsConfig adk.ToolsConfig
	if needResumeTool {
		toolsConfig = adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: []componenttool.BaseTool{
					tool2.GetResumeInfoTool(),
				},
			},
		}
	}

	model, err := chat.CreatOpenAiChatModel(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenAI chat model: %w", err)
	}

	baseAgent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:          "MQSpecializedAgent",
		Description:   "MQ 专项面试官智能体，专注于评估候选人在消息队列技术方面的专业能力和深度",
		Instruction:   MQSpecializedAgentInstruction,
		Model:         model,
		ToolsConfig:   toolsConfig,
		MaxIterations: 15,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MQ specialized agent: %w", err)
	}
	return baseAgent, nil
}
