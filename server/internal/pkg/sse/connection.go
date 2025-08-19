/*
 * SSE连接管理器
 * 管理客户端连接的生命周期和状态
 */
package sse

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// ConnectionState 连接状态
type ConnectionState string

const (
	Connected    ConnectionState = "connected"
	Disconnected ConnectionState = "disconnected"
	Reconnecting ConnectionState = "reconnecting"
)

// ClientConnection 客户端连接信息
type ClientConnection struct {
	ID        string          `json:"id"`
	SessionID string          `json:"session_id"`
	ServerID  string          `json:"server_id"` // 服务器实例ID
	State     ConnectionState `json:"state"`
	LastSeen  time.Time       `json:"last_seen"`
	CreatedAt time.Time       `json:"created_at"`
}

// ConnectionManager 连接管理器接口
type ConnectionManager interface {
	// 注册连接
	RegisterConnection(ctx context.Context, conn *ClientConnection) error
	// 注销连接
	UnregisterConnection(ctx context.Context, clientID string) error
	// 更新连接状态
	UpdateConnectionState(ctx context.Context, clientID string, state ConnectionState) error
	// 更新最后活跃时间
	UpdateLastSeen(ctx context.Context, clientID string) error
	// 获取连接信息
	GetConnection(ctx context.Context, clientID string) (*ClientConnection, error)
	// 获取会话中的所有连接
	GetSessionConnections(ctx context.Context, sessionID string) ([]*ClientConnection, error)
	// 获取服务器实例的所有连接
	GetServerConnections(ctx context.Context, serverID string) ([]*ClientConnection, error)
	// 清理过期连接
	CleanupExpiredConnections(ctx context.Context, timeout time.Duration) error
}

// RedisConnectionManager Redis实现的连接管理器
type RedisConnectionManager struct {
	redis    *redis.Client
	serverID string
	mutex    sync.RWMutex
}

// NewRedisConnectionManager 创建Redis连接管理器
func NewRedisConnectionManager(redisClient *redis.Client, serverID string) *RedisConnectionManager {
	return &RedisConnectionManager{
		redis:    redisClient,
		serverID: serverID,
	}
}

// Redis key patterns
const (
	connectionPrefix = "sse:connection:"
	sessionConnPrefix = "sse:session_conn:"
	serverConnPrefix = "sse:server_conn:"
	connectionTTL = 5 * time.Minute // 连接信息TTL
)

// RegisterConnection 注册连接
func (cm *RedisConnectionManager) RegisterConnection(ctx context.Context, conn *ClientConnection) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	conn.ServerID = cm.serverID
	conn.CreatedAt = time.Now()
	conn.LastSeen = time.Now()
	conn.State = Connected

	data, err := json.Marshal(conn)
	if err != nil {
		return fmt.Errorf("marshal connection failed: %w", err)
	}

	pipe := cm.redis.Pipeline()
	
	// 存储连接信息
	pipe.Set(ctx, connectionPrefix+conn.ID, data, connectionTTL)
	
	// 添加到会话连接集合
	pipe.SAdd(ctx, sessionConnPrefix+conn.SessionID, conn.ID)
	pipe.Expire(ctx, sessionConnPrefix+conn.SessionID, connectionTTL)
	
	// 添加到服务器连接集合
	pipe.SAdd(ctx, serverConnPrefix+cm.serverID, conn.ID)
	pipe.Expire(ctx, serverConnPrefix+cm.serverID, connectionTTL)

	_, err = pipe.Exec(ctx)
	return err
}

// UnregisterConnection 注销连接
func (cm *RedisConnectionManager) UnregisterConnection(ctx context.Context, clientID string) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// 获取连接信息
	conn, err := cm.GetConnection(ctx, clientID)
	if err != nil {
		return err
	}

	pipe := cm.redis.Pipeline()
	
	// 删除连接信息
	pipe.Del(ctx, connectionPrefix+clientID)
	
	// 从会话连接集合中移除
	pipe.SRem(ctx, sessionConnPrefix+conn.SessionID, clientID)
	
	// 从服务器连接集合中移除
	pipe.SRem(ctx, serverConnPrefix+cm.serverID, clientID)

	_, err = pipe.Exec(ctx)
	return err
}

