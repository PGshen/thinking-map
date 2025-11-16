package sse

import (
	"context"
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

// Client 表示一个SSE客户端元数据
type Client struct {
	ID        string `json:"id"`
	SessionID string `json:"session_id"`
	UserID    string `json:"user_id,omitempty"`
	CreatedAt int64  `json:"created_at"`
}

// LocalClient 本地客户端连接
type LocalClient struct {
	*Client
	EventChan    chan Event
	Done         chan bool
	LastActiveAt int64  // 最后活跃时间
	HandlerID    string // 事件处理器ID
}

// Broker 管理所有客户端连接和事件分发
type Broker struct {
	eventBus      EventBus
	connManager   ConnectionManager
	mutex         sync.RWMutex
	pingInterval  time.Duration
	clientTimeout time.Duration
	serverID      string
	localClients  map[string]*LocalClient // 本地活跃连接
}

// NewBroker 创建一个新的事件代理
func NewBroker(eventBus EventBus, connManager ConnectionManager, serverID string, pingInterval, clientTimeout time.Duration) *Broker {
	b := &Broker{
		eventBus:      eventBus,
		connManager:   connManager,
		serverID:      serverID,
		pingInterval:  pingInterval,
		clientTimeout: clientTimeout,
		localClients:  make(map[string]*LocalClient),
	}

	// 设置本地客户端提供者，用于性能优化
	eventBus.SetLocalClientProvider(b)

	// 启动连接状态监控
	go b.startConnectionMonitor()

	return b
}

// NewClient 创建一个新的客户端
func (b *Broker) NewClient(clientID, sessionID string) *LocalClient {
	now := time.Now().Unix()
	clientMeta := &Client{
		ID:        clientID,
		SessionID: sessionID,
		CreatedAt: now,
	}
	localClient := &LocalClient{
		Client:       clientMeta,
		EventChan:    make(chan Event, 10240),
		Done:         make(chan bool),
		LastActiveAt: now,
	}
	b.mutex.Lock()
	defer b.mutex.Unlock()

	// 检查是否已存在相同clientID的连接，如果存在则先移除旧连接
	if existingClient, exists := b.localClients[clientID]; exists {
		log.Printf("发现已存在的客户端连接，移除旧连接: %s", clientID)
		// 关闭旧连接的事件通道
		close(existingClient.EventChan)
		// 移除旧的事件处理器
		if existingClient.HandlerID != "" {
			if err := b.eventBus.RemoveSessionHandler(context.Background(), sessionID, existingClient.HandlerID); err != nil {
				log.Printf("移除旧会话事件处理器失败: %v", err)
			}
		}
	}

	// 注册连接到ConnectionManager
	conn := &ClientConnection{
		ID:        clientID,
		SessionID: sessionID,
	}
	if err := b.connManager.RegisterConnection(context.Background(), conn); err != nil {
		log.Printf("注册连接失败: %v", err)
		return nil
	} else {
		log.Printf("注册连接成功: %s - %s", sessionID, clientID)
	}

	// 添加到本地客户端映射
	b.localClients[clientID] = localClient

	return localClient
}

// RemoveClient 移除客户端
func (b *Broker) RemoveClient(clientID, sessionID string) {
	b.RemoveClientWithTimestamp(clientID, sessionID, 0)
}

// RemoveClientWithTimestamp 移除指定时间戳的客户端，避免并发问题
func (b *Broker) RemoveClientWithTimestamp(clientID, sessionID string, createdAt int64) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	// 检查客户端是否存在
	client, exists := b.localClients[clientID]
	if !exists {
		log.Printf("客户端不存在，无需移除: %s", clientID)
		return
	}

	// 如果指定了时间戳，检查是否匹配，避免移除新建立的连接
	if createdAt > 0 && client.CreatedAt != createdAt {
		log.Printf("客户端时间戳不匹配，跳过移除: %s (期望: %d, 实际: %d)", clientID, createdAt, client.CreatedAt)
		return
	}

	// 注销连接
	if err := b.connManager.UnregisterConnection(context.Background(), clientID); err != nil {
		log.Printf("注销连接失败: %v", err)
	}

	// 移除会话事件处理器
	if client.HandlerID != "" {
		if err := b.eventBus.RemoveSessionHandler(context.Background(), sessionID, client.HandlerID); err != nil {
			log.Printf("移除会话事件处理器失败: %v", err)
		}
	}

	// 从本地客户端映射中移除
	delete(b.localClients, clientID)

	log.Printf("移除客户端: %s, 会话: %s, 时间戳: %d", clientID, sessionID, client.CreatedAt)
}

// GetLocalClient 获取本地客户端（实现LocalClientProvider接口）
func (b *Broker) GetLocalClient(clientID string) *LocalClient {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return b.localClients[clientID]
}

