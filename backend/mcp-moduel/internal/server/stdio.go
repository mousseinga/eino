package server

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	"github.com/bytedance/sonic"
	"github.com/mark3labs/mcp-go/mcp"
)

// ServerConfig 服务器配置（用于 stdio 服务器）
type ServerConfig struct {
	// ServerName 服务器名称
	ServerName string

	// ServerVersion 服务器版本
	ServerVersion string
}

// DefaultServerConfig 返回默认服务器配置
func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		ServerName:    "mcp-stdio-server",
		ServerVersion: "1.0.0",
	}
}

// Validate 验证配置
func (c *ServerConfig) Validate() error {
	if c.ServerName == "" {
		c.ServerName = "mcp-stdio-server"
	}
	if c.ServerVersion == "" {
		c.ServerVersion = "1.0.0"
	}
	return nil
}

// JSONRPCRequest JSON-RPC 2.0 请求结构
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// JSONRPCError JSON-RPC 2.0 错误结构
type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data,omitempty"`
}

// JSONRPCErrorResponse JSON-RPC 2.0 错误响应
type JSONRPCErrorResponse struct {
	JSONRPC string       `json:"jsonrpc"`
	ID      interface{}  `json:"id"`
	Error   JSONRPCError `json:"error"`
}

// StdioServer stdio 服务器实现
type StdioServer struct {
	toolRegistry *ToolRegistry
	config       *ServerConfig
	mu           sync.RWMutex
	stdin        *bufio.Scanner
	stdout       io.Writer
}

// NewStdioServer 创建新的 stdio 服务器实例
func NewStdioServer(config *ServerConfig) *StdioServer {
	if config == nil {
		config = DefaultServerConfig()
	}

	return &StdioServer{
		toolRegistry: NewToolRegistry(),
		config:       config,
		stdin:        bufio.NewScanner(os.Stdin),
		stdout:       os.Stdout,
	}
}

// RegisterTool 注册工具到服务器
func (s *StdioServer) RegisterTool(tool Tool) error {
	return s.toolRegistry.Register(tool)
}

// RegisterTools 批量注册工具
func (s *StdioServer) RegisterTools(tools []Tool) error {
	for _, tool := range tools {
		if err := s.toolRegistry.Register(tool); err != nil {
			return fmt.Errorf("failed to register tool %s: %w", tool.Name(), err)
		}
	}
	return nil
}

// Start 启动 stdio 服务器
func (s *StdioServer) Start() error {
	log.Printf("MCP Stdio Server starting")

	// 从标准输入读取 JSON-RPC 请求
	for s.stdin.Scan() {
		line := s.stdin.Text()
		if line == "" {
			continue
		}

		// 解析 JSON-RPC 请求
		var jsonrpcReq JSONRPCRequest
		if err := json.Unmarshal([]byte(line), &jsonrpcReq); err != nil {
			s.sendJSONRPCError(nil, -32700, "Parse error", err.Error())
			continue
		}

		// 处理请求
		go s.handleRequest(&jsonrpcReq)
	}

	// 检查扫描错误
	if err := s.stdin.Err(); err != nil {
		return fmt.Errorf("error reading from stdin: %w", err)
	}

	return nil
}

// Stop 停止服务器
func (s *StdioServer) Stop(ctx context.Context) error {
	// stdio 服务器通过关闭标准输入来停止
	return nil
}

// handleRequest 处理 JSON-RPC 请求
func (s *StdioServer) handleRequest(request *JSONRPCRequest) {
	var response interface{}

	// 处理不同的请求方法
	switch request.Method {
	case "initialize":
		response = s.handleInitialize(request)
	case "tools/list":
		response = s.handleListTools(request)
	case "tools/call":
		response = s.handleCallTool(request)
	default:
		response = JSONRPCErrorResponse{
			JSONRPC: "2.0",
			ID:      request.ID,
			Error: JSONRPCError{
				Code:    -32601,
				Message: "Method not found",
				Data:    fmt.Sprintf("Unknown method: %s", request.Method),
			},
		}
	}

	// 发送响应
	s.sendResponse(response)
}

