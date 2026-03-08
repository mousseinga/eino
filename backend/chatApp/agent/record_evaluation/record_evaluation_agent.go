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

func NewRecordEvaluationAgent(userId uint) (adk.Agent, error) {
	ctx := context.Background()
	model, err := chat.CreatOpenAiChatModel(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenAI chat model: %w", err)
	}
	baseAgent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "RecordEvaluationAgent",
		Description: "一个专业评估面试记录并生成专业报告的智能体",
		Model:       model,
		Instruction: RecordEvaluationInstruction,
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
