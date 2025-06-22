/*
 * @Date: 2025-06-22 14:56:38
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-06-22 15:09:02
 * @FilePath: /thinking-map/server/internal/service/test_utils.go
 */
package service

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/thinking-map/server/internal/config"
)

// TestConfig 测试配置结构体
type TestConfig struct {
	DB    *gorm.DB
	Redis *redis.Client
}

// LoadTestConfig 加载测试配置
func LoadTestConfig() (*config.Config, error) {
	// 优先尝试加载测试配置文件
	configPath := "../../configs/config.test.yaml"
	if _, err := os.Stat(configPath); err != nil {
		// 如果测试配置文件不存在，使用默认配置
		configPath = "../../configs/config.yaml"
	}
	_ = godotenv.Load("../../.env")
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load test config: %w", err)
	}

	// 覆盖为测试数据库名称
	if os.Getenv("TEST_DB_NAME") != "" {
		cfg.Database.DBName = os.Getenv("TEST_DB_NAME")
	} else if cfg.Database.DBName == "thinking_map" {
		cfg.Database.DBName = "thinking_map_test"
	}

	return cfg, nil
}

// InitTestDatabase 初始化测试数据库
func InitTestDatabase(cfg *config.Config) (*gorm.DB, error) {
	// 构建数据库 DSN
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Shanghai",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.Username, cfg.Database.Password, cfg.Database.DBName, cfg.Database.SSLMode)

	// 连接数据库
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect db: %w", err)
	}

	return db, nil
}

// InitTestRedis 初始化测试Redis
func InitTestRedis(cfg *config.Config) (*redis.Client, error) {
	// 连接 Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       2, // 测试专用数据库
	})

	// 测试 Redis 连接
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect redis: %w", err)
	}

	// 清理 Redis
	if err := redisClient.FlushDB(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to flush redis: %w", err)
	}

	return redisClient, nil
}

// SetupTestEnvironment 设置测试环境
func SetupTestEnvironment() (*TestConfig, error) {
	// 加载测试配置
	cfg, err := LoadTestConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load test config: %w", err)
	}

	// 初始化数据库
	db, err := InitTestDatabase(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to init test database: %w", err)
	}

	// 初始化Redis
	redisClient, err := InitTestRedis(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to init test redis: %w", err)
	}

	return &TestConfig{
		DB:    db,
		Redis: redisClient,
	}, nil
}

// CleanupTestEnvironment 清理测试环境
func CleanupTestEnvironment(testConfig *TestConfig) {
	if testConfig != nil {
		if testConfig.DB != nil {
			_ = testConfig.DB.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE").Error
		}
		if testConfig.Redis != nil {
			_ = testConfig.Redis.FlushDB(context.Background()).Err()
		}
	}
}
