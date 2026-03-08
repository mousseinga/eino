package main

import (
	// "ai-eino-interview-agent/internal/config"
	// "ai-eino-interview-agent/internal/eino/milvus"
	// "ai-eino-interview-agent/internal/repository"

	"ai-eino-interview-agent/internal/service/common"
	"bufio"
	"fmt"
	"log"
	"os"

	// "path/filepath"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"

	// "github.com/joho/godotenv"
	"golang.org/x/net/context"
)

func main() {
	// 测试解密
	testDecryption()

	// 初始化配置和数据库
	// initApp()

	ctx := context.Background()

	// 1. 创建 Milvus Agent
	// var agent adk.Agent

	// Hardcoded credentials
	apiKey := "c1c8f7ce-266f-4af5-a832-9a8457f36e74"
	baseURL := "https://ark.cn-beijing.volces.com/api/v3"
	modelName := "doubao-seed-1-6-flash-250828"

	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:  apiKey,
		Model:   modelName,
		BaseURL: baseURL,
	})
	if err != nil {
		log.Fatalf("Failed to create chat model: %v", err)
	}

	agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:          "TestAgent",
		Description:   "聊天测试小助手",
		Instruction:   "你是一个聊天测试小助手，你的任务是回答用户的问题",
		Model:         chatModel,
		MaxIterations: 15,
	})
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	// 2. 创建 Runner
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent: agent,
	})

	// 3. 命令行交互
	fmt.Println("====== Milvus Agent 交互测试 ======")
	fmt.Println("输入问题进行测试，输入 'exit' 或 'quit' 退出")
	fmt.Println("=====================================")

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("\n👤 你: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("读取输入失败: %v", err)
			continue
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		// 退出命令
		if input == "exit" || input == "quit" {
			fmt.Println("再见！")
			break
		}

		// 构建消息
		messages := []adk.Message{
			schema.UserMessage(input),
		}

		// 运行 Agent
		iter := runner.Run(ctx, messages)

		fmt.Print("\n🤖 Agent: ")
		for {
			event, ok := iter.Next()
			if !ok {
				break
			}

			if event.Err != nil {
				log.Printf("Agent 执行出错: %v", event.Err)
				break
			}

			// 打印 Agent 输出
			if event.Output != nil && event.Output.MessageOutput != nil {
				content := event.Output.MessageOutput.Message.Content
				if content != "" {
					fmt.Printf("\n%s", content)
				}
			}
		}
		fmt.Println()
	}
}

func testDecryption() {
	encryptedData := "eyJ2ZXJzaW9uIjoiYWVzLWNiYy12MSIsIml2IjoiMFhzQzZYSHdaSzQxL2pmVjNtZkY4UT09IiwiZW5jcnlwdGVkX2RhdGEiOiJNbGVWdEI3RXR6RDBUUDREZ1NzUVp0MitReGdTekhBTzNtUllmWGtKcFA1UzJGMmozdFJlc2UzN1NEVnRlWDBpIn0"
	decrypted, err := common.DecryptAPIKey(encryptedData)
	if err != nil {
		fmt.Printf("解密失败: %v\n", err)
		return
	}
	fmt.Printf("====== 解密测试 ======\n")
	fmt.Printf("加密字符串: %s\n", encryptedData)
	fmt.Printf("解密结果: %s\n", decrypted)
	fmt.Printf("======================\n")
}

// initApp 初始化应用（配置、数据库）
/*
func initApp() {
	// 加载 .env 文件
	if err := godotenv.Load("../.env"); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	// 加载配置文件
	configPath := findConfigFile()
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	cfg.ExpandEnv()

	// 初始化数据库
	err = repository.InitDatabase(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	log.Println("数据库初始化成功")

	// 初始化 Milvus
	ctx := context.Background()
	_, err = milvus.InitMilvusManager(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to initialize Milvus: %v", err)
	}
	log.Println("Milvus 初始化成功")
}

// findConfigFile 查找配置文件
func findConfigFile() string {
	// 尝试多个可能的路径
	paths := []string{
		"../config.yaml",
		"../../config.yaml",
		"config.yaml",
	}

	for _, p := range paths {
		absPath, _ := filepath.Abs(p)
		if _, err := os.Stat(absPath); err == nil {
			return absPath
		}
	}

	return "../config.yaml"
}
*/
