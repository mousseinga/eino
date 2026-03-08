// Package config tool/config.go（配置结构体定义）
package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

// 1. 根配置结构体：对应 config.yaml 整个文件
// 类比 Spring Boot 中的配置类（用 @Configuration 注解的类）
type Config struct {
	Google GoogleConfig `yaml:"google"` // 对应 YAML 中的 google 节点
	OpenAI OpenAIConfig `yaml:"openai"` // 对应 YAML 中的 openai 节点（可选）
}

// 2. 谷歌配置结构体：对应 YAML 中的 google 节点
type GoogleConfig struct {
	APIKey         string `yaml:"api_key"`          // 对应 google.api_key
	SearchEngineID string `yaml:"search_engine_id"` // 对应 google.search_engine_id
}

// 3. OpenAI 配置结构体（可选，对应 YAML 中的 openai 节点）
type OpenAIConfig struct {
	APIKey  string `yaml:"api_key"`
	Model   string `yaml:"model_name"`
	BaseURL string `yaml:"base_url"`
}

// LoadConfig：读取 config.yaml 文件，返回配置结构体
// 类比 Spring Boot 的配置自动加载（手动实现，但能复用）
func LoadConfig() (*Config, error) {
	// 关键修改1：获取 config.go 源代码文件的真实路径（不受执行目录影响）
	_, currentSourceFile, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("获取源代码文件路径失败")
	}
	// currentSourceFile 格式：D:\Bear\ai-eino-agent\chatApp\tool\config.go（真实源代码路径）
	log.Printf("✅ 源代码文件路径：%s", currentSourceFile)

	// 关键修改2：向上追溯到项目根目录（按你的目录结构调整）
	// 目录结构：config.go → chatApp/tool → chatApp → 项目根目录（D:\Bear\ai-eino-agent\）
	toolDir := filepath.Dir(currentSourceFile) // 得到：chatApp/tool
	chatAppDir := filepath.Dir(toolDir)        // 得到：chatApp
	projectRootDir := filepath.Dir(chatAppDir) // 得到：项目根目录（D:\Bear\ai-eino-agent\）

	// 关键修改3：拼接项目根目录 + config.yaml（最终路径100%正确）
	configPath := filepath.Join(projectRootDir, "config.yaml")
	log.Printf("✅ 最终配置文件路径：%s", configPath)

	// 下面的代码不变！
	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("打开配置文件失败（路径：%s）：%v", configPath, err)
	}
	defer file.Close()

	var config Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("解析 YAML 失败：%v", err)
	}

	if config.Google.APIKey == "" || config.Google.SearchEngineID == "" {
		return nil, fmt.Errorf("config.yaml 中 google.api_key 或 google.search_engine_id 未填写")
	}

	return &config, nil
}
