package record_evaluation

import (
	"ai-eino-interview-agent/chatApp/chat"
	tool2 "ai-eino-interview-agent/chatApp/tool"
	"fmt"

	"github.com/cloudwego/eino/adk"
	componenttool "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"golang.org/x/net/context"
)

func NewAnswerRecordAgent(userId uint) (adk.Agent, error) {
	ctx := context.Background()

	model, err := chat.CreatOpenAiChatModel(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenAI chat model: %w", err)
	}

	baseAgent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "AnswerRecordAgent",
		Description: "一个专业用于评价答题记录中每个parent问题和子问题的记录的智能体",
		Instruction: AnswerRecordAgentInstruction,
		Model:       model,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: []componenttool.BaseTool{
					tool2.GetMianshiInfoTool(),
				},
			},
		},
		MaxIterations: 15,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create evaluation agent: %w", err)
	}
	return baseAgent, nil
}
