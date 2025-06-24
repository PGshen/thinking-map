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
		Type: "connection-established",
		Data: map[string]string{
			"session_id": sessionID,
			"client_id":  clientID,
			"message":    "SSE连接已建立",
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

// EventManager manages SSE connections and events
type EventManager struct {
	clients map[string]map[string]chan []byte // mapID -> connectionID -> channel
	mu      sync.RWMutex
	done    chan struct{}
}

// NewEventManager creates a new EventManager
func NewEventManager() *EventManager {
	return &EventManager{
		clients: make(map[string]map[string]chan []byte),
		done:    make(chan struct{}),
	}
}

// Connect creates a new SSE connection for a map
func (m *EventManager) Connect(mapID string) (string, chan []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()

	connectionID := uuid.New().String()
	eventChan := make(chan []byte, 100)

	if _, exists := m.clients[mapID]; !exists {
		m.clients[mapID] = make(map[string]chan []byte)
	}
	m.clients[mapID][connectionID] = eventChan

	// Send connection event
	event := dto.SSEConnectionResponse{
		ConnectionID: connectionID,
		MapID:        mapID,
		Timestamp:    time.Now(),
	}
	data, _ := json.Marshal(event)
	eventChan <- []byte(fmt.Sprintf("event: connected\ndata: %s\n\n", data))

	return connectionID, eventChan
}

// Disconnect removes an SSE connection
func (m *EventManager) Disconnect(mapID, connectionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if clients, exists := m.clients[mapID]; exists {
		if ch, ok := clients[connectionID]; ok {
			// Send disconnection event
			event := dto.SSEDisconnectionResponse{
				ConnectionID: connectionID,
				Reason:       "user_disconnected",
				Timestamp:    time.Now(),
			}
			data, _ := json.Marshal(event)
			ch <- []byte(fmt.Sprintf("event: disconnected\ndata: %s\n\n", data))

			close(ch)
			delete(clients, connectionID)
		}
		if len(clients) == 0 {
			delete(m.clients, mapID)
		}
	}
}

// BroadcastEvent sends an event to all clients of a map
func (m *EventManager) BroadcastEvent(mapID, eventType string, data interface{}) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	clients, exists := m.clients[mapID]
	if !exists {
		return fmt.Errorf("no clients for map %s", mapID)
	}

	eventData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	message := fmt.Sprintf("event: %s\ndata: %s\n\n", eventType, eventData)
	for _, ch := range clients {
		select {
		case ch <- []byte(message):
		default:
			// Channel is full, skip this client
		}
	}

	return nil
}

// Close closes the event manager and all connections
func (m *EventManager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()

	close(m.done)
	for _, clients := range m.clients {
		for _, ch := range clients {
			close(ch)
		}
	}
	m.clients = make(map[string]map[string]chan []byte)
}
