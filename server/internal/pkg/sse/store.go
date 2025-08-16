package sse

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// SessionStore 定义会话存储接口
type SessionStore interface {
	// 客户端相关操作
	AddClient(clientID string, client *Client) error
	GetClient(clientID string) (*Client, error)
	RemoveClient(clientID string) error

	// 会话相关操作
	AddClientToSession(sessionID, clientID string) error
	RemoveClientFromSession(sessionID, clientID string) error
	GetSessionClients(sessionID string) (map[string]bool, error)

	// 清理操作
	Cleanup() error
}

// MemorySessionStore 基于内存的会话存储实现
type MemorySessionStore struct {
	clients  map[string]*Client
	sessions map[string]map[string]bool
}

// NewMemorySessionStore 创建新的内存会话存储
func NewMemorySessionStore() *MemorySessionStore {
	return &MemorySessionStore{
		clients:  make(map[string]*Client),
		sessions: make(map[string]map[string]bool),
	}
}

func (s *MemorySessionStore) AddClient(clientID string, client *Client) error {
	s.clients[clientID] = client
	return nil
}

func (s *MemorySessionStore) GetClient(clientID string) (*Client, error) {
	client, exists := s.clients[clientID]
	if !exists {
		return nil, fmt.Errorf("client not found: %s", clientID)
	}
	return client, nil
}

func (s *MemorySessionStore) RemoveClient(clientID string) error {
	delete(s.clients, clientID)
	return nil
}

func (s *MemorySessionStore) AddClientToSession(sessionID, clientID string) error {
	if _, exists := s.sessions[sessionID]; !exists {
		s.sessions[sessionID] = make(map[string]bool)
	}
	s.sessions[sessionID][clientID] = true
	return nil
}

func (s *MemorySessionStore) RemoveClientFromSession(sessionID, clientID string) error {
	if clients, exists := s.sessions[sessionID]; exists {
		delete(clients, clientID)
		if len(clients) == 0 {
			delete(s.sessions, sessionID)
		}
	}
	return nil
}

func (s *MemorySessionStore) GetSessionClients(sessionID string) (map[string]bool, error) {
	clients, exists := s.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}
	return clients, nil
}

func (s *MemorySessionStore) Cleanup() error {
	// 清理内存数据
	s.clients = make(map[string]*Client)
	s.sessions = make(map[string]map[string]bool)
	return nil
}

// RedisSessionStore Redis会话存储实现
type RedisSessionStore struct {
	redis *redis.Client
}

// NewRedisSessionStore 创建新的Redis会话存储
func NewRedisSessionStore(client *redis.Client) *RedisSessionStore {
	return &RedisSessionStore{redis: client}
}

// Redis key前缀
const (
	clientPrefix  = "sse:client:"
	sessionPrefix = "sse:session:"
	sessionTTL    = 24 * time.Hour
)

func (s *RedisSessionStore) AddClient(clientID string, client *Client) error {
	data, err := json.Marshal(&client)
	if err != nil {
		return err
	}
	return s.redis.Set(context.Background(), clientPrefix+clientID, data, sessionTTL).Err()
}

func (s *RedisSessionStore) GetClient(clientID string) (*Client, error) {
	data, err := s.redis.Get(context.Background(), clientPrefix+clientID).Bytes()
	if err != nil {
		return nil, err
	}

	var client Client
	if err := json.Unmarshal(data, &client); err != nil {
		return nil, err
	}
	return &client, nil
}

func (s *RedisSessionStore) RemoveClient(clientID string) error {
	return s.redis.Del(context.Background(), clientPrefix+clientID).Err()
}

func (s *RedisSessionStore) AddClientToSession(sessionID, clientID string) error {
	return s.redis.HSet(context.Background(), sessionPrefix+sessionID, clientID, true).Err()
}

func (s *RedisSessionStore) RemoveClientFromSession(sessionID, clientID string) error {
	return s.redis.HDel(context.Background(), sessionPrefix+sessionID, clientID).Err()
}

func (s *RedisSessionStore) GetSessionClients(sessionID string) (map[string]bool, error) {
	result := make(map[string]bool)
	clients, err := s.redis.HGetAll(context.Background(), sessionPrefix+sessionID).Result()
	if err != nil {
		return nil, err
	}

	for clientID := range clients {
		result[clientID] = true
	}
	return result, nil
}

func (s *RedisSessionStore) Cleanup() error {
	// 实际应用中可能需要更复杂的清理逻辑
	return nil
}
