package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"mcpserver/internal/protocol"
)

// SSEHandler SSE 请求处理器
type SSEHandler struct {
	server *Server
}

// NewSSEHandler 创建 SSE 处理器
func NewSSEHandler(srv *Server) *SSEHandler {
	return &SSEHandler{
		server: srv,
	}
}

// HandleSSEConnection 处理 SSE 连接
func (h *SSEHandler) HandleSSEConnection(c *gin.Context) {
	// 创建 SSE 会话
	sessionID := uuid.New().String()
	session := &protocol.SSESession{
		ID:          sessionID,
		EventChan:   make(chan string, 100),
		RequestChan: make(chan *protocol.Request, 100),
		Done:        make(chan struct{}),
	}

	h.server.StoreSSESession(sessionID, session)

	// 设置 SSE 响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Cache-Control")

	// 发送会话 ID（使用标准 SSE 格式）
	sessionData := fmt.Sprintf(`{"type":"session","sessionId":"%s"}`, sessionID)
	fmt.Fprintf(c.Writer, "data: %s\n\n", sessionData)
	c.Writer.Flush()

	// 启动响应处理协程
	go h.handleSSEResponses(session, c)

	// 等待连接关闭
	<-c.Request.Context().Done()
	close(session.Done)
	h.server.DeleteSSESession(sessionID)
	log.Printf("SSE session %s closed", sessionID)
}

// HandleSSERequest 处理通过 SSE 发送的请求
func (h *SSEHandler) HandleSSERequest(c *gin.Context) {
	// 从 URL 参数获取会话 ID
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Session ID required"})
		return
	}

	// 获取会话
	session, ok := h.server.GetSSESession(sessionID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}

	// 解析 JSON-RPC 请求
	var jsonrpcReq protocol.Request
	if err := c.ShouldBindJSON(&jsonrpcReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON-RPC request"})
		return
	}

	// 将请求发送到会话的请求通道
	select {
	case session.RequestChan <- &jsonrpcReq:
		c.JSON(http.StatusAccepted, gin.H{"message": "Request queued"})
	default:
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Request queue full"})
	}
}

// handleSSEResponses 处理 SSE 响应
func (h *SSEHandler) handleSSEResponses(session *protocol.SSESession, c *gin.Context) {
	// 处理请求
	go func() {
		for {
			select {
			case <-session.Done:
				return
			case req := <-session.RequestChan:
				h.processSSERequest(session, req)
			}
		}
	}()

	// 发送事件
	for {
		select {
		case <-session.Done:
			return
		case <-c.Request.Context().Done():
			return
		case event := <-session.EventChan:
			// 使用标准 SSE 格式发送事件
			fmt.Fprintf(c.Writer, "data: %s\n\n", event)
			c.Writer.Flush()
		}
	}
}

// processSSERequest 处理 SSE 请求
func (h *SSEHandler) processSSERequest(session *protocol.SSESession, request *protocol.Request) {
	var response interface{}

	// 处理不同的请求方法
	switch request.Method {
	case "initialize":
		response = h.handleInitializeSSE(request)
	case "tools/list":
		response = h.handleListToolsSSE(request)
	case "tools/call":
		response = h.handleCallToolSSE(request)
	default:
		response = protocol.ErrorResponse{
			JSONRPC: "2.0",
			ID:      request.ID,
			Error: protocol.Error{
				Code:    -32601,
				Message: "Method not found",
				Data:    fmt.Sprintf("Unknown method: %s", request.Method),
			},
		}
	}

	// 序列化响应
	responseJSON, err := json.Marshal(response)
	if err != nil {
		log.Printf("Failed to marshal SSE response: %v", err)
		return
	}

	// 发送响应
	select {
	case session.EventChan <- string(responseJSON):
	case <-time.After(5 * time.Second):
		log.Printf("Timeout sending SSE response for request %v", request.ID)
	case <-session.Done:
		return
	}
}

// handleInitializeSSE 处理 SSE 初始化请求
func (h *SSEHandler) handleInitializeSSE(request *protocol.Request) interface{} {
	_, err := parseInitializeParams(request.Params)
	if err != nil {
		return protocol.ErrorResponse{
			JSONRPC: "2.0",
			ID:      request.ID,
			Error: protocol.Error{
				Code:    -32602,
				Message: "Invalid params",
				Data:    err.Error(),
			},
		}
	}
	return buildInitializeResult(h.server, request.ID)
}

// handleListToolsSSE 处理 SSE 工具列表请求
func (h *SSEHandler) handleListToolsSSE(request *protocol.Request) interface{} {
	return buildListToolsResult(h.server, request.ID)
}

// handleCallToolSSE 处理 SSE 工具调用请求
func (h *SSEHandler) handleCallToolSSE(request *protocol.Request) interface{} {
	params, err := parseCallToolParams(request.Params)
	if err != nil {
		return protocol.ErrorResponse{
			JSONRPC: "2.0",
			ID:      request.ID,
			Error: protocol.Error{
				Code:    -32602,
				Message: "Invalid params",
				Data:    err.Error(),
			},
		}
	}

	ctx := context.Background()
	return buildCallToolResult(h.server, ctx, request.ID, params)
}
