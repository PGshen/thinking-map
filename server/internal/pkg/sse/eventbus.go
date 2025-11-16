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
	PublishToClient(ctx context.Context, clientID string, event Event) error
	PublishToSession(ctx context.Context, sessionID string, event Event) error
	SubscribeClient(ctx context.Context, clientID string) error
	SubscribeSession(ctx context.Context, sessionID string) error
	UnsubscribeClient(ctx context.Context, clientID string) error
	UnsubscribeSessionIfNoLocalClients(ctx context.Context, sessionID string) error
	SetLocalClientProvider(provider LocalClientProvider)
	Close() error
}

// RedisEventBus Redis实现的分布式事件总线
type RedisEventBus struct {
	redis               *redis.Client
	subscribers         map[string]*redis.PubSub
	mutex               sync.RWMutex
	ctx                 context.Context
	cancel              context.CancelFunc
	localClientProvider LocalClientProvider // 本地客户端提供者，用于性能优化
	connManager         ConnectionManager   // 连接管理器，用于判断客户端位置
	serverID            string              // 当前服务器ID
	sessionRemote       map[string]bool
	sessionServers      map[string]map[string]struct{}
}

// NewRedisEventBus 创建Redis事件总线
func NewRedisEventBus(redisClient *redis.Client, connManager ConnectionManager, serverID string) *RedisEventBus {
	ctx, cancel := context.WithCancel(context.Background())
	return &RedisEventBus{
		redis:          redisClient,
		subscribers:    make(map[string]*redis.PubSub),
		ctx:            ctx,
		cancel:         cancel,
		connManager:    connManager,
		serverID:       serverID,
		sessionRemote:  make(map[string]bool),
		sessionServers: make(map[string]map[string]struct{}),
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
			sent, err := bus.tryLocalSend(ctx, localClient, event)
			if err != nil {
				return err
			}
			if sent {
				return nil
			}
			log.Printf("本地客户端 %s 的EventChan不可用或已满，回退到Redis分发", clientID)
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

func (bus *RedisEventBus) tryLocalSend(ctx context.Context, lc *LocalClient, event Event) (bool, error) {
	defer func() {
		if r := recover(); r != nil {
		}
	}()
	select {
	case lc.EventChan <- event:
		return true, nil
	case <-ctx.Done():
		return false, ctx.Err()
	default:
		return false, nil
	}
}

// PublishToSession 发布事件到会话中的所有客户端
func (bus *RedisEventBus) PublishToSession(ctx context.Context, sessionID string, event Event) error {
	bus.mutex.RLock()
	localProvider := bus.localClientProvider
	hasRemote := bus.sessionRemote[sessionID]
	bus.mutex.RUnlock()

	if localProvider != nil {
		localClients := localProvider.GetLocalSessionClients(sessionID)
		for _, lc := range localClients {
			bus.tryLocalSend(ctx, lc, event)
		}
	}

	if !hasRemote {
		return nil
	}

	event.OriginServerID = bus.serverID
	channel := fmt.Sprintf("sse:session:%s", sessionID)
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event failed: %w", err)
	}
	return bus.redis.Publish(ctx, channel, data).Err()
}

// SubscribeClient 订阅客户端事件
func (bus *RedisEventBus) SubscribeClient(ctx context.Context, clientID string) error {
	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	subscriptionKey := "client:" + clientID
	channel := fmt.Sprintf("sse:client:%s", clientID)

	if pubsub, exists := bus.subscribers[subscriptionKey]; exists {
		pubsub.Close()
		delete(bus.subscribers, subscriptionKey)
	}

	pubsub := bus.redis.Subscribe(ctx, channel)
	bus.subscribers[subscriptionKey] = pubsub

	go bus.handleClientMessages(clientID, pubsub)

	return nil
}

// SubscribeSession 订阅会话事件
func (bus *RedisEventBus) SubscribeSession(ctx context.Context, sessionID string) error {
	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	subscriptionKey := "session:" + sessionID
	channel := fmt.Sprintf("sse:session:%s", sessionID)

	if pubsub, exists := bus.subscribers[subscriptionKey]; exists {
		pubsub.Close()
		delete(bus.subscribers, subscriptionKey)
	}

	pubsub := bus.redis.Subscribe(ctx, channel)
	bus.subscribers[subscriptionKey] = pubsub

	go bus.handleSessionMessages(sessionID, pubsub)

	servers, err := bus.redis.SMembers(ctx, "sse:session_servers:"+sessionID).Result()
	if err == nil {
		ss := make(map[string]struct{})
		remote := false
		for _, s := range servers {
			ss[s] = struct{}{}
			if s != bus.serverID {
				remote = true
			}
		}
		bus.sessionServers[sessionID] = ss
		bus.sessionRemote[sessionID] = remote
	}

	updatesKey := "session_servers_updates:" + sessionID
	updatesChannel := "sse:session_servers_updates:" + sessionID
	if pubsubUpdates, exists := bus.subscribers[updatesKey]; exists {
		pubsubUpdates.Close()
		delete(bus.subscribers, updatesKey)
	}
	pubsubUpdates := bus.redis.Subscribe(ctx, updatesChannel)
	bus.subscribers[updatesKey] = pubsubUpdates
	go bus.handleSessionServersUpdates(sessionID, pubsubUpdates)

	return nil
}

// UnsubscribeSession 取消订阅会话事件
func (bus *RedisEventBus) UnsubscribeClient(ctx context.Context, clientID string) error {
	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	subscriptionKey := "client:" + clientID
	if pubsub, exists := bus.subscribers[subscriptionKey]; exists {
		pubsub.Close()
		delete(bus.subscribers, subscriptionKey)
	}
	return nil
}

func (bus *RedisEventBus) UnsubscribeSessionIfNoLocalClients(ctx context.Context, sessionID string) error {
	bus.mutex.Lock()
	defer func() {
		bus.mutex.Unlock()
	}()

	subscriptionKey := "session:" + sessionID
	if bus.localClientProvider != nil {
		clients := bus.localClientProvider.GetLocalSessionClients(sessionID)
		if len(clients) == 0 {
			if pubsub, exists := bus.subscribers[subscriptionKey]; exists {
				pubsub.Close()
				delete(bus.subscribers, subscriptionKey)
			}
			delete(bus.sessionRemote, sessionID)
			delete(bus.sessionServers, sessionID)
			updatesKey := "session_servers_updates:" + sessionID
			if pubsubUpdates, exists := bus.subscribers[updatesKey]; exists {
				pubsubUpdates.Close()
				delete(bus.subscribers, updatesKey)
			}
		}
	}
	return nil
}

// RemoveSessionHandler 移除会话中的特定处理器
// 处理器相关逻辑已移除

// handleMessages 处理Redis订阅消息
func (bus *RedisEventBus) handleClientMessages(clientID string, pubsub *redis.PubSub) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("handleClientMessages recovered from panic: %v", r)
		}
	}()
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
			provider := bus.localClientProvider
			bus.mutex.RUnlock()
			if provider != nil {
				lc := provider.GetLocalClient(clientID)
				if lc != nil {
					select {
					case lc.EventChan <- event:
					default:
						log.Printf("客户端缓冲区已满，丢弃消息: %s", clientID)
					}
				}
			}
		case <-bus.ctx.Done():
			return
		}
	}
}

