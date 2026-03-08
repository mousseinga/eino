package client

import (
	"context"
)

// Invocation 调用接口
type Invocation interface {
	Do(ctx context.Context, args *InvocationArgs) (request string, resp string, err error)
}

// ToolFetcher 工具获取器接口
type ToolFetcher interface {
	GetTools(ctx context.Context, mcpConfig string) ([]*ToolInfo, error)
}

// InvocationArgs 调用参数结构
type InvocationArgs struct {
	// Tool 工具信息对象
	Tool *ToolInfo

	// Path 路径参数映射
	Path map[string]any

	// Query 查询参数映射
	Query map[string]any

	// Body 请求体参数映射
	Body map[string]any

	// McpConfig MCP 配置的 JSON 字符串
	McpConfig string
}

// ToolInfo 工具信息结构
type ToolInfo struct {
	ID          int64
	Name        string
	Description string
}
