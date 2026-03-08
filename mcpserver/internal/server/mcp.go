package server

import (
	"context"
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/mark3labs/mcp-go/mcp"
	"mcpserver/internal/protocol"
)

// buildInitializeResult 构建初始化结果（HTTP 和 SSE 共用）
func buildInitializeResult(srv *Server, requestID interface{}) interface{} {
	cfg := srv.Config()
	result := mcp.InitializeResult{
		ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
		ServerInfo: mcp.Implementation{
			Name:    cfg.ServerName,
			Version: cfg.ServerVersion,
		},
		Capabilities: mcp.ServerCapabilities{
			Tools: &struct {
				ListChanged bool `json:"listChanged,omitempty"`
			}{},
		},
	}
	return mcp.NewJSONRPCResultResponse(mcp.NewRequestId(requestID), result)
}

// buildListToolsResult 构建工具列表结果（HTTP 和 SSE 共用）
func buildListToolsResult(srv *Server, requestID interface{}) interface{} {
	tools := srv.ToolRegistry().ListTools()

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
	return mcp.NewJSONRPCResultResponse(mcp.NewRequestId(requestID), result)
}

// buildCallToolResult 构建工具调用结果（HTTP 和 SSE 共用）
func buildCallToolResult(srv *Server, ctx context.Context, requestID interface{}, params mcp.CallToolParams) interface{} {
	// 获取工具
	tool, err := srv.ToolRegistry().GetTool(params.Name)
	if err != nil {
		return protocol.ErrorResponse{
			JSONRPC: "2.0",
			ID:      requestID,
			Error: protocol.Error{
				Code:    -32601,
				Message: "Tool not found",
				Data:    err.Error(),
			},
		}
	}

	// 类型断言参数
	arguments, ok := params.Arguments.(map[string]interface{})
	if !ok {
		if params.Arguments == nil {
			arguments = make(map[string]interface{})
		} else {
			return protocol.ErrorResponse{
				JSONRPC: "2.0",
				ID:      requestID,
				Error: protocol.Error{
					Code:    -32602,
					Message: "Invalid params",
					Data:    "arguments must be a map",
				},
			}
		}
	}

	// 执行工具
	result, err := tool.Execute(ctx, arguments)
	if err != nil {
		errorResult := mcp.NewToolResultError(fmt.Sprintf("Tool execution error: %v", err))
		return mcp.NewJSONRPCResultResponse(mcp.NewRequestId(requestID), errorResult)
	}

	if result == nil {
		errorResult := mcp.NewToolResultError("Tool returned nil result")
		return mcp.NewJSONRPCResultResponse(mcp.NewRequestId(requestID), errorResult)
	}

	return mcp.NewJSONRPCResultResponse(mcp.NewRequestId(requestID), result)
}

// parseInitializeParams 解析初始化参数
func parseInitializeParams(paramsJSON []byte) (mcp.InitializeParams, error) {
	var params mcp.InitializeParams
	if len(paramsJSON) > 0 {
		if err := sonic.Unmarshal(paramsJSON, &params); err != nil {
			return params, err
		}
	}
	return params, nil
}

// parseCallToolParams 解析工具调用参数
func parseCallToolParams(paramsJSON []byte) (mcp.CallToolParams, error) {
	var params mcp.CallToolParams
	if err := sonic.Unmarshal(paramsJSON, &params); err != nil {
		return params, err
	}
	return params, nil
}
