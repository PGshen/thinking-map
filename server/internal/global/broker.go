/*
 * @Date: 2025-01-27 00:00:00
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-01-27 00:00:00
 * @FilePath: /thinking-map/server/internal/global/broker.go
 */
package global

import (
	"sync"
	"time"

	"github.com/PGshen/thinking-map/server/internal/pkg/sse"
)

var (
	// GlobalBroker 全局SSE broker实例
	GlobalBroker *sse.Broker
	brokerOnce   sync.Once
)

// InitBroker 初始化全局broker
func InitBroker(eventBus sse.EventBus, connManager sse.ConnectionManager, serverID string, pingInterval, clientTimeout time.Duration) {
	brokerOnce.Do(func() {
		GlobalBroker = sse.NewBroker(eventBus, connManager, serverID, pingInterval, clientTimeout)
	})
}

// GetBroker 获取全局broker实例
func GetBroker() *sse.Broker {
	if GlobalBroker == nil {
		panic("broker not initialized, call InitBroker first")
	}
	return GlobalBroker
}
