package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// CalculatorTool 计算器工具
type CalculatorTool struct{}

func NewCalculatorTool() *CalculatorTool {
	return &CalculatorTool{}
}

func (t *CalculatorTool) Name() string {
	return "calculate"
}

func (t *CalculatorTool) Description() string {
	return "Perform basic arithmetic calculations (add, subtract, multiply, divide)."
}

func (t *CalculatorTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"operation": map[string]interface{}{
				"type":        "string",
				"description": "The arithmetic operation to perform",
				"enum":        []string{"add", "subtract", "multiply", "divide"},
			},
			"a": map[string]interface{}{
				"type":        "number",
				"description": "First number",
			},
			"b": map[string]interface{}{
				"type":        "number",
				"description": "Second number",
			},
		},
		"required": []string{"operation", "a", "b"},
	}
}

func (t *CalculatorTool) Execute(ctx context.Context, arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	operation, ok := arguments["operation"].(string)
	if !ok {
		return mcp.NewToolResultError("Invalid argument: operation must be a string"), nil
	}

	// 处理数字类型（可能是 float64 或 int）
	var a, b float64

	if aVal, ok := arguments["a"].(float64); ok {
		a = aVal
	} else if aVal, ok := arguments["a"].(int); ok {
		a = float64(aVal)
	} else {
		return mcp.NewToolResultError("Invalid argument: a must be a number"), nil
	}

	if bVal, ok := arguments["b"].(float64); ok {
		b = bVal
	} else if bVal, ok := arguments["b"].(int); ok {
		b = float64(bVal)
	} else {
		return mcp.NewToolResultError("Invalid argument: b must be a number"), nil
	}

	var result float64
	var errMsg string

	switch operation {
	case "add":
		result = a + b
	case "subtract":
		result = a - b
	case "multiply":
		result = a * b
	case "divide":
		if b == 0 {
			errMsg = "Division by zero is not allowed"
		} else {
			result = a / b
		}
	default:
		errMsg = fmt.Sprintf("Unknown operation: %s", operation)
	}

	if errMsg != "" {
		return mcp.NewToolResultError(errMsg), nil
	}

	resultJSON, _ := json.Marshal(map[string]interface{}{
		"result":    result,
		"operation": operation,
		"a":         a,
		"b":         b,
	})

	return mcp.NewToolResultText(string(resultJSON)), nil
}