// UpdateConnectionState 更新连接状态
func (cm *RedisConnectionManager) UpdateConnectionState(ctx context.Context, clientID string, state ConnectionState) error {
	conn, err := cm.GetConnection(ctx, clientID)
	if err != nil {
		return err
	}

	conn.State = state
	conn.LastSeen = time.Now()

	data, err := json.Marshal(conn)
	if err != nil {
		return fmt.Errorf("marshal connection failed: %w", err)
	}

	return cm.redis.Set(ctx, connectionPrefix+clientID, data, connectionTTL).Err()
}

// UpdateLastSeen 更新最后活跃时间
func (cm *RedisConnectionManager) UpdateLastSeen(ctx context.Context, clientID string) error {
	conn, err := cm.GetConnection(ctx, clientID)
	if err != nil {
		return err
	}

	conn.LastSeen = time.Now()

	data, err := json.Marshal(conn)
	if err != nil {
		return fmt.Errorf("marshal connection failed: %w", err)
	}

	return cm.redis.Set(ctx, connectionPrefix+clientID, data, connectionTTL).Err()
}

// GetConnection 获取连接信息
func (cm *RedisConnectionManager) GetConnection(ctx context.Context, clientID string) (*ClientConnection, error) {
	data, err := cm.redis.Get(ctx, connectionPrefix+clientID).Bytes()
	if err != nil {
		return nil, err
	}

	var conn ClientConnection
	if err := json.Unmarshal(data, &conn); err != nil {
		return nil, fmt.Errorf("unmarshal connection failed: %w", err)
	}

	return &conn, nil
}

// GetSessionConnections 获取会话中的所有连接
func (cm *RedisConnectionManager) GetSessionConnections(ctx context.Context, sessionID string) ([]*ClientConnection, error) {
	clientIDs, err := cm.redis.SMembers(ctx, sessionConnPrefix+sessionID).Result()
	if err != nil {
		return nil, err
	}

	var connections []*ClientConnection
	for _, clientID := range clientIDs {
		conn, err := cm.GetConnection(ctx, clientID)
		if err != nil {
			log.Printf("Failed to get connection %s: %v", clientID, err)
			continue
		}
		connections = append(connections, conn)
	}

	return connections, nil
}

// GetServerConnections 获取服务器实例的所有连接
func (cm *RedisConnectionManager) GetServerConnections(ctx context.Context, serverID string) ([]*ClientConnection, error) {
	clientIDs, err := cm.redis.SMembers(ctx, serverConnPrefix+serverID).Result()
	if err != nil {
		return nil, err
	}

	var connections []*ClientConnection
	for _, clientID := range clientIDs {
		conn, err := cm.GetConnection(ctx, clientID)
		if err != nil {
			log.Printf("Failed to get connection %s: %v", clientID, err)
			continue
		}
		connections = append(connections, conn)
	}

	return connections, nil
}

// CleanupExpiredConnections 清理过期连接
func (cm *RedisConnectionManager) CleanupExpiredConnections(ctx context.Context, timeout time.Duration) error {
	// 获取当前服务器的所有连接
	connections, err := cm.GetServerConnections(ctx, cm.serverID)
	if err != nil {
		return err
	}

	now := time.Now()
	for _, conn := range connections {
		if now.Sub(conn.LastSeen) > timeout {
			log.Printf("Cleaning up expired connection: %s", conn.ID)
			if err := cm.UnregisterConnection(ctx, conn.ID); err != nil {
				log.Printf("Failed to cleanup connection %s: %v", conn.ID, err)
			}
		}
	}

	return nil
}