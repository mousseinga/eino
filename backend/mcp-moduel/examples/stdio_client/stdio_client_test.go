package stdio_client

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"ai-eino-interview-agent/mcp-moduel/internal/client"
)

func TestStdioClient(t *testing.T) {
	serverExe, err := findServerExe()
	if err != nil {
		t.Fatalf("查找 stdio 服务器失败: %v", err)
	}

	mcpConfigJSON := fmt.Sprintf(`{
		"transport_type": "stdio",
		"command": "%s",
		"command_args": [],
		"command_env": [],
		"timeout": 30,
		"retry_times": 3
	}`, serverExe)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 工具初始化
	toolFetcher := client.NewEinoToolFetcher()
	tools, err := toolFetcher.GetTools(ctx, mcpConfigJSON)
	if err != nil {
		t.Fatalf("获取工具列表失败: %v", err)
	}
	if len(tools) == 0 {
		t.Fatalf("没有可用的工具")
	}
	t.Logf("获取到 %d 个工具\n", len(tools))
	for _, tool := range tools {
		t.Logf("  - %s: %s\n", tool.Name, tool.Description)
	}

	// 测试 Echo 工具
	t.Run("测试 Echo 工具", func(t *testing.T) {
		selectedTool := findTool(tools, "echo")
		if selectedTool == nil {
			selectedTool = tools[0]
		}

		invoker := client.NewMcpCallImpl()
		args := &client.InvocationArgs{
			Tool: selectedTool,
			Body: map[string]any{
				"message": "Hello, MCP Stdio Server!",
			},
			McpConfig: mcpConfigJSON,
		}

		_, response, err := invoker.Do(ctx, args)
		if err != nil {
			t.Fatalf("工具调用失败: %v", err)
		}

		t.Logf("\n工具 '%s' 调用结果:\n%s\n", selectedTool.Name, response)
	})

	// 测试计算器工具
	t.Run("测试计算器工具", func(t *testing.T) {
		calcTool := findTool(tools, "calculate")
		if calcTool == nil {
			t.Skip("计算器工具不可用")
			return
		}

		invoker := client.NewMcpCallImpl()
		calcArgs := &client.InvocationArgs{
			Tool: calcTool,
			Body: map[string]any{
				"operation": "add",
				"a":         10,
				"b":         20,
			},
			McpConfig: mcpConfigJSON,
		}

		_, calcResponse, err := invoker.Do(ctx, calcArgs)
		if err != nil {
			t.Errorf("计算器工具调用失败: %v", err)
			return
		}
		t.Logf("\n计算器工具调用结果:\n%s\n", calcResponse)
	})
}

// 获取服务器可执行文件路径
func findServerExe() (string, error) {
	if envPath := os.Getenv("STDIO_SERVER_PATH"); envPath != "" {
		if _, err := os.Stat(envPath); err == nil {
			return envPath, nil
		}
		return "", fmt.Errorf("STDIO_SERVER_PATH is set but file not found: %s", envPath)
	}

	var possiblePaths []string

	if _, currentFile, _, ok := runtime.Caller(0); ok {
		currentDir := filepath.Dir(currentFile)
		possiblePaths = append(possiblePaths,
			filepath.Join(currentDir, "..", "..", "internal", "server", "stdio_server", "stdio_server"),
			filepath.Join(currentDir, "..", "stdio_server", "stdio_server"),
			filepath.Join(filepath.Dir(currentDir), "stdio_server", "stdio_server"),
		)
	}

	if wd, err := os.Getwd(); err == nil {
		possiblePaths = append(possiblePaths,
			filepath.Join(wd, "internal", "server", "stdio_server", "stdio_server"),
			filepath.Join(wd, "..", "stdio_server", "stdio_server"),
			filepath.Join(wd, "stdio_server", "stdio_server"),
			filepath.Join(wd, "examples", "stdio_server", "stdio_server"),
			filepath.Join(wd, "backend", "mcp-moduel", "examples", "stdio_server", "stdio_server"),
			filepath.Join(wd, "mcp-moduel", "examples", "stdio_server", "stdio_server"),
		)
	}

	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		possiblePaths = append(possiblePaths,
			filepath.Join(exeDir, "stdio_server"),
			filepath.Join(exeDir, "..", "internal", "server", "stdio_server", "stdio_server"),
			filepath.Join(exeDir, "..", "stdio_server", "stdio_server"),
			filepath.Join(exeDir, "..", "examples", "stdio_server", "stdio_server"),
		)
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			if absPath, err := filepath.Abs(path); err == nil {
				return absPath, nil
			}
		}
	}

	return "", fmt.Errorf("找不到 stdio_server，请设置 STDIO_SERVER_PATH 环境变量或先构建服务器")
}

// 获取工具
func findTool(tools []*client.ToolInfo, name string) *client.ToolInfo {
	for _, tool := range tools {
		if tool.Name == name {
			return tool
		}
	}
	return nil
}
