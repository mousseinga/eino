package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"mcpserver/internal/config"
	"mcpserver/internal/protocol"
)

// Server MCP 服务器实现（HTTP 和 SSE）
type Server struct {
	// 工具注册表
	toolRegistry *ToolRegistry

	// 服务器配置
	config *config.ServerConfig

	// 请求处理锁
	mu sync.RWMutex

	// Gin 引擎
	engine *gin.Engine

	// HTTP 服务器
	httpServer *http.Server

	// SSE 会话管理
	sseSessions sync.Map // map[string]*protocol.SSESession
}

// Config 返回服务器配置
func (s *Server) Config() *config.ServerConfig {
	return s.config
}

// ToolRegistry 返回工具注册表
func (s *Server) ToolRegistry() *ToolRegistry {
	return s.toolRegistry
}

// NewServer 初始化 MCP 服务器实例
func NewServer(config *config.ServerConfig) *Server {
	if config == nil {
		panic("server config cannot be nil")
	}

	// 设置 Gin 模式
	gin.SetMode(gin.ReleaseMode)

	// 创建 Gin 引擎
	engine := gin.New()
	engine.Use(gin.Recovery())

	// 添加日志中间件
	engine.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))

	// 配置 CORS
	if config.EnableCORS {
		corsConfig := cors.Config{
			AllowOrigins:     config.AllowedOrigins,
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		}
		engine.Use(cors.New(corsConfig))
	}

	s := &Server{
		toolRegistry: NewToolRegistry(),
		config:       config,
		engine:       engine,
	}

	// 注册路由
	s.setupRoutes()

	return s
}

// setupRoutes 设置路由
func (s *Server) setupRoutes() {
	// 创建 handlers
	httpHandler := NewHTTPHandler(s)
	sseHandler := NewSSEHandler(s)

	// 健康检查端点
	s.engine.GET("/health", httpHandler.HandleHealthCheck)

	// MCP 协议路由组
	mcpGroup := s.engine.Group("/mcp")
	{
		// HTTP 方式 MCP 端点
		mcpGroup.POST("", httpHandler.HandleMCPRequest)

		// SSE 路由组
		sseGroup := mcpGroup.Group("/sse")
		{
			// 建立 SSE 连接
			sseGroup.GET("", sseHandler.HandleSSEConnection)
			// 通过 SSE 发送请求
			sseGroup.POST("/:sessionId", sseHandler.HandleSSERequest)
		}
	}
}

// GetSSESession 获取 SSE 会话
func (s *Server) GetSSESession(sessionID string) (*protocol.SSESession, bool) {
	sessionInterface, ok := s.sseSessions.Load(sessionID)
	if !ok {
		return nil, false
	}
	return sessionInterface.(*protocol.SSESession), true
}

// StoreSSESession 存储 SSE 会话
func (s *Server) StoreSSESession(sessionID string, session *protocol.SSESession) {
	s.sseSessions.Store(sessionID, session)
}

// DeleteSSESession 删除 SSE 会话
func (s *Server) DeleteSSESession(sessionID string) {
	s.sseSessions.Delete(sessionID)
}

// RegisterTool 注册工具到服务器
func (s *Server) RegisterTool(tool Tool) error {
	return s.toolRegistry.Register(tool)
}

// RegisterTools 批量注册工具
func (s *Server) RegisterTools(tools []Tool) error {
	for _, tool := range tools {
		if err := s.toolRegistry.Register(tool); err != nil {
			return fmt.Errorf("failed to register tool %s: %w", tool.Name(), err)
		}
	}
	return nil
}

// Start 启动 HTTP 服务器
func (s *Server) Start() error {
	// 创建 HTTP 服务器
	s.httpServer = &http.Server{
		Addr:    s.config.GetAddress(),
		Handler: s.engine,
	}

	log.Printf("MCP Server starting on %s", s.config.GetAddress())
	return s.httpServer.ListenAndServe()
}

// Stop 停止服务器
func (s *Server) Stop(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}
	return s.httpServer.Shutdown(ctx)
}
