package service

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/PGshen/thinking-map/server/internal/repository"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var (
	testDB            *gorm.DB
	testRedis         *redis.Client
	authSvc           AuthService
	mapSvc            *MapService
	nodeSvc           *NodeService
	dependencyChecker *DependencyChecker
)

func TestMain(m *testing.M) {
	// 设置测试环境
	testConfig, err := SetupTestEnvironment()
	if err != nil {
		panic(fmt.Sprintf("failed to setup test environment: %v", err))
	}

	// 设置全局变量
	testDB = testConfig.DB
	testRedis = testConfig.Redis

	// 解析 JWT 配置
	cfg, _ := LoadTestConfig()
	expireDuration, err := time.ParseDuration(cfg.JWT.Expire)
	if err != nil {
		expireDuration = time.Minute * 10 // 默认值
	}

	// 初始化 AuthService
	authSvc = NewAuthService(testDB, testRedis, JWTConfig{
		SecretKey:       cfg.JWT.Secret,
		AccessTokenTTL:  expireDuration,
		RefreshTokenTTL: expireDuration * 2,
		TokenIssuer:     "test",
	})
	mapRepo := repository.NewThinkingMapRepository(testDB)
	mapSvc = NewMapService(mapRepo)

	nodeRepo := repository.NewThinkingNodeRepository(testDB)
	nodeSvc = NewNodeService(nodeRepo, mapRepo)
	dependencyChecker = NewDependencyChecker(nodeRepo)

	code := m.Run()

	// 清理测试数据
	CleanupTestEnvironment(testConfig)

	os.Exit(code)
}
