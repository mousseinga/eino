package server

import (
	"context"
	"fmt"
	"sync"

	"github.com/mark3labs/mcp-go/mcp"
)

// Tool 工具接口
type Tool interface {
	// Name 返回工具名称
	Name() string

	// Description 返回工具描述
	Description() string

	// InputSchema 返回工具的输入参数 schema
	InputSchema() map[string]interface{}

	// Execute 执行工具
	Execute(ctx context.Context, arguments map[string]interface{}) (*mcp.CallToolResult, error)
}

// ToolRegistry 定义工具仓库
type ToolRegistry struct {
	tools map[string]Tool
	mu    sync.RWMutex
}

// NewToolRegistry 初始化工具仓库
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]Tool),
	}
}

// Register 注册工具
func (r *ToolRegistry) Register(tool Tool) error {
	if tool == nil {
		return fmt.Errorf("tool cannot be nil")
	}

	name := tool.Name()
	if name == "" {
		return fmt.Errorf("tool name cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tools[name]; exists {
		return fmt.Errorf("tool %s already registered", name)
	}

	r.tools[name] = tool
	return nil
}

// GetTool 获取工具
func (r *ToolRegistry) GetTool(name string) (Tool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tool, exists := r.tools[name]
	if !exists {
		return nil, fmt.Errorf("tool %s not found", name)
	}

	return tool, nil
}

// ListTools 列出所有工具
func (r *ToolRegistry) ListTools() []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tools := make([]Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}

	return tools
}

// Count 返回工具数量
func (r *ToolRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.tools)
}

// Unregister 注销工具
func (r *ToolRegistry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tools[name]; !exists {
		return fmt.Errorf("tool %s not found", name)
	}

	delete(r.tools, name)
	return nil
}
