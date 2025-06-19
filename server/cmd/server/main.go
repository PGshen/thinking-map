package main

import (
	"fmt"
	"log"

	"github.com/thinking-map/server/internal/config"
	"github.com/thinking-map/server/internal/pkg/logger"
	"github.com/thinking-map/server/internal/pkg/validator"
	"go.uber.org/zap"
)

func main() {
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
	// TODO: 启动 HTTP 服务器
}
