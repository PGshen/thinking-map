/*
 * @Date: 2025-06-18 22:16:47
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-06-20 23:52:18
 * @FilePath: /thinking-map/server/cmd/server/main.go
 */
package main

import (
	"context"
	"fmt"
	"github.com/cloudwego/eino-ext/devops"
	"log"
	"time"

	"github.com/PGshen/thinking-map/server/internal/config"
	"github.com/PGshen/thinking-map/server/internal/pkg/database"
	"github.com/PGshen/thinking-map/server/internal/pkg/global"
	"github.com/PGshen/thinking-map/server/internal/pkg/logger"
	"github.com/PGshen/thinking-map/server/internal/pkg/sse"
	"github.com/PGshen/thinking-map/server/internal/pkg/validator"
	"github.com/PGshen/thinking-map/server/internal/router"
	"github.com/PGshen/thinking-map/server/internal/service"

	"go.uber.org/zap"
)

func main() {
	// Init eino devops server
	err := devops.Init(context.Background())
	if err != nil {
		log.Fatalf("[eino dev] init failed, err=%v", err)
		return
	}
	// 加载配置
	cfg, err := config.LoadConfig("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	if err := logger.Init(&cfg.Log); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// 注册验证器
	if err := validator.RegisterValidators(); err != nil {
		logger.Fatal("Failed to register validators", zap.Error(err))
	}

	// 启动服务器
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	logger.Info("Server starting", zap.String("addr", addr))

	// 初始化数据库
	db, err := database.NewPostgresDB(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	// 初始化 Redis
	redisClientRaw, err := database.NewClient(&cfg.Redis)
	if err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	redisClient := redisClientRaw

	// 初始化全局SSE broker
	store := sse.NewMemorySessionStore()
	global.InitBroker(store, 10*time.Second, 60*time.Second)

	// 解析 JWT 配置
	expireDuration, err := time.ParseDuration(cfg.JWT.Expire)
	if err != nil {
		logger.Fatal("Invalid JWT expire duration", zap.Error(err))
	}
	jwtConfig := service.JWTConfig{
		SecretKey:       cfg.JWT.Secret,
		AccessTokenTTL:  expireDuration,
		RefreshTokenTTL: expireDuration * 2, // 可根据实际需求调整
		TokenIssuer:     "thinking-map",
	}

	r := router.SetupRouter(db, redisClient, jwtConfig)
	if err := r.Run(addr); err != nil {
		logger.Fatal("Failed to start HTTP server", zap.Error(err))
	}
}
