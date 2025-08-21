/*
 * 分布式事件总线接口和实现
 * 支持跨服务实例的SSE事件分发
 */
package sse

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/redis/go-redis/v9"
)

// LocalClientProvider 本地客户端提供者接口
type LocalClientProvider interface {
	// 获取本地客户端
	GetLocalClient(clientID string) *LocalClient
	// 获取会话中的本地客户端
	GetLocalSessionClients(sessionID string) []*LocalClient
}

// EventBus 事件总线接口
type EventBus interface {
	// 发布事件到指定客户端
	PublishToClient(ctx context.Context, clientID string, event Event) error
	// 发布事件到会话中的所有客户端
	PublishToSession(ctx context.Context, sessionID string, event Event) error
	// 订阅会话事件，返回处理器ID
	SubscribeSession(ctx context.Context, sessionID string, handler EventHandler) (string, error)
	// 取消订阅会话事件
	UnsubscribeSession(ctx context.Context, sessionID string) error
	// 移除会话中的特定处理器
	RemoveSessionHandler(ctx context.Context, sessionID string, handlerID string) error
	// 设置本地客户端提供者（用于性能优化）
	SetLocalClientProvider(provider LocalClientProvider)
	// 关闭事件总线
	Close() error
}

// EventHandler 事件处理器
type EventHandler func(event Event) error

// RedisEventBus Redis实现的分布式事件总线
type RedisEventBus struct {
	redis               *redis.Client
	subscribers         map[string]*redis.PubSub
	handlers            map[string]map[string]EventHandler // sessionKey -> handlerID -> handler
	mutex               sync.RWMutex
	ctx                 context.Context
	cancel              context.CancelFunc
	localClientProvider LocalClientProvider // 本地客户端提供者，用于性能优化
	connManager         ConnectionManager   // 连接管理器，用于判断客户端位置
	serverID            string              // 当前服务器ID
}

// NewRedisEventBus 创建Redis事件总线
func NewRedisEventBus(redisClient *redis.Client, connManager ConnectionManager, serverID string) *RedisEventBus {
	ctx, cancel := context.WithCancel(context.Background())
	return &RedisEventBus{
		redis:       redisClient,
		subscribers: make(map[string]*redis.PubSub),
		handlers:    make(map[string]map[string]EventHandler),
		ctx:         ctx,
		cancel:      cancel,
		connManager: connManager,
		serverID:    serverID,
	}
}

// SetLocalClientProvider 设置本地客户端提供者
func (bus *RedisEventBus) SetLocalClientProvider(provider LocalClientProvider) {
	bus.mutex.Lock()
	defer bus.mutex.Unlock()
	bus.localClientProvider = provider
}

// PublishToClient 发布事件到指定客户端
func (bus *RedisEventBus) PublishToClient(ctx context.Context, clientID string, event Event) error {
	bus.mutex.RLock()
	localProvider := bus.localClientProvider
	bus.mutex.RUnlock()

	// 优先尝试本地直接投递
	if localProvider != nil {
		if localClient := localProvider.GetLocalClient(clientID); localClient != nil {
			// 直接发送到本地客户端的EventChan
			select {
			case localClient.EventChan <- event:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			default:
				// EventChan满了，记录警告但继续使用Redis作为备选
				log.Printf("本地客户端 %s 的EventChan已满，回退到Redis分发", clientID)
			}
			return nil
		}
	}

	// 检查客户端是否在本地服务器上
	if bus.connManager != nil {
		if conn, err := bus.connManager.GetConnection(ctx, clientID); err == nil && conn != nil {
			if conn.ServerID == bus.serverID {
				// 客户端在本地但LocalClient不可用，可能是连接刚建立或已断开
				log.Printf("客户端 %s 在本地服务器但LocalClient不可用，使用Redis分发", clientID)
			}
		}
	}

	// 使用Redis pub/sub分发（跨服务器或本地备选）
	channel := fmt.Sprintf("sse:client:%s", clientID)
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event failed: %w", err)
	}
	return bus.redis.Publish(ctx, channel, data).Err()
}

