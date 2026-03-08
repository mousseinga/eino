package protocol

import (
	"sync"
)

// SSESession SSE 会话
type SSESession struct {
	ID          string
	EventChan   chan string
	RequestChan chan *Request
	Done        chan struct{}
	ResponseMap sync.Map // map[string]chan interface{}
	mu          sync.RWMutex
}
