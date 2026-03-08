package main

import (
	"context"
	"testing"
	"time"

	"ai-eino-interview-agent/mcp-moduel/mcpmodule"
)

func TestMCPClient(t *testing.T) {
	// MCP 服务器地址（mcpserver 默认运行在 8080 端口）
	serverURL := "http://localhost:8080/mcp"

	// 创建 MCP 客户端
	client := mcpmodule.NewMCPClient(serverURL, 30, 3)

	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 1. 获取工具列表
	t.Run("获取工具列表", func(t *testing.T) {
		tools, err := client.GetTools(ctx)
		if err != nil {
			t.Fatalf("获取工具列表失败: %v", err)
		}

		if len(tools) == 0 {
			t.Fatalf("没有可用的工具")
		}

		t.Logf("获取到 %d 个工具:", len(tools))
		for _, tool := range tools {
			t.Logf("  - %s: %s", tool.Name, tool.Description)
		}
	})

	// 2. 测试计算器工具
	t.Run("测试计算器工具", func(t *testing.T) {
		testCalculator(t, client, ctx)
	})

	// 3. 测试文本处理工具
	t.Run("测试文本处理工具", func(t *testing.T) {
		testTextProcessor(t, client, ctx)
	})

	// 4. 测试时间工具
	t.Run("测试时间工具", func(t *testing.T) {
		testTimeTool(t, client, ctx)
	})

	// 5. 测试 Echo 工具
	t.Run("测试 Echo 工具", func(t *testing.T) {
		testEchoTool(t, client, ctx)
	})

	// 6. 测试天气工具（如果可用）
	t.Run("测试天气工具", func(t *testing.T) {
		testWeatherTool(t, client, ctx)
	})

	t.Log("所有测试完成!")
}

// testCalculator 测试计算器工具
func testCalculator(t *testing.T, client *mcpmodule.MCPClient, ctx context.Context) {
	// 加法
	result, err := client.CallTool(ctx, "calculate", map[string]interface{}{
		"operation": "add",
		"a":         10,
		"b":         20,
	})
	if err != nil {
		t.Errorf("计算器工具调用失败: %v", err)
		return
	}
	t.Logf("  10 + 20 = %s", result)

	// 乘法
	result, err = client.CallTool(ctx, "calculate", map[string]interface{}{
		"operation": "multiply",
		"a":         7,
		"b":         8,
	})
	if err != nil {
		t.Errorf("计算器工具调用失败: %v", err)
		return
	}
	t.Logf("  7 × 8 = %s", result)

	// 除法
	result, err = client.CallTool(ctx, "calculate", map[string]interface{}{
		"operation": "divide",
		"a":         100,
		"b":         4,
	})
	if err != nil {
		t.Errorf("计算器工具调用失败: %v", err)
		return
	}
	t.Logf("  100 ÷ 4 = %s", result)
}

// testTextProcessor 测试文本处理工具
func testTextProcessor(t *testing.T, client *mcpmodule.MCPClient, ctx context.Context) {
	// 反转字符串
	result, err := client.CallTool(ctx, "text_processor", map[string]interface{}{
		"text":      "Hello, World!",
		"operation": "reverse",
	})
	if err != nil {
		t.Errorf("文本处理工具调用失败: %v", err)
		return
	}
	t.Logf("  反转 'Hello, World!' = %s", result)

	// 转换为大写
	result, err = client.CallTool(ctx, "text_processor", map[string]interface{}{
		"text":      "hello world",
		"operation": "uppercase",
	})
	if err != nil {
		t.Errorf("文本处理工具调用失败: %v", err)
		return
	}
	t.Logf("  大写 'hello world' = %s", result)

	// 统计字符数
	result, err = client.CallTool(ctx, "text_processor", map[string]interface{}{
		"text":      "Hello, 世界!",
		"operation": "count",
	})
	if err != nil {
		t.Errorf("文本处理工具调用失败: %v", err)
		return
	}
	t.Logf("  字符数 'Hello, 世界!' = %s", result)
}

// testTimeTool 测试时间工具
func testTimeTool(t *testing.T, client *mcpmodule.MCPClient, ctx context.Context) {
	result, err := client.CallTool(ctx, "get_current_time", map[string]interface{}{
		"format": "rfc3339",
	})
	if err != nil {
		t.Errorf("时间工具调用失败: %v", err)
		return
	}
	t.Logf("  当前时间 (RFC3339): %s", result)

	result, err = client.CallTool(ctx, "get_current_time", map[string]interface{}{
		"format": "unix",
	})
	if err != nil {
		t.Errorf("时间工具调用失败: %v", err)
		return
	}
	t.Logf("  当前时间 (Unix): %s", result)
}

// testEchoTool 测试 Echo 工具
func testEchoTool(t *testing.T, client *mcpmodule.MCPClient, ctx context.Context) {
	result, err := client.CallTool(ctx, "echo", map[string]interface{}{
		"message": "Hello from MCP Client!",
	})
	if err != nil {
		t.Errorf("Echo 工具调用失败: %v", err)
		return
	}
	t.Logf("  Echo: %s", result)
}

// testWeatherTool 测试天气工具（如果可用）
func testWeatherTool(t *testing.T, client *mcpmodule.MCPClient, ctx context.Context) {
	// 检查工具是否存在
	tools, err := client.GetTools(ctx)
	if err != nil {
		t.Errorf("获取工具列表失败: %v", err)
		return
	}

	hasWeather := false
	for _, tool := range tools {
		if tool.Name == "get_weather" {
			hasWeather = true
			break
		}
	}

	if !hasWeather {
		t.Log("  天气工具不可用（可能需要配置 API 密钥）")
		return
	}

	// 调用天气工具
	result, err := client.CallTool(ctx, "get_weather", map[string]interface{}{
		"city":  "Beijing",
		"units": "metric",
	})
	if err != nil {
		t.Errorf("天气工具调用失败: %v", err)
		return
	}
	t.Logf("  北京天气: %s", result)
}
