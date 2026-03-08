package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// TextProcessorTool 文本处理工具
type TextProcessorTool struct{}

func NewTextProcessorTool() *TextProcessorTool {
	return &TextProcessorTool{}
}

func (t *TextProcessorTool) Name() string {
	return "text_processor"
}

func (t *TextProcessorTool) Description() string {
	return "Process text with various operations: reverse, uppercase, lowercase, count characters, remove spaces, etc."
}

func (t *TextProcessorTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"text": map[string]interface{}{
				"type":        "string",
				"description": "The text to process",
			},
			"operation": map[string]interface{}{
				"type":        "string",
				"description": "The operation to perform on the text",
				"enum": []string{
					"reverse",
					"uppercase",
					"lowercase",
					"count",
					"remove_spaces",
					"trim",
					"word_count",
				},
			},
		},
		"required": []string{"text", "operation"},
	}
}

func (t *TextProcessorTool) Execute(ctx context.Context, arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	text, ok := arguments["text"].(string)
	if !ok {
		return mcp.NewToolResultError("Invalid argument: text must be a string"), nil
	}

	operation, ok := arguments["operation"].(string)
	if !ok {
		return mcp.NewToolResultError("Invalid argument: operation must be a string"), nil
	}

	var result interface{}
	var errMsg string

	switch operation {
	case "reverse":
		result = reverseString(text)
	case "uppercase":
		result = strings.ToUpper(text)
	case "lowercase":
		result = strings.ToLower(text)
	case "count":
		result = map[string]interface{}{
			"character_count": len(text),
			"byte_count":      len([]byte(text)),
			"rune_count":      len([]rune(text)),
		}
	case "remove_spaces":
		result = strings.ReplaceAll(text, " ", "")
	case "trim":
		result = strings.TrimSpace(text)
	case "word_count":
		words := strings.Fields(text)
		result = map[string]interface{}{
			"word_count": len(words),
			"words":      words,
		}
	default:
		errMsg = fmt.Sprintf("Unknown operation: %s. Supported operations: reverse, uppercase, lowercase, count, remove_spaces, trim, word_count", operation)
	}

	if errMsg != "" {
		return mcp.NewToolResultError(errMsg), nil
	}

	// 将结果序列化为 JSON 字符串
	resultJSON, err := json.Marshal(map[string]interface{}{
		"operation": operation,
		"input":     text,
		"result":    result,
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultJSON)), nil
}

// reverseString 反转字符串
func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
