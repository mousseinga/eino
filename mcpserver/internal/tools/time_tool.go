package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

// TimeTool 获取当前时间工具
type TimeTool struct{}

func NewTimeTool() *TimeTool {
	return &TimeTool{}
}

func (t *TimeTool) Name() string {
	return "get_current_time"
}

func (t *TimeTool) Description() string {
	return "Get the current time in various formats."
}

func (t *TimeTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"format": map[string]interface{}{
				"type":        "string",
				"description": "Time format: 'unix' for Unix timestamp, 'rfc3339' for RFC3339 format, 'human' for human-readable format",
				"enum":        []string{"unix", "rfc3339", "human"},
				"default":     "rfc3339",
			},
		},
	}
}

func (t *TimeTool) Execute(ctx context.Context, arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	now := time.Now()
	format := "rfc3339"

	if f, ok := arguments["format"].(string); ok && f != "" {
		format = f
	}

	var timeStr string
	switch format {
	case "unix":
		timeStr = fmt.Sprintf("%d", now.Unix())
	case "rfc3339":
		timeStr = now.Format(time.RFC3339)
	case "human":
		timeStr = now.Format("2006-01-02 15:04:05")
	default:
		timeStr = now.Format(time.RFC3339)
	}

	return mcp.NewToolResultText(timeStr), nil
}
