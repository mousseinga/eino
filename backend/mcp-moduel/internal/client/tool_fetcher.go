/*
 * Package client 提供 MCP 客户端功能
 *
 * 本文件实现了使用标准的 mcp-go 客户端来获取 MCP 工具列表的功能。
 */

package client

import (
	"context"
	"crypto/md5"
	"fmt"
	"time"

	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"

	"ai-eino-interview-agent/mcp-moduel/internal/config"
)

// einoToolFetcher 使用标准 mcp-go 客户端的工具获取器实现
type einoToolFetcher struct{}

// NewEinoToolFetcher 创建使用标准 mcp-go 客户端的工具获取器实例
func NewEinoToolFetcher() ToolFetcher {
	return &einoToolFetcher{}
}

// GetTools 使用标准 mcp-go 客户端获取工具列表
func (e *einoToolFetcher) GetTools(ctx context.Context, mcpConfig string) ([]*ToolInfo, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	}

	if mcpConfig == "" {
		return nil, fmt.Errorf("mcp config is required")
	}

	// 解析配置
	cfg, err := config.Parse(mcpConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to parse mcp config: %w", err)
	}

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(ctx, time.Duration(cfg.Timeout)*time.Second)
	defer cancel()

	// 创建 MCP 客户端
	var mcpClient *mcpclient.Client
	mcpClient, err = createMCPClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create mcp client: %w", err)
	}
	defer mcpClient.Close()

	// 初始化客户端
	if err := initializeMCPClient(ctx, mcpClient, "mcp-tool-fetcher"); err != nil {
		return nil, fmt.Errorf("failed to initialize mcp client: %w", err)
	}

	// 使用标准 mcp-go 客户端获取工具列表
	listToolsRequest := mcp.ListToolsRequest{}

	result, err := mcpClient.ListTools(ctx, listToolsRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to get tools from mcp server: %w", err)
	}

	// 转换为 ToolInfo 列表
	tools := make([]*ToolInfo, 0, len(result.Tools))
	for i, tool := range result.Tools {
		// 使用工具名称和索引生成唯一的 ID
		toolID := generateToolID(tool.Name, i)

		toolInfo := &ToolInfo{
			ID:          toolID,
			Name:        tool.Name,
			Description: tool.Description,
		}
		tools = append(tools, toolInfo)
	}

	return tools, nil
}

// generateToolID 生成唯一的工具 ID
func generateToolID(toolName string, index int) int64 {
	// 使用工具名称和时间戳生成哈希
	hash := md5.Sum([]byte(fmt.Sprintf("%s-%d-%d", toolName, index, time.Now().UnixNano())))
	// 取前 8 个字节转换为 int64
	var id int64
	for i := 0; i < 8 && i < len(hash); i++ {
		id = (id << 8) | int64(hash[i])
	}
	// 确保 ID 为正数
	if id < 0 {
		id = -id
	}
	// 如果 ID 为 0，使用索引和时间戳
	if id == 0 {
		id = int64(index+1)*1000000 + time.Now().UnixNano()%1000000
	}
	return id
}
