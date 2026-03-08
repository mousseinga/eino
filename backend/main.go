package main

import (
	"ai-eino-interview-agent/api/router"
	interviewRouter "ai-eino-interview-agent/api/router/interview"
	routerMiddleware "ai-eino-interview-agent/api/router/middleware"
	"ai-eino-interview-agent/internal/config"
	appMiddleware "ai-eino-interview-agent/internal/middleware"
	"ai-eino-interview-agent/internal/mq"
	"ai-eino-interview-agent/internal/repository"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/joho/godotenv"

	"github.com/cloudwego/hertz/pkg/app" // 新增这行
)

func main() {
	// 1. 加载 .env 文件（如果存在）
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
		log.Println("Application will use system environment variables or config.yaml defaults")
	} else {
		log.Println("Successfully loaded .env file")
	}

	// 2. 加载配置文件
	// 获取配置文件路径（相对于 main.go 所在目录）
	configPath := findConfigFile()
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// 3. 展开配置中的环境变量引用（${VAR_NAME}）
	cfg.ExpandEnv()
	log.Println("Environment variables expanded in configuration")

	// 4. 初始化数据库
	log.Println("Initializing database connection...")
	err = repository.InitDatabase(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	log.Println("Database initialized successfully")

	//5. 初始化Redis
	log.Println("Initializing Redis connection...")
	err = repository.InitRedis(cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}
	log.Println("Redis initialized successfully")

	//7. 初始化 Milvus Manager（向量数据库、Embedding、检索等服务）
	//log.Println("Initializing Milvus Manager...")
	//ctx := context.Background()
	//milvusManager, err := milvus.InitMilvusManager(ctx, cfg)
	//if err != nil {
	//	log.Fatalf("Failed to initialize Milvus Manager: %v", err)
	//}
	//// 进行健康检查
	//if err := milvusManager.HealthCheck(ctx); err != nil {
	//	log.Printf("Warning: Milvus health check failed: %v", err)
	//}
	//log.Println("Milvus Manager initialized successfully")

	// 8. 初始化消息队列（使用 Redis）
	log.Println("Initializing Redis message queue...")
	redisClient := repository.GetRedis()
	if redisClient == nil {
		log.Fatalf("Redis client not initialized")
	}
	messageQueue := mq.NewRedisQueue(redisClient)
	mq.InitMessageQueue(messageQueue)
	log.Println("Redis message queue initialized successfully")

	// 9. 启动消费者
	log.Println("Starting message consumer...")
	consumerCtx, cancelConsumer := context.WithCancel(context.Background())
	go func() {
		if err := mq.StartConsumer(consumerCtx); err != nil {
			log.Printf("Error starting consumer: %v", err)
		}
	}()
	// 给消费者一点时间启动
	time.Sleep(500 * time.Millisecond)
	defer cancelConsumer()

	// 初始化Hertz服务器
	s := server.Default(server.WithHostPorts(fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)))

	// 添加错误处理中间件（必须在最前面）
	// Recovery: 捕获请求处理中的 panic，防止服务崩溃
	// ErrorHandler: 统一处理业务错误，返回标准格式的错误响应
	s.Use(routerMiddleware.Recovery()) // 捕获 Panic

	// 添加全局CORS中间件，处理OPTIONS预检请求
	s.Use(func(ctx context.Context, c *app.RequestContext) {
		// 设置CORS头
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Cache-Control, X-Auth-Token")
		c.Header("Access-Control-Max-Age", "86400")

		// 如果是OPTIONS请求，直接返回204
		if string(c.Method()) == "OPTIONS" {
			log.Printf("[CORS] OPTIONS request: %s", c.Path())
			c.AbortWithStatus(204)
			return
		}

		c.Next(ctx)
	})

	s.Use(appMiddleware.JWTMiddlewareWithSkipper(interviewRouter.AuthSkipper()))
	router.GeneratedRegister(s)

	// 创建一个通道来监听中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 在单独的goroutine中启动服务器
	go func() {
		log.Printf("Server is running on %s:%d", cfg.Host, cfg.Port)
		if err := s.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待中断信号
	<-quit
	log.Println("Shutting down server...")

	// 关闭消费者
	cancelConsumer()
	log.Println("Message consumer stopped")

	// 关闭消息队列
	if err := messageQueue.Close(); err != nil {
		log.Printf("Warning: Failed to close message queue: %v", err)
	}
	log.Println("Message queue closed")

	// 关闭 Milvus Manager
	//if milvusManager != nil {
	//	if err := milvusManager.Close(); err != nil {
	//		log.Printf("Warning: Failed to close Milvus Manager: %v", err)
	//	}
	//}

	// 创建一个带有超时的上下文，用于关闭
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 关闭服务器
	if err := s.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}

// findConfigFile 查找配置文件路径
// 优先使用相对于 main.go 所在目录的 config.yaml
func findConfigFile() string {
	// 获取 main.go 源代码文件的路径
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		// 如果无法获取，尝试使用当前工作目录
		return "config.yaml"
	}

	// main.go 在 backend 目录下，config.yaml 也在 backend 目录下
	backendDir := filepath.Dir(currentFile)
	configPath := filepath.Join(backendDir, "config.yaml")

	// 检查文件是否存在
	if _, err := os.Stat(configPath); err == nil {
		return configPath
	}

	// 如果不存在，尝试当前工作目录
	if wd, err := os.Getwd(); err == nil {
		// 尝试 backend/config.yaml（从项目根目录运行）
		path1 := filepath.Join(wd, "backend", "config.yaml")
		if _, err := os.Stat(path1); err == nil {
			return path1
		}
		// 尝试 config.yaml（从 backend 目录运行）
		path2 := filepath.Join(wd, "config.yaml")
		if _, err := os.Stat(path2); err == nil {
			return path2
		}
	}

	// 默认返回相对于 main.go 的路径
	return configPath
}
