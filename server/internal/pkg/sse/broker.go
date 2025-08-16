package sse

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/PGshen/thinking-map/server/internal/model/dto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	if err := b.store.AddClient(clientID, client); err != nil {
		log.Printf("添加客户端失败: %v", err)
		return nil
	}
	if err := b.store.AddClientToSession(sessionID, clientID); err != nil {
		log.Printf("添加客户端到会话失败: %v", err)
		return nil
	}
	return client
}

// RemoveClient 移除客户端
func (b *Broker) RemoveClient(clientID, sessionID string) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	if err := b.store.RemoveClientFromSession(sessionID, clientID); err != nil {
		log.Printf("从会话移除客户端失败: %v", err)
	}
	if err := b.store.RemoveClient(clientID); err != nil {
		log.Printf("移除客户端失败: %v", err)
	}
	log.Printf("移除客户端: %s, 会话: %s", clientID, sessionID)
}

// GetClientIDs 获取会话中的所有客户端ID
func (b *Broker) GetClients(sesstionID string) []*Client {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	clientIDs, err := b.store.GetSessionClients(sesstionID)
	if err != nil {
		return nil
	}
	var clients []*Client
	for clientID := range clientIDs {
		client, err := b.store.GetClient(clientID)
		if err != nil {
			continue
		}
		clients = append(clients, client)
	}
	return clients
}

// PublishToClients 向多个客户端发送事件
func (b *Broker) PublishToClients(clients []*Client, event Event) {
	for _, client := range clients {
		b.sendToClient(client, event)
	}
}

// Publish 向特定会话的所有客户端发送事件
// @deprecated 高速频繁调用时存在性能问题, 请使用GetClients获取客户端后，再使用PublishToClients
func (b *Broker) Publish(sessionID string, event Event) {
	b.mutex.RLock()
	clients, err := b.store.GetSessionClients(sessionID)
	b.mutex.RUnlock()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	for clientID := range clients {
		client, err := b.store.GetClient(clientID)
		if err != nil {
			continue
		}
		b.sendToClient(client, event)
	}
}

// sendToClient 向特定客户端发送事件
func (b *Broker) sendToClient(client *Client, event Event) {
	select {
	case client.EventChan <- event:
	default:
		log.Printf("客户端缓冲区已满，丢弃消息: %s", client.ID)
	}
}

// HandleSSE 处理SSE请求（Gin专用）
func (b *Broker) HandleSSE(c *gin.Context, sessionID, clientID string) {
	client := b.NewClient(clientID, sessionID)
	if client == nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Code:      http.StatusInternalServerError,
			Message:   "创建客户端失败",
			Data:      nil,
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}
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
