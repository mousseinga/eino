package client

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"net/http"
	"strings"
	"time"

	"ai-eino-interview-agent/mcp-moduel/internal/config"
	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
)

// initializeMCPClient 初始化 MCP 客户端
func initializeMCPClient(ctx context.Context, client *mcpclient.Client, clientName string) error {
	// 启动客户端
	if err := client.Start(ctx); err != nil {
		return fmt.Errorf("failed to start client: %w", err)
	}

	// 初始化协议
	initRequest := mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
			ClientInfo: mcp.Implementation{
				Name:    clientName,
				Version: "1.0.0",
			},
		},
	}

	_, err := client.Initialize(ctx, initRequest)
	if err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}

	return nil
}

// mergeArguments 合并所有参数到统一的映射中
func mergeArguments(args *InvocationArgs) map[string]interface{} {
	if args == nil {
		return make(map[string]interface{})
	}
	arguments := make(map[string]interface{})

	if args.Path != nil {
		for k, v := range args.Path {
			arguments[k] = v
		}
	}

	if args.Query != nil {
		for k, v := range args.Query {
			arguments[k] = v
		}
	}

	if args.Body != nil {
		for k, v := range args.Body {
			arguments[k] = v
		}
	}
	return arguments
}

// createMCPClient 创建 MCP 客户端
func createMCPClient(cfg *config.Config) (*mcpclient.Client, error) {
	// 根据传输类型创建不同的客户端
	switch cfg.TransportType {
	case config.TransportTypeHTTP:
		return createHTTPClient(cfg)
	case config.TransportTypeSSE:
		return createSSEClient(cfg)
	case config.TransportTypeStdio:
		return createStdioClient(cfg)
	default:
		// 默认使用 HTTP
		return createHTTPClient(cfg)
	}
}

// createHTTPClient 创建 HTTP 客户端
func createHTTPClient(cfg *config.Config) (*mcpclient.Client, error) {
	options := []transport.StreamableHTTPCOption{}

	// 如果配置了 API Key，添加认证头
	if cfg.APIKey != "" {
		options = append(options, transport.WithHTTPHeaders(map[string]string{
			"Authorization": "Bearer " + cfg.APIKey,
		}))
	}

	// 如果配置了超时时间，设置 HTTP 客户端超时
	if cfg.Timeout > 0 {
		httpClient := &http.Client{
			Timeout: time.Duration(cfg.Timeout) * time.Second,
		}
		options = append(options, transport.WithHTTPBasicClient(httpClient))
	}

	client, err := mcpclient.NewStreamableHttpClient(cfg.ServerURL, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create streamable http client: %w", err)
	}

	return client, nil
}

// createSSEClient 创建 SSE 客户端
func createSSEClient(cfg *config.Config) (*mcpclient.Client, error) {
	options := []transport.ClientOption{}

	// 如果配置了 API Key，添加认证头
	if cfg.APIKey != "" {
		options = append(options, transport.WithHeaders(map[string]string{
			"Authorization": "Bearer " + cfg.APIKey,
		}))
	}

	// 如果配置了超时时间，设置 HTTP 客户端超时
	if cfg.Timeout > 0 {
		httpClient := &http.Client{
			Timeout: time.Duration(cfg.Timeout) * time.Second,
		}
		options = append(options, transport.WithHTTPClient(httpClient))
	}

	client, err := mcpclient.NewSSEMCPClient(cfg.ServerURL, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create sse client: %w", err)
	}

	return client, nil
}

// createStdioClient 创建 stdio 客户端
func createStdioClient(cfg *config.Config) (*mcpclient.Client, error) {
	// 确保命令参数不为 nil
	args := cfg.CommandArgs
	if args == nil {
		args = []string{}
	}

	// 确保环境变量不为 nil
	env := cfg.CommandEnv
	if env == nil {
		env = []string{}
	}

	client, err := mcpclient.NewStdioMCPClient(cfg.Command, env, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to create stdio client: %w", err)
	}

	return client, nil
}

// isRetryableError 判断错误是否可重试
// 某些错误（如网络超时、临时服务器错误）可以重试，
// 而其他错误（如认证失败、参数错误）不应该重试。
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	// 网络相关错误通常可以重试
	retryableKeywords := []string{
		"timeout",
		"connection",
		"network",
		"temporary",
		"unavailable",
		"503",
		"502",
		"504",
	}

	errLower := strings.ToLower(errStr)
	for _, keyword := range retryableKeywords {
		if strings.Contains(errLower, strings.ToLower(keyword)) {
			return true
		}
	}

	// 认证错误、参数错误等不应该重试
	nonRetryableKeywords := []string{
		"authentication",
		"authorization",
		"unauthorized",
		"forbidden",
		"400",
		"401",
		"403",
		"404",
		"invalid",
		"parse",
	}

	for _, keyword := range nonRetryableKeywords {
		if strings.Contains(errLower, strings.ToLower(keyword)) {
			return false
		}
	}

	return true
}
