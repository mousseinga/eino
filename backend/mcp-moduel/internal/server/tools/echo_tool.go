package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// EchoTool 回显工具示例
type EchoTool struct{}

func NewEchoTool() *EchoTool {
	return &EchoTool{}
}

func (t *EchoTool) Name() string {
	return "echo"
}

func (t *EchoTool) Description() string {
	return "Echo back the input message. Useful for testing and debugging."
}

func (t *EchoTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"message": map[string]interface{}{
				"type":        "string",
				"description": "The message to echo back",
			},
		},
		"required": []string{"message"},
	}
}

func (t *EchoTool) Execute(ctx context.Context, arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	message, ok := arguments["message"].(string)
	if !ok {
		return mcp.NewToolResultError("Invalid argument: message must be a string"), nil
	}

	result := fmt.Sprintf("Echo: %s", message)
	return mcp.NewToolResultText(result), nil
}
