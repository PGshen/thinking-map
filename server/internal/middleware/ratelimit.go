/*
 * @Date: 2025-06-18 22:57:15
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-06-18 23:44:34
 * @FilePath: /thinking-map/server/internal/middleware/ratelimit.go
 */
package middleware

import (
	"context"
	"sync"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"golang.org/x/time/rate"
)

// RateLimiter 限流器
type RateLimiter struct {
	ips        map[string]*rate.Limiter
	lastAccess map[string]time.Time
	mu         *sync.RWMutex
	rate       rate.Limit
	burst      int
	ttl        time.Duration
	ticker     *time.Ticker
}

// NewRateLimiter 创建限流器
func NewRateLimiter(r rate.Limit, b int, ttl time.Duration) *RateLimiter {
	limiter := &RateLimiter{
		ips:        make(map[string]*rate.Limiter),
		lastAccess: make(map[string]time.Time),
		mu:         &sync.RWMutex{},
		rate:       r,
		burst:      b,
		ttl:        ttl,
		ticker:     time.NewTicker(ttl),
	}

	// 启动清理过期限流器的协程
	go limiter.cleanup()
	return limiter
}

// cleanup 清理过期的限流器
func (rl *RateLimiter) cleanup() {
	for range rl.ticker.C {
		rl.mu.Lock()
		for ip, lastTime := range rl.lastAccess {
			if time.Since(lastTime) > rl.ttl {
				delete(rl.ips, ip)
				delete(rl.lastAccess, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// getLimiter 获取或创建限流器
func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.ips[ip]
	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.ips[ip] = limiter
	}
	rl.lastAccess[ip] = time.Now()

	return limiter
}

// RateLimit 限流中间件
func RateLimit(r rate.Limit, b int, ttl time.Duration) app.HandlerFunc {
	limiter := NewRateLimiter(r, b, ttl)
	return func(ctx context.Context, c *app.RequestContext) {
		ip := c.ClientIP()
		if !limiter.getLimiter(ip).Allow() {
			c.JSON(429, map[string]interface{}{
				"code":    429,
				"message": "请求过于频繁，请稍后再试",
			})
			c.Abort()
			return
		}
		c.Next(ctx)
	}
}
