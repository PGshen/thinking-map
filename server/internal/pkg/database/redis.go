/*
 * @Date: 2025-06-19 00:19:22
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-06-20 23:36:26
 * @FilePath: /thinking-map/server/internal/pkg/database/redis.go
 */
package database

import (
	"context"

	"github.com/PGshen/thinking-map/server/internal/config"

	"github.com/redis/go-redis/v9"
)

// NewClient creates a new Redis client
func NewClient(cfg *config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
	})

	// Test connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return client, nil
}
