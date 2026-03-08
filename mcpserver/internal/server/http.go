package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"mcpserver/internal/protocol"
)

// HTTPHandler HTTP 请求处理器
type HTTPHandler struct {
	server *Server
}

// NewHTTPHandler 创建 HTTP 处理器
func NewHTTPHandler(srv *Server) *HTTPHandler {
	return &HTTPHandler{
		server: srv,
	}
}

// HandleMCPRequest 处理 MCP 协议请求
func (h *HTTPHandler) HandleMCPRequest(c *gin.Context) {
	// 解析 JSON-RPC 请求
	var jsonrpcReq protocol.Request
	if err := c.ShouldBindJSON(&jsonrpcReq); err != nil {
		h.sendJSONRPCError(c, nil, -32700, "Parse error", err.Error())
		return
	}

	// 处理不同的请求方法
	switch jsonrpcReq.Method {
	case "initialize":
		h.handleInitialize(c, &jsonrpcReq)
	case "tools/list":
		h.handleListTools(c, &jsonrpcReq)
	case "tools/call":
		h.handleCallTool(c, &jsonrpcReq)
	default:
		h.sendJSONRPCError(c, jsonrpcReq.ID, -32601, "Method not found",
			fmt.Sprintf("Unknown method: %s", jsonrpcReq.Method))
	}
}

// handleInitialize 处理初始化请求
func (h *HTTPHandler) handleInitialize(c *gin.Context, request *protocol.Request) {
	_, err := parseInitializeParams(request.Params)
	if err != nil {
		h.sendJSONRPCError(c, request.ID, -32602, "Invalid params", err.Error())
		return
	}

	response := buildInitializeResult(h.server, request.ID)
	h.sendResponse(c, response)
}

// handleListTools 处理工具列表请求
func (h *HTTPHandler) handleListTools(c *gin.Context, request *protocol.Request) {
	response := buildListToolsResult(h.server, request.ID)
	h.sendResponse(c, response)
}

// handleCallTool 处理工具调用请求
func (h *HTTPHandler) handleCallTool(c *gin.Context, request *protocol.Request) {
	params, err := parseCallToolParams(request.Params)
	if err != nil {
		h.sendJSONRPCError(c, request.ID, -32602, "Invalid params", err.Error())
		return
	}

	ctx := c.Request.Context()
	response := buildCallToolResult(h.server, ctx, request.ID, params)

	// 如果是错误响应，需要特殊处理
	if errResp, ok := response.(protocol.ErrorResponse); ok {
		h.sendResponse(c, errResp)
		return
	}

	h.sendResponse(c, response)
}

// HandleHealthCheck 处理健康检查请求
func (h *HTTPHandler) HandleHealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"tools":   h.server.ToolRegistry().Count(),
		"version": h.server.Config().ServerVersion,
	})
}

// sendResponse 发送成功响应
func (h *HTTPHandler) sendResponse(c *gin.Context, response interface{}) {
	c.JSON(http.StatusOK, response)
}

// sendJSONRPCError 发送 JSON-RPC 错误响应
func (h *HTTPHandler) sendJSONRPCError(c *gin.Context, id interface{}, code int, message, data string) {
	response := protocol.ErrorResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: protocol.Error{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
	// MCP 协议错误也返回 200
	c.JSON(http.StatusOK, response)
}
