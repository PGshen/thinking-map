/*
 * @Date: 2025-06-22 14:56:38
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-06-22 15:09:02
 * @FilePath: /thinking-map/server/internal/service/test_utils.go
 */
package global

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/PGshen/thinking-map/server/internal/config"
	"github.com/PGshen/thinking-map/server/internal/pkg/sse"
	"github.com/PGshen/thinking-map/server/internal/repository"
)

// getProjectRoot finds the project root directory by searching for the go.mod file.
func getProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found")
		}
		dir = parent
	}
}

// TestConfig 测试配置结构体
type TestConfig struct {
	DB    *gorm.DB
	Redis *redis.Client
}

// LoadTestConfig 加载测试配置
func LoadTestConfig() (*config.Config, error) {
	projectRoot, err := getProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	// 优先尝试加载测试配置文件
	configPath := filepath.Join(projectRoot, "configs", "config.test.yaml")
	if _, err2 := os.Stat(configPath); err2 != nil {
		// 如果测试配置文件不存在，使用默认配置
		configPath = filepath.Join(projectRoot, "configs", "config.yaml")
	}

	envPath := filepath.Join(projectRoot, ".env")
	_ = godotenv.Load(envPath)

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
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.Username, cfg.Database.Password, cfg.Database.DBName)

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

	// 初始化分布式SSE组件
	// 生成服务器ID
	serverID := fmt.Sprintf("server-%d", time.Now().Unix())

	// 创建Redis连接管理器
	connManager := sse.NewRedisConnectionManager(redisClient, serverID)

	// 创建Redis事件总线（支持本地优化）
	eventBus := sse.NewRedisEventBus(redisClient, connManager, serverID)

	// 初始化全局SSE broker（支持分布式）
	InitBroker(eventBus, connManager, serverID, 10*time.Second, 60*time.Second)

	// 初始化全局消息管理器
	InitMessageManager(repository.NewMessageRepository(db), repository.NewThinkingNodeRepository(db), repository.NewRAGRecordRepository(db), db)

	// 初始化全局节点操作器
	InitNodeOperator(repository.NewThinkingNodeRepository(db), repository.NewThinkingMapRepository(db))

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
