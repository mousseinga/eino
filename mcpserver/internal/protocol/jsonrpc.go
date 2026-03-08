package protocol

import (
	"encoding/json"
)

// Request JSON-RPC 2.0 请求结构
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Error JSON-RPC 2.0 错误结构
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data,omitempty"`
}

// ErrorResponse JSON-RPC 2.0 错误响应
type ErrorResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Error   Error       `json:"error"`
}
