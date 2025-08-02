package global

import (
	"fmt"
	"os"
	"testing"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var (
	testDB         *gorm.DB
	testRedis      *redis.Client
	messageManager *MessageManager
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

	messageManager = GetMessageManager()

	code := m.Run()

	// 清理测试数据
	CleanupTestEnvironment(testConfig)

	os.Exit(code)
}
