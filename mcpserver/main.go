package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mcpserver/internal/config"
	"mcpserver/internal/server"
	"mcpserver/internal/tools"
)

func main() {
	// 解析命令行参数
	configPath := flag.String("config", "", "Path to config file (default: auto-detect)")
	flag.Parse()

	// 加载配置文件
	var cfg *config.Config
	var err error
	if *configPath != "" {
		cfg, err = config.LoadConfig(*configPath)
		if err != nil {
			log.Fatalf("Failed to load config from %s: %v", *configPath, err)
		}
		log.Printf("Loaded config from: %s", *configPath)
	} else {
		// 自动查找配置文件
		cfg, err = config.LoadConfig("")
		if err != nil {
			log.Printf("Warning: Could not load config file, using default config: %v", err)
			cfg = config.DefaultConfig()
		}
	}

	// 从配置文件创建服务器配置
	serverConfig := config.ServerConfigFromFile(cfg)

	// 创建服务器实例
	srv := server.NewServer(serverConfig)

	// 注册默认工具
	defaultTools := tools.GetDefaultTools()
	for _, tool := range defaultTools {
		if t, ok := tool.(server.Tool); ok {
			if err := srv.RegisterTool(t); err != nil {
				log.Fatalf("Failed to register tool %s: %v", t.Name(), err)
			}
			log.Printf("Registered tool: %s", t.Name())
		}
	}

	// 如果配置了天气 API 密钥，注册天气工具
	weatherAPIKey := cfg.GetAPIKey("weather_api_key")
	if weatherAPIKey != "" {
		weatherTool := tools.NewWeatherTool(weatherAPIKey)
		if err := srv.RegisterTool(weatherTool); err != nil {
			log.Printf("Warning: Failed to register weather tool: %v", err)
		} else {
			log.Printf("Registered tool: %s", weatherTool.Name())
		}
	} else {
		log.Printf("Weather API key not configured, skipping weather tool registration")
	}

	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// 在 goroutine 中启动服务器
	go func() {
		address := serverConfig.GetAddress()
		log.Printf("MCP Server starting on %s", address)
		log.Printf("Server name: %s v%s", serverConfig.ServerName, serverConfig.ServerVersion)
		log.Printf("Health check: http://%s/health", address)
		log.Printf("MCP endpoint: http://%s/mcp", address)
		if err := srv.Start(); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待中断信号
	<-sigChan
	log.Println("Shutting down server...")

	// 关闭
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Stop(ctx); err != nil {
		log.Printf("Error during server shutdown: %v", err)
	}

	log.Println("Server stopped")
}