func (bus *RedisEventBus) handleSessionMessages(sessionID string, pubsub *redis.PubSub) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("handleSessionMessages recovered from panic: %v", r)
		}
	}()
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
			if event.OriginServerID == bus.serverID {
				continue
			}
			bus.mutex.RLock()
			provider := bus.localClientProvider
			bus.mutex.RUnlock()
			if provider != nil {
				clients := provider.GetLocalSessionClients(sessionID)
				for _, lc := range clients {
					select {
					case lc.EventChan <- event:
					default:
						log.Printf("客户端缓冲区已满，丢弃消息: %s", lc.ClientID)
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
	bus.sessionServers = make(map[string]map[string]struct{})
	bus.sessionRemote = make(map[string]bool)

	return nil
}

func (bus *RedisEventBus) handleSessionServersUpdates(sessionID string, pubsub *redis.PubSub) {
	defer func() {
		if r := recover(); r != nil {
		}
	}()
	ch := pubsub.Channel()
	for {
		select {
		case msg := <-ch:
			if msg == nil {
				return
			}
			var upd struct {
				SessionID string `json:"sessionID"`
				ServerID  string `json:"serverID"`
				Op        string `json:"op"`
			}
			if err := json.Unmarshal([]byte(msg.Payload), &upd); err != nil {
				continue
			}
			if upd.SessionID != sessionID {
				continue
			}
			bus.mutex.Lock()
			ss := bus.sessionServers[sessionID]
			if ss == nil {
				ss = make(map[string]struct{})
			}
			if upd.Op == "add" {
				ss[upd.ServerID] = struct{}{}
			} else if upd.Op == "remove" {
				delete(ss, upd.ServerID)
			}
			bus.sessionServers[sessionID] = ss
			remote := false
			for s := range ss {
				if s != bus.serverID {
					remote = true
					break
				}
			}
			bus.sessionRemote[sessionID] = remote
			bus.mutex.Unlock()
		case <-bus.ctx.Done():
			return
		}
	}
}