// PublishToSession 发布事件到会话中的所有客户端
func (bus *RedisEventBus) PublishToSession(ctx context.Context, sessionID string, event Event) error {
	bus.mutex.RLock()
	localProvider := bus.localClientProvider
	bus.mutex.RUnlock()

	// 优先向本地会话客户端直接投递
	localDelivered := 0
	if localProvider != nil {
		localClients := localProvider.GetLocalSessionClients(sessionID)
		if len(localClients) == 0 {
			log.Printf("会话 %s 没有本地客户端", sessionID)
		}
		for _, localClient := range localClients {
			select {
			case localClient.EventChan <- event:
				localDelivered++
			case <-ctx.Done():
				return ctx.Err()
			default:
				// EventChan满了，记录警告但继续
				log.Printf("本地客户端 %s 的EventChan已满，跳过本地投递", localClient.ID)
			}
		}
		return nil
	}

	// 使用Redis pub/sub确保跨服务器分发
	// 注意：本地客户端也会收到Redis消息，但由于已经直接投递，可以在handler中去重
	channel := fmt.Sprintf("sse:session:%s", sessionID)
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event failed: %w", err)
	}
	return bus.redis.Publish(ctx, channel, data).Err()
}

// SubscribeClient 订阅客户端事件
func (bus *RedisEventBus) SubscribeClient(ctx context.Context, clientID string, handler EventHandler) error {
	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	channel := fmt.Sprintf("sse:client:%s", clientID)

	// 如果已经订阅，先取消订阅
	if pubsub, exists := bus.subscribers[clientID]; exists {
		pubsub.Close()
		delete(bus.subscribers, clientID)
		delete(bus.handlers, clientID)
	}

	// 创建新的订阅
	pubsub := bus.redis.Subscribe(ctx, channel)
	bus.subscribers[clientID] = pubsub
	bus.handlers[clientID] = map[string]EventHandler{clientID: handler}

	// 启动消息处理协程
	go bus.handleMessages(clientID, pubsub)

	return nil
}

// SubscribeSession 订阅会话事件
func (bus *RedisEventBus) SubscribeSession(ctx context.Context, sessionID string, handler EventHandler) (string, error) {
	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	channel := fmt.Sprintf("sse:session:%s", sessionID)
	subscriptionKey := "session:" + sessionID

	// 生成处理器ID并添加处理器
	handlerID := fmt.Sprintf("%s-%d", sessionID, len(bus.handlers[subscriptionKey]))
	// 如果还没有订阅，创建新的订阅
	if _, exists := bus.subscribers[subscriptionKey]; !exists {
		pubsub := bus.redis.Subscribe(ctx, channel)
		bus.subscribers[subscriptionKey] = pubsub
		bus.handlers[subscriptionKey] = make(map[string]EventHandler)
		bus.handlers[subscriptionKey][handlerID] = handler
		// 启动消息处理协程
		go bus.handleMessages(subscriptionKey, pubsub)
	} else {
		bus.handlers[subscriptionKey][handlerID] = handler
	}

	return handlerID, nil
}

// UnsubscribeSession 取消订阅会话事件
func (bus *RedisEventBus) UnsubscribeSession(ctx context.Context, sessionID string) error {
	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	subscriptionKey := "session:" + sessionID
	if pubsub, exists := bus.subscribers[subscriptionKey]; exists {
		pubsub.Close()
		delete(bus.subscribers, subscriptionKey)
		delete(bus.handlers, subscriptionKey)
	}

	return nil
}

// RemoveSessionHandler 移除会话中的特定处理器
func (bus *RedisEventBus) RemoveSessionHandler(ctx context.Context, sessionID string, handlerID string) error {
	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	subscriptionKey := "session:" + sessionID
	if handlers, exists := bus.handlers[subscriptionKey]; exists {
		delete(handlers, handlerID)
		// 如果没有处理器了，取消订阅
		if len(handlers) == 0 {
			if pubsub, exists := bus.subscribers[subscriptionKey]; exists {
				pubsub.Close()
				delete(bus.subscribers, subscriptionKey)
				delete(bus.handlers, subscriptionKey)
			}
		}
	}

	return nil
}

// handleMessages 处理Redis订阅消息
func (bus *RedisEventBus) handleMessages(subscriptionKey string, pubsub *redis.PubSub) {
	ch := pubsub.Channel()
	for {
		select {
		case msg := <-ch:
			if msg == nil {
				return
			}

			var event Event
			if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
				log.Printf("解析事件失败: %v", err)
				continue
			}

			bus.mutex.RLock()
			handlers, exists := bus.handlers[subscriptionKey]
			bus.mutex.RUnlock()

			if exists {
				for _, handler := range handlers {
					if handler != nil {
						if err := handler(event); err != nil {
							log.Printf("处理事件失败: %v", err)
						}
					}
				}
			}
		case <-bus.ctx.Done():
			return
		}
	}
}

// Close 关闭事件总线
func (bus *RedisEventBus) Close() error {
	bus.cancel()

	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	for _, pubsub := range bus.subscribers {
		pubsub.Close()
	}

	bus.subscribers = make(map[string]*redis.PubSub)
	bus.handlers = make(map[string]map[string]EventHandler)

	return nil
}
