package config

import "fmt"

// ServerConfigFromFile 从配置文件创建服务器配置
func ServerConfigFromFile(cfg *Config) *ServerConfig {
	return &ServerConfig{
		Host:           cfg.Server.Host,
		Port:           cfg.Server.Port,
		ServerName:     cfg.Server.ServerName,
		ServerVersion:  cfg.Server.ServerVersion,
		EnableCORS:     cfg.Server.EnableCORS,
		AllowedOrigins: cfg.Server.AllowedOrigins,
	}
}

// GetAddress 获取服务器地址（host:port 格式）
func (c *ServerConfig) GetAddress() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
