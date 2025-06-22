# Service 层测试工具使用指南

本目录提供了可复用的测试工具，用于简化 service 层测试的配置和初始化工作。

## 文件结构

- `test_utils.go` - 测试工具函数
- `auth_test.go` - 认证服务测试示例

## 主要功能

### 1. 配置加载

```go
// LoadTestConfig 加载测试配置
cfg, err := LoadTestConfig()
if err != nil {
    t.Fatalf("failed to load test config: %v", err)
}
```

### 2. 数据库初始化

```go
// InitTestDatabase 初始化测试数据库
db, err := InitTestDatabase(cfg)
if err != nil {
    t.Fatalf("failed to init test database: %v", err)
}
```

### 3. Redis 初始化

```go
// InitTestRedis 初始化测试Redis
redis, err := InitTestRedis(cfg)
if err != nil {
    t.Fatalf("failed to init test redis: %v", err)
}
```

### 4. 完整测试环境设置

```go
// SetupTestEnvironment 设置完整的测试环境
testConfig, err := SetupTestEnvironment()
if err != nil {
    t.Fatalf("failed to setup test environment: %v", err)
}
defer CleanupTestEnvironment(testConfig)
```

## 使用方式

### 方式1: 使用全局测试环境（推荐）

适用于需要共享数据库连接的测试：

```go
package service

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/assert"
)

func TestMyService(t *testing.T) {
    // 使用全局的 testDB 和 testRedis
    svc := NewMyService(testDB, testRedis)
    
    ctx := context.Background()
    result, err := svc.MyMethod(ctx)
    assert.NoError(t, err)
    assert.NotNil(t, result)
}
```

### 方式2: 创建独立测试环境

适用于需要隔离的测试：

```go
func TestMyService_Isolated(t *testing.T) {
    // 设置独立的测试环境
    testConfig, err := SetupTestEnvironment()
    assert.NoError(t, err)
    defer CleanupTestEnvironment(testConfig)

    // 使用独立的数据库和Redis连接
    svc := NewMyService(testConfig.DB, testConfig.Redis)
    
    ctx := context.Background()
    result, err := svc.MyMethod(ctx)
    assert.NoError(t, err)
    assert.NotNil(t, result)
}
```

### 方式3: 手动初始化

适用于需要自定义配置的测试：

```go
func TestMyService_Custom(t *testing.T) {
    // 加载测试配置
    cfg, err := LoadTestConfig()
    assert.NoError(t, err)

    // 自定义配置
    cfg.Database.DBName = "custom_test_db"

    // 初始化数据库
    db, err := InitTestDatabase(cfg)
    assert.NoError(t, err)

    // 初始化Redis
    redis, err := InitTestRedis(cfg)
    assert.NoError(t, err)

    // 清理资源
    defer func() {
        if db != nil {
            _ = db.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE").Error
        }
        if redis != nil {
            _ = redis.FlushDB(context.Background()).Err()
        }
    }()

    // 使用自定义的数据库和Redis连接
    svc := NewMyService(db, redis)
    
    ctx := context.Background()
    result, err := svc.MyMethod(ctx)
    assert.NoError(t, err)
    assert.NotNil(t, result)
}
```

## TestMain 函数示例

如果你的测试文件需要全局初始化，可以这样写：

```go
func TestMain(m *testing.M) {
    // 设置测试环境
    testConfig, err := SetupTestEnvironment()
    if err != nil {
        panic(fmt.Sprintf("failed to setup test environment: %v", err))
    }

    // 设置全局变量（如果需要）
    // testDB = testConfig.DB
    // testRedis = testConfig.Redis

    code := m.Run()

    // 清理测试数据
    CleanupTestEnvironment(testConfig)

    os.Exit(code)
}
```

## 配置说明

### 环境变量

- `TEST_DB_NAME`: 指定测试数据库名称，如果不设置则使用默认的 `thinking_map_test`

### 配置文件

- 优先加载 `../../configs/config.test.yaml`
- 如果不存在，则加载 `../../configs/config.yaml`
- 同时加载 `../../.env` 环境变量文件

### 数据库配置

- 使用 PostgreSQL 数据库
- 自动迁移 User 模型
- 测试完成后自动清理数据

### Redis 配置

- 使用 DB 2 作为测试专用数据库
- 测试完成后自动清理数据

## 注意事项

1. 确保测试数据库和 Redis 服务正在运行
2. 测试会清理数据库和 Redis 中的数据，请勿在生产环境使用
3. 每个测试文件只能有一个 `TestMain` 函数
4. 建议使用方式1（全局测试环境）来避免重复初始化 