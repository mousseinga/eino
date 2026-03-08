package mcpmodule

import (
	"context"
	"fmt"
	"time"

	"ai-eino-interview-agent/mcp-moduel/internal/client"
	"ai-eino-interview-agent/mcp-moduel/internal/config"
)

// MCPClient MCP 客户端包装
// 提供简化的接口来使用 MCP 服务器
type MCPClient struct {
	config      string
	toolFetcher client.ToolFetcher
	invoker     client.Invocation
}

// NewMCPClient 创建新的 MCP 客户端
// serverURL: MCP 服务器地址，例如 "http://localhost:8080/mcp"
// timeout: 超时时间（秒），默认 30
// retryTimes: 重试次数，默认 3
func NewMCPClient(serverURL string, timeout, retryTimes int) *MCPClient {
	if timeout <= 0 {
		timeout = 30
	}
	if retryTimes < 0 {
		retryTimes = 3
	}

	cfg := &config.Config{
		TransportType: config.TransportTypeHTTP,
		ServerURL:     serverURL,
		Timeout:       timeout,
		RetryTimes:    retryTimes,
	}

	cfgJSON, _ := config.Marshal(cfg)

	return &MCPClient{
		config:      cfgJSON,
		toolFetcher: client.NewEinoToolFetcher(),
		invoker:     client.NewMcpCallImpl(),
	}
}

// GetTools 获取服务器上可用的工具列表
func (c *MCPClient) GetTools(ctx context.Context) ([]*client.ToolInfo, error) {
	return c.toolFetcher.GetTools(ctx, c.config)
}

// CallTool 调用指定的工具
// toolName: 工具名称
// arguments: 工具参数
func (c *MCPClient) CallTool(ctx context.Context, toolName string, arguments map[string]interface{}) (string, error) {
	// 首先获取工具列表以验证工具是否存在
	tools, err := c.GetTools(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get tools: %w", err)
	}

	// 查找工具
	var toolInfo *client.ToolInfo
	for _, tool := range tools {
		if tool.Name == toolName {
			toolInfo = tool
			break
		}
	}

	if toolInfo == nil {
		return "", fmt.Errorf("tool '%s' not found", toolName)
	}

	// 构建调用参数
	args := &client.InvocationArgs{
		Tool:      toolInfo,
		Body:      arguments,
		McpConfig: c.config,
	}

	// 执行调用
	_, response, err := c.invoker.Do(ctx, args)
	if err != nil {
		return "", fmt.Errorf("failed to call tool: %w", err)
	}

	return response, nil
}

// CallToolWithTimeout 调用工具并指定超时时间
func (c *MCPClient) CallToolWithTimeout(ctx context.Context, toolName string, arguments map[string]interface{}, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return c.CallTool(ctx, toolName, arguments)
}