// GetLocalSessionClients 获取会话中的本地客户端（实现LocalClientProvider接口）
func (b *Broker) GetLocalSessionClients(sessionID string) []*LocalClient {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	var clients []*LocalClient
	for _, client := range b.localClients {
		if client.SessionID == sessionID {
			clients = append(clients, client)
		}
	}
	return clients
}

// GetClients 获取会话中的所有客户端元数据
func (b *Broker) GetClients(sessionID string) []*Client {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	connections, err := b.connManager.GetSessionConnections(context.Background(), sessionID)
	if err != nil {
		return nil
	}
	var clients []*Client
	for _, conn := range connections {
		client := &Client{
			ID:        conn.ID,
			SessionID: conn.SessionID,
			CreatedAt: conn.CreatedAt.Unix(),
		}
		clients = append(clients, client)
	}
	return clients
}

// PublishToSession 向会话发布事件（使用事件总线）
func (b *Broker) PublishToSession(sessionID string, event Event) error {
	return b.eventBus.PublishToSession(context.Background(), sessionID, event)
}

// PublishToClient 向特定客户端发布事件（使用事件总线）
func (b *Broker) PublishToClient(clientID string, event Event) error {
	return b.eventBus.PublishToClient(context.Background(), clientID, event)
}

// HandleSSE 处理SSE请求（Gin专用）
func (b *Broker) HandleSSE(c *gin.Context, sessionID, clientID string) {
	client := b.NewClient(clientID, sessionID)
	if client == nil || client.EventChan == nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Code:      http.StatusInternalServerError,
			Message:   "创建客户端失败",
			Data:      nil,
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}
	// 使用带时间戳的移除方法，确保只移除当前连接创建的客户端
	defer b.RemoveClientWithTimestamp(clientID, sessionID, client.CreatedAt)

	// 设置SSE响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// 订阅会话事件（如果还没有订阅）
	handlerID, err := b.eventBus.SubscribeSession(c.Request.Context(), sessionID, b.createEventHandler(client))
	if err != nil {
		log.Printf("订阅会话事件失败: %v", err)
	} else {
		client.HandlerID = handlerID
	}

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
				// fmt.Printf("Event data: %s\n", str)
				data = []byte(str)
			} else {
				data, err = json.Marshal(event.Data)
				if err != nil {
					fmt.Printf("Error marshaling event data: %v\n", err)
					return true
				}
				// fmt.Printf("Event data: %s\n", string(data))
			}
			c.SSEvent(event.Type, string(data))
			return true
		case <-c.Request.Context().Done():
			close(client.Done)
			return false
		}
	})
}

// createEventHandler 创建事件处理器
func (b *Broker) createEventHandler(localClient *LocalClient) EventHandler {
	return func(event Event) error {
		// 更新客户端活跃时间
		b.updateClientActivity(localClient.ID)

		select {
		case localClient.EventChan <- event:
			return nil
		default:
			log.Printf("客户端缓冲区已满，丢弃消息: %s", localClient.ID)
			return fmt.Errorf("client buffer full")
		}
	}
}

// startConnectionMonitor 启动连接状态监控
func (b *Broker) startConnectionMonitor() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("startConnectionMonitor recovered from panic: %v", r)
		}
	}()
	ticker := time.NewTicker(b.clientTimeout / 2) // 每半个超时时间检查一次
	defer ticker.Stop()

	for range ticker.C {
		b.cleanupExpiredConnections()
	}
}

// cleanupExpiredConnections 清理过期连接
func (b *Broker) cleanupExpiredConnections() {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	now := time.Now().Unix()
	for clientID, client := range b.localClients {
		// 检查连接是否超时（使用最后活跃时间）
		if now-client.LastActiveAt > int64(b.clientTimeout.Seconds()) {
			log.Printf("清理超时连接: %s (最后活跃: %d秒前)", clientID, now-client.LastActiveAt)
			// 关闭客户端
			close(client.Done)
			close(client.EventChan)
			// 从映射中移除
			delete(b.localClients, clientID)
			// 注销连接
			b.connManager.UnregisterConnection(context.Background(), clientID)
		}
	}
}

// updateClientActivity 更新客户端活跃时间
func (b *Broker) updateClientActivity(clientID string) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if client, exists := b.localClients[clientID]; exists {
		client.LastActiveAt = time.Now().Unix()
	}
}

// ping 发送心跳
func (b *Broker) ping(client *LocalClient) {
	ticker := time.NewTicker(b.pingInterval)
	defer func() {
		ticker.Stop()
		if err := recover(); err != nil {
			log.Printf("ping goroutine recovered from panic: %v", err)
		}
	}()
	for {
		select {
		case <-ticker.C:
			// 更新客户端活跃时间
			b.updateClientActivity(client.ID)
			client.EventChan <- Event{
				Type: "ping",
				Data: time.Now().Unix(),
			}
		case <-client.Done:
			return
		}
	}
}
