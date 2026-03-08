package client

import (
	"context"
	"fmt"
	"time"

	"github.com/bytedance/sonic"
	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"

	"ai-eino-interview-agent/mcp-moduel/internal/config"
)

// mcpCallImpl MCP 调用实现
// 实现了 Invocation 接口，提供基于 JSON-RPC 2.0 的 MCP 协议调用功能。
type mcpCallImpl struct{}

// NewMcpCallImpl 创建 MCP 调用实现实例
func NewMcpCallImpl() Invocation {
	return &mcpCallImpl{}
}

// Do 执行 MCP 工具调用
func (m *mcpCallImpl) Do(ctx context.Context, args *InvocationArgs) (request string, resp string, err error) {
	//参数验证
	if args == nil {
		return "", "", fmt.Errorf("invocation args cannot be nil")
	}
	if ctx == nil {
		return "", "", fmt.Errorf("context cannot be nil")
	}
	if args.McpConfig == "" {
		return "", "", fmt.Errorf("mcp config is required")
	}
	cfg, err := config.Parse(args.McpConfig)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse mcp config: %w", err)
	}
	ctx, cancel := context.WithTimeout(ctx, time.Duration(cfg.Timeout)*time.Second)
	defer cancel()

	// 创建客户端
	mcpClient, err := createMCPClient(cfg)
	if err != nil {
		return "", "", fmt.Errorf("failed to create mcp client: %w", err)
	}
	defer mcpClient.Close()

	// 初始化客户端
	if err := initializeMCPClient(ctx, mcpClient, "mcp-go"); err != nil {
		return "", "", fmt.Errorf("failed to initialize mcp client: %w", err)
	}

	// 构建请求
	mcpReq, requestJSON, err := m.buildRequest(args)
	if err != nil {
		return "", "", fmt.Errorf("failed to build request: %w", err)
	}

	// 执行调用（带重试）
	responseJSON, err := m.callWithRetry(ctx, mcpClient, mcpReq, cfg.RetryTimes)
	if err != nil {
		return requestJSON, "", err
	}

	return requestJSON, responseJSON, nil
}

// buildRequest 构建 MCP 请求
func (m *mcpCallImpl) buildRequest(args *InvocationArgs) (*mcp.CallToolRequest, string, error) {
	// 验证工具信息
	if args.Tool == nil || args.Tool.Name == "" {
		return nil, "", fmt.Errorf("tool name is required")
	}

	// 合并参数
	arguments := mergeArguments(args)

	// 构建请求
	request := &mcp.CallToolRequest{
		Request: mcp.Request{
			Method: "tools/call",
		},
		Params: mcp.CallToolParams{
			Name:      args.Tool.Name,
			Arguments: arguments,
		},
	}

	// 序列化请求
	requestJSON, err := sonic.MarshalString(request)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal request: %w", err)
	}

	return request, requestJSON, nil
}

// callWithRetry 执行带重试机制的工具调用
func (m *mcpCallImpl) callWithRetry(ctx context.Context, client *mcpclient.Client, req *mcp.CallToolRequest, retryTimes int) (string, error) {
	var lastErr error

	for attempt := 0; attempt <= retryTimes; attempt++ {
		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("context cancelled: %w", ctx.Err())
		default:
		}

		// 执行调用
		result, err := client.CallTool(ctx, *req)
		if err == nil {
			// 序列化响应
			responseJSON, err := m.serializeResponse(result)
			if err != nil {
				return "", fmt.Errorf("failed to serialize response: %w", err)
			}
			return responseJSON, nil
		}

		lastErr = err

		// 如果错误不可重试，立即返回
		if !isRetryableError(err) {
			return "", fmt.Errorf("non-retryable error: %w", err)
		}

		// 如果已达到最大重试次数，返回错误
		if attempt >= retryTimes {
			return "", fmt.Errorf("failed to call tool after %d attempts: %w", retryTimes+1, err)
		}

		// 指数退避：第 1 次重试等待 1 秒，第 2 次等待 2 秒，第 3 次等待 4 秒...
		backoff := time.Duration(1<<uint(attempt)) * time.Second
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("context cancelled during retry: %w", ctx.Err())
		case <-time.After(backoff):
		}
	}

	return "", lastErr
}

// serializeResponse 序列化响应结果
func (m *mcpCallImpl) serializeResponse(result *mcp.CallToolResult) (string, error) {
	// 验证结果不为 nil
	if result == nil {
		return "", fmt.Errorf("result cannot be nil")
	}

	// 如果结果包含错误，返回错误信息
	if result.IsError {
		// 提取文本内容作为错误消息
		errorMsg := m.extractTextContent(result.Content)
		if errorMsg != "" {
			return "", fmt.Errorf("mcp tool returned error: %s", errorMsg)
		}
		return "", fmt.Errorf("mcp tool returned error")
	}

	// 优先提取文本内容
	textContent := m.extractTextContent(result.Content)
	if textContent != "" {
		return textContent, nil
	}

	// 如果没有文本内容，返回完整结果的 JSON
	resultJSON, err := sonic.MarshalString(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}
	return resultJSON, nil
}

// extractTextContent 从内容数组中提取文本内容
func (m *mcpCallImpl) extractTextContent(contents []mcp.Content) string {
	var texts []string
	for _, content := range contents {
		text := mcp.GetTextFromContent(content)
		if text != "" {
			texts = append(texts, text)
		}
	}
	if len(texts) == 0 {
		return ""
	}
	if len(texts) == 1 {
		return texts[0]
	}
	// 连接多个文本
	result := ""
	for i, text := range texts {
		if i > 0 {
			result += "\n"
		}
		result += text
	}
	return result
}
