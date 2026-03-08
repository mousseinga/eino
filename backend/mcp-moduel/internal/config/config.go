package config

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/bytedance/sonic"
)

// TransportType 传输类型
type TransportType string

const (
	// TransportTypeHTTP HTTP 传输
	TransportTypeHTTP TransportType = "http"
	// TransportTypeSSE SSE (Server-Sent Events) 传输
	TransportTypeSSE TransportType = "sse"
	// TransportTypeStdio stdio (标准输入输出) 传输
	TransportTypeStdio TransportType = "stdio"
)

// Config MCP 客户端配置结构
type Config struct {
	// ServerURL MCP 服务器的完整 URL 地址（HTTP/SSE 传输时必需）
	ServerURL string `json:"server_url,omitempty"`

	// TransportType 传输类型：http, sse, stdio
	TransportType TransportType `json:"transport_type,omitempty"`

	// Command stdio 传输时使用的命令（stdio 传输时必需）
	Command string `json:"command,omitempty"`

	// CommandArgs stdio 传输时使用的命令参数
	CommandArgs []string `json:"command_args,omitempty"`

	// CommandEnv stdio 传输时使用的环境变量
	CommandEnv []string `json:"command_env,omitempty"`

	APIKey string `json:"api_key,omitempty"`

	Timeout int `json:"timeout,omitempty"`

	RetryTimes int `json:"retry_times,omitempty"`

	AdditionalConfig map[string]interface{} `json:"additional_config,omitempty"`
}

// Validate 验证配置的有效性
func (c *Config) Validate() error {
	// 如果没有指定传输类型，默认为 HTTP
	if c.TransportType == "" {
		c.TransportType = TransportTypeHTTP
	}

	// 验证传输类型
	if c.TransportType != TransportTypeHTTP && c.TransportType != TransportTypeSSE && c.TransportType != TransportTypeStdio {
		return ErrInvalidConfig{Field: "transport_type", Reason: fmt.Sprintf("transport_type must be one of: %s, %s, %s", TransportTypeHTTP, TransportTypeSSE, TransportTypeStdio)}
	}

	// 对于 HTTP 和 SSE 传输，需要 ServerURL
	if c.TransportType == TransportTypeHTTP || c.TransportType == TransportTypeSSE {
		if c.ServerURL == "" {
			return ErrInvalidConfig{Field: "server_url", Reason: "server_url is required for http and sse transport"}
		}

		if err := validateURL(c.ServerURL); err != nil {
			return ErrInvalidConfig{Field: "server_url", Reason: err.Error()}
		}
	}

	// 对于 stdio 传输，需要 Command
	if c.TransportType == TransportTypeStdio {
		if c.Command == "" {
			return ErrInvalidConfig{Field: "command", Reason: "command is required for stdio transport"}
		}
	}

	if c.Timeout < 0 {
		return ErrInvalidConfig{Field: "timeout", Reason: "timeout must be non-negative"}
	}

	if c.RetryTimes < 0 {
		return ErrInvalidConfig{Field: "retry_times", Reason: "retry_times must be non-negative"}
	}

	return nil
}

// validateURL 验证 URL 格式的有效性
func validateURL(urlStr string) error {
	urlStr = strings.TrimSpace(urlStr)
	if urlStr == "" {
		return fmt.Errorf("server_url cannot be empty")
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	if parsedURL.Scheme == "" {
		return fmt.Errorf("URL must include a scheme (http:// or https://)")
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("URL scheme must be http or https, got: %s", parsedURL.Scheme)
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("URL must include a host")
	}

	return nil
}

// SetDefaults 设置配置的默认值
func (c *Config) SetDefaults() {
	// 如果超时时间未设置或无效，设置为 30 秒
	if c.Timeout <= 0 {
		c.Timeout = 30
	}
	// 如果重试次数未设置，设置为 0（不重试）
	if c.RetryTimes < 0 {
		c.RetryTimes = 0
	}
}

// ErrInvalidConfig 配置验证错误类型
type ErrInvalidConfig struct {
	Field string

	Reason string
}

// Error 实现 error 接口
func (e ErrInvalidConfig) Error() string {
	return "invalid mcp config: " + e.Field + " - " + e.Reason
}

// Parse 解析 MCP 配置 JSON 字符串
func Parse(configJSON string) (*Config, error) {
	if configJSON == "" {
		return nil, fmt.Errorf("mcp config is required")
	}

	var cfg Config
	if err := sonic.UnmarshalString(configJSON, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse mcp config json: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid mcp config: %w", err)
	}

	cfg.SetDefaults()
	return &cfg, nil
}

// Marshal 将配置序列化为 JSON 字符串
func Marshal(cfg *Config) (string, error) {
	if cfg == nil {
		return "", fmt.Errorf("config cannot be nil")
	}
	return sonic.MarshalString(cfg)
}
