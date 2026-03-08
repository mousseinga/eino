package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config MCP 服务器配置
type Config struct {
	// Server 服务器配置
	Server ServerConfig `yaml:"server"`

	// APIKeys API 密钥配置
	APIKeys APIKeysConfig `yaml:"api_keys"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	// Host 服务器监听地址（IP 或域名）
	Host string `yaml:"host"`

	// Port 服务器监听端口
	Port int `yaml:"port"`

	// ServerName 服务器名称
	ServerName string `yaml:"server_name"`

	// ServerVersion 服务器版本
	ServerVersion string `yaml:"server_version"`

	// EnableCORS 是否启用 CORS
	EnableCORS bool `yaml:"enable_cors"`

	// AllowedOrigins CORS 允许的来源
	AllowedOrigins []string `yaml:"allowed_origins"`
}

// APIKeysConfig API 密钥配置
type APIKeysConfig struct {
	// WeatherAPIKey 天气 API 密钥（OpenWeatherMap）
	WeatherAPIKey string `yaml:"weather_api_key"`

	// OpenAIAPIKey OpenAI API 密钥
	OpenAIAPIKey string `yaml:"openai_api_key"`

	// GoogleAPIKey Google API 密钥
	GoogleAPIKey string `yaml:"google_api_key"`

	// 其他 API 密钥可以通过 AdditionalKeys 添加
	AdditionalKeys map[string]string `yaml:"additional_keys"`
}

// LoadConfig 从文件加载配置
func LoadConfig(configPath string) (*Config, error) {
	// 如果配置文件路径为空，尝试查找默认配置文件
	if configPath == "" {
		configPath = findConfigFile()
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	// 解析 YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &config, nil
}

// findConfigFile 查找配置文件
func findConfigFile() string {
	// 可能的配置文件路径
	possiblePaths := []string{
		"config.yaml",
		"config.yml",
		"./config.yaml",
		"./config.yml",
		"../config.yaml",
		"../config.yml",
	}

	// 获取当前工作目录
	wd, _ := os.Getwd()

	// 获取可执行文件所在目录
	execPath, _ := os.Executable()
	execDir := filepath.Dir(execPath)

	// 添加到搜索路径
	for _, path := range possiblePaths {
		// 检查当前目录
		if _, err := os.Stat(path); err == nil {
			return path
		}

		// 检查工作目录
		fullPath := filepath.Join(wd, path)
		if _, err := os.Stat(fullPath); err == nil {
			return fullPath
		}

		// 检查可执行文件目录
		fullPath = filepath.Join(execDir, path)
		if _, err := os.Stat(fullPath); err == nil {
			return fullPath
		}
	}

	// 如果找不到，返回默认路径
	return "config.yaml"
}

// Validate 验证配置
func (c *Config) Validate() error {
	// 验证服务器配置
	if c.Server.Host == "" {
		c.Server.Host = "0.0.0.0"
	}
	if c.Server.Port <= 0 {
		c.Server.Port = 8080
	}
	if c.Server.ServerName == "" {
		c.Server.ServerName = "mcp-tool-server"
	}
	if c.Server.ServerVersion == "" {
		c.Server.ServerVersion = "1.0.0"
	}
	if len(c.Server.AllowedOrigins) == 0 {
		c.Server.AllowedOrigins = []string{"*"}
	}

	return nil
}

// GetAddress 获取服务器地址（host:port 格式）
func (c *Config) GetAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// GetServerAddress 获取服务器地址（host:port 格式）
func GetAddress() string {
	// 这个方法用于从默认配置获取地址
	// 实际使用时应该通过 Config 实例调用
	return "0.0.0.0:8080"
}

// GetAPIKey 获取 API 密钥
func (c *Config) GetAPIKey(keyName string) string {
	switch strings.ToLower(keyName) {
	case "weather", "weather_api_key":
		return c.APIKeys.WeatherAPIKey
	case "google", "google_api_key":
		return c.APIKeys.GoogleAPIKey
	default:
		if c.APIKeys.AdditionalKeys != nil {
			if key, ok := c.APIKeys.AdditionalKeys[keyName]; ok {
				return key
			}
		}
		return ""
	}
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host:          "0.0.0.0",
			Port:          8080,
			ServerName:    "mcp-tool-server",
			ServerVersion: "1.0.0",
			EnableCORS:    true,
			AllowedOrigins: []string{
				"*",
			},
		},
		APIKeys: APIKeysConfig{
			AdditionalKeys: make(map[string]string),
		},
	}
}