// handleInitialize 处理初始化请求
func (s *StdioServer) handleInitialize(request *JSONRPCRequest) interface{} {
	var params mcp.InitializeParams
	if len(request.Params) > 0 {
		if err := sonic.Unmarshal(request.Params, &params); err != nil {
			return JSONRPCErrorResponse{
				JSONRPC: "2.0",
				ID:      request.ID,
				Error: JSONRPCError{
					Code:    -32602,
					Message: "Invalid params",
					Data:    err.Error(),
				},
			}
		}
	}

	result := mcp.InitializeResult{
		ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
		ServerInfo: mcp.Implementation{
			Name:    s.config.ServerName,
			Version: s.config.ServerVersion,
		},
		Capabilities: mcp.ServerCapabilities{
			Tools: &struct {
				ListChanged bool `json:"listChanged,omitempty"`
			}{},
		},
	}

	requestID := mcp.NewRequestId(request.ID)
	return mcp.NewJSONRPCResultResponse(requestID, result)
}

// handleListTools 处理工具列表请求
func (s *StdioServer) handleListTools(request *JSONRPCRequest) interface{} {
	tools := s.toolRegistry.ListTools()

	// 转换为 MCP 格式的工具列表
	mcpTools := make([]mcp.Tool, 0, len(tools))
	for _, tool := range tools {
		schema := tool.InputSchema()
		schemaBytes, _ := sonic.Marshal(schema)

		mcpTool := mcp.NewToolWithRawSchema(
			tool.Name(),
			tool.Description(),
			schemaBytes,
		)
		mcpTools = append(mcpTools, mcpTool)
	}

	result := mcp.ListToolsResult{
		Tools: mcpTools,
	}

	requestID := mcp.NewRequestId(request.ID)
	return mcp.NewJSONRPCResultResponse(requestID, result)
}

// handleCallTool 处理工具调用请求
func (s *StdioServer) handleCallTool(request *JSONRPCRequest) interface{} {
	var params mcp.CallToolParams
	if err := sonic.Unmarshal(request.Params, &params); err != nil {
		return JSONRPCErrorResponse{
			JSONRPC: "2.0",
			ID:      request.ID,
			Error: JSONRPCError{
				Code:    -32602,
				Message: "Invalid params",
				Data:    err.Error(),
			},
		}
	}

	// 获取工具
	tool, err := s.toolRegistry.GetTool(params.Name)
	if err != nil {
		return JSONRPCErrorResponse{
			JSONRPC: "2.0",
			ID:      request.ID,
			Error: JSONRPCError{
				Code:    -32601,
				Message: "Tool not found",
				Data:    err.Error(),
			},
		}
	}

	// 执行工具
	ctx := context.Background()

	// 类型断言参数
	arguments, ok := params.Arguments.(map[string]interface{})
	if !ok {
		if params.Arguments == nil {
			arguments = make(map[string]interface{})
		} else {
			return JSONRPCErrorResponse{
				JSONRPC: "2.0",
				ID:      request.ID,
				Error: JSONRPCError{
					Code:    -32602,
					Message: "Invalid params",
					Data:    "arguments must be a map",
				},
			}
		}
	}

	result, err := tool.Execute(ctx, arguments)
	if err != nil {
		errorResult := mcp.NewToolResultError(fmt.Sprintf("Tool execution error: %v", err))
		requestID := mcp.NewRequestId(request.ID)
		return mcp.NewJSONRPCResultResponse(requestID, errorResult)
	}

	if result == nil {
		errorResult := mcp.NewToolResultError("Tool returned nil result")
		requestID := mcp.NewRequestId(request.ID)
		return mcp.NewJSONRPCResultResponse(requestID, errorResult)
	}

	requestID := mcp.NewRequestId(request.ID)
	return mcp.NewJSONRPCResultResponse(requestID, result)
}

// sendResponse 发送响应
func (s *StdioServer) sendResponse(response interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	responseJSON, err := json.Marshal(response)
	if err != nil {
		log.Printf("Failed to marshal response: %v", err)
		return
	}

	// 写入标准输出
	fmt.Fprintf(s.stdout, "%s\n", string(responseJSON))
}

// sendJSONRPCError 发送 JSON-RPC 错误响应
func (s *StdioServer) sendJSONRPCError(id interface{}, code int, message, data string) {
	response := JSONRPCErrorResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: JSONRPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}

	s.sendResponse(response)
}
