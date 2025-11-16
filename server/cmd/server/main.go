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
	"log"
	"os"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/cloudwego/eino-ext/devops"

	"github.com/PGshen/thinking-map/server/internal/config"
	"github.com/PGshen/thinking-map/server/internal/global"
	"github.com/PGshen/thinking-map/server/internal/model"
	"github.com/PGshen/thinking-map/server/internal/pkg/database"
	"github.com/PGshen/thinking-map/server/internal/pkg/logger"
	"github.com/PGshen/thinking-map/server/internal/pkg/sse"
	"github.com/PGshen/thinking-map/server/internal/pkg/validator"
	"github.com/PGshen/thinking-map/server/internal/repository"
	"github.com/PGshen/thinking-map/server/internal/router"
	"github.com/PGshen/thinking-map/server/internal/service"

	"go.uber.org/zap"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("panic recovered", zap.Any("error", r), zap.String("stack", string(debug.Stack())))
		}
	}()
	// Init eino devops server (only if not disabled)
	if os.Getenv("EINO_DEVOPS_DISABLE") != "true" {
		err := devops.Init(context.Background())
		if err != nil {
			log.Fatalf("[eino dev] init failed, err=%v", err)
			return
		}
	}
	// 加载配置
	cfg, err := config.LoadConfig("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	if err = logger.Init(&cfg.Log); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// 注册验证器
	if err = validator.RegisterValidators(); err != nil {
		logger.Fatal("Failed to register validators", zap.Error(err))
	}

	// 检查PORT环境变量，如果存在则覆盖配置文件中的端口
	if portEnv := os.Getenv("PORT"); portEnv != "" {
		if port, err2 := strconv.Atoi(portEnv); err2 == nil {
			cfg.Server.Port = port
		}
	}

	// 启动服务器
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	logger.Info("Server starting", zap.String("addr", addr))

	// 初始化数据库
	db, err := database.NewPostgresDB(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	if err := db.AutoMigrate(
		&model.User{},
		&model.Message{},
		&model.ThinkingMap{},
		&model.ThinkingNode{},
		&model.RAGRecord{},
	); err != nil {
		logger.Fatal("Failed to migrate database", zap.Error(err))
	}

	// 初始化 Redis
	redisClientRaw, err := database.NewClient(&cfg.Redis)
	if err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	redisClient := redisClientRaw

	// 初始化分布式SSE组件
	// 生成服务器ID
	serverID := fmt.Sprintf("server-%d", time.Now().UnixMicro())

	// 创建Redis连接管理器
	connManager := sse.NewRedisConnectionManager(redisClient, serverID)

	// 创建Redis事件总线（支持本地优化）
	eventBus := sse.NewRedisEventBus(redisClient, connManager, serverID)

	// 初始化全局SSE broker（支持分布式）
	global.InitBroker(eventBus, connManager, serverID, 10*time.Second, 60*time.Second)

	// 初始化 RAG Record 仓库
	global.InitRAGRecordRepository(repository.NewRAGRecordRepository(db))

	// 初始化全局消息管理器
	global.InitMessageManager(repository.NewMessageRepository(db), repository.NewThinkingNodeRepository(db), repository.NewRAGRecordRepository(db), db)

	// 初始化全局节点操作器
	global.InitNodeOperator(repository.NewThinkingNodeRepository(db), repository.NewThinkingMapRepository(db))

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
