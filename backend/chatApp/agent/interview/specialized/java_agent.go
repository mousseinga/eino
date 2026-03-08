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

// NewJavaSpecializedAgent 创建 Java 专项面试官智能体
// 专注于评估候选人在 Java 方面的专业技能和深度
func NewJavaSpecializedAgent(userId uint, needResumeTool bool) (adk.Agent, error) {
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
		Name:          "JavaSpecializedAgent",
		Description:   "Java 专项面试官智能体，专注于评估候选人在 Java 方面的专业技能和深度",
		Instruction:   JavaSpecializedAgentInstruction,
		Model:         model,
		ToolsConfig:   toolsConfig,
		MaxIterations: 15,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Java specialized agent: %w", err)
	}
	return baseAgent, nil
}
