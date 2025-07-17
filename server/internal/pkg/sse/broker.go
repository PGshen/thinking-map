package sse

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/PGshen/thinking-map/server/internal/model/dto"
	"github.com/gin-gonic/gin"
)

// Event 表示一个SSE事件
type Event struct {
	ID    string      `json:"id"`
	Type  string      `json:"type"`
	Data  interface{} `json:"data"`
	Retry uint64      `json:"retry,omitempty"`
}

// Client 表示一个SSE客户端连接
type Client struct {
	ID        string
	EventChan chan Event
	Done      chan struct{}
}

// Broker 管理所有客户端连接和事件分发
type Broker struct {
	store         SessionStore
	mutex         sync.RWMutex
	pingInterval  time.Duration
	clientTimeout time.Duration
}

// NewBroker 创建一个新的事件代理
func NewBroker(store SessionStore, pingInterval, clientTimeout time.Duration) *Broker {
	return &Broker{
		store:         store,
		pingInterval:  pingInterval,
		clientTimeout: clientTimeout,
	}
}

// NewClient 创建一个新的客户端
func (b *Broker) NewClient(clientID, sessionID string) *Client {
	client := &Client{
		ID:        clientID,
		EventChan: make(chan Event, 100),
		Done:      make(chan struct{}),
	}
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.store.AddClient(clientID, client)
	b.store.AddClientToSession(sessionID, clientID)
	return client
}

// RemoveClient 移除客户端
func (b *Broker) RemoveClient(clientID, sessionID string) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.store.RemoveClientFromSession(sessionID, clientID)
	b.store.RemoveClient(clientID)
}

// Publish 向特定会话的所有客户端发送事件
func (b *Broker) Publish(sessionID string, event Event) {
	b.mutex.RLock()
	clients, err := b.store.GetSessionClients(sessionID)
	b.mutex.RUnlock()
	if err != nil {
		return
	}
	for clientID := range clients {
		b.sendToClient(clientID, event)
	}
}

// SendToAll 向所有客户端发送事件
func (b *Broker) SendToAll(event Event) {
	b.mutex.RLock()
	clients, err := b.store.GetSessionClients("")
	b.mutex.RUnlock()
	if err != nil {
		return
	}
	for clientID := range clients {
		b.sendToClient(clientID, event)
	}
}

// sendToClient 向特定客户端发送事件
func (b *Broker) sendToClient(clientID string, event Event) {
	b.mutex.RLock()
	client, err := b.store.GetClient(clientID)
	b.mutex.RUnlock()
	if err != nil {
		return
	}
	select {
	case client.EventChan <- event:
	default:
		log.Printf("客户端缓冲区已满，丢弃消息: %s", clientID)
	}
}

// HandleSSE 处理SSE请求（Gin专用）
func (b *Broker) HandleSSE(c *gin.Context, sessionID, clientID string) {
	client := b.NewClient(clientID, sessionID)
	defer b.RemoveClient(clientID, sessionID)

	// 设置SSE响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// 发送连接建立事件
	client.EventChan <- Event{
		Type: dto.ConnectionEstablishedEventType,
		Data: dto.ConnectionEstablishedEvent{
			SessionID: sessionID,
			ClientID:  clientID,
			Message:   "SSE连接已建立",
		},
	}

	// 启动心跳
	go b.ping(client)

	// 事件循环
	c.Stream(func(w io.Writer) bool {
		select {
		case event, ok := <-client.EventChan:
			if !ok {
				return false
			}
			// 序列化事件数据
			var data []byte
			var err error
			if str, ok := event.Data.(string); ok {
				data = []byte(str)
			} else {
				data, err = json.Marshal(event.Data)
				if err != nil {
					fmt.Printf("Error marshaling event data: %v\n", err)
					return true
				}
			}
			c.SSEvent(event.Type, string(data))
			return true
		case <-c.Request.Context().Done():
			close(client.Done)
			return false
		}
	})
}

// ping 发送心跳
func (b *Broker) ping(client *Client) {
	ticker := time.NewTicker(b.pingInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			client.EventChan <- Event{
				Type: "ping",
				Data: time.Now().Unix(),
			}
		case <-client.Done:
			return
		}
	}
}
