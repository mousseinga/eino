package main

import (
	"log"

	"ai-eino-interview-agent/mcp-moduel/internal/server"
	"ai-eino-interview-agent/mcp-moduel/internal/server/tools"
)

func main() {
	// 创建 stdio 服务器配置
	config := &server.ServerConfig{
		ServerName:    "mcp-stdio-server",
		ServerVersion: "1.0.0",
	}
	if err := config.Validate(); err != nil {
		log.Fatalf("Invalid server config: %v", err)
	}

	// 创建 stdio 服务器实例
	srv := server.NewStdioServer(config)

	// 注册默认工具
	defaultTools := tools.GetDefaultTools()
	for _, tool := range defaultTools {
		if t, ok := tool.(server.Tool); ok {
			if err := srv.RegisterTool(t); err != nil {
				log.Fatalf("Failed to register tool: %v", err)
			}
			// 注意：stdio 服务器的日志会输出到 stderr，避免干扰 JSON-RPC 通信
			log.Printf("Registered tool: %s", t.Name())
		}
	}

	// 启动服务器（会阻塞，直到 stdin 关闭或出错）
	if err := srv.Start(); err != nil {
		log.Fatalf("Failed to start stdio server: %v", err)
	}
}
