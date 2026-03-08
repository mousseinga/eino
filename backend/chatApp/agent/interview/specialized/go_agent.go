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

// NewGoSpecializedAgent 创建 Go 专项面试官智能体
// 专注于评估候选人在 Go 语言方面的专业技能和深度
func NewGoSpecializedAgent(userId uint, needResumeTool bool) (adk.Agent, error) {
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
		Name:          "GoSpecializedAgent",
		Description:   "Go 专项面试官智能体，专注于评估候选人在 Go 语言方面的专业技能和深度",
		Instruction:   GoSpecializedAgentInstruction,
		Model:         model,
		ToolsConfig:   toolsConfig,
		MaxIterations: 15,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Go specialized agent: %w", err)
	}
	return baseAgent, nil
}
