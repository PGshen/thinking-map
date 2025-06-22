# 测试环境配置

本文档说明如何配置和运行测试。

## 环境变量配置

测试支持通过环境变量进行配置，支持的环境变量如下：

### 数据库配置
- `TEST_DB_HOST`: 数据库主机地址 (默认: localhost)
- `TEST_DB_PORT`: 数据库端口 (默认: 5432)
- `TEST_DB_USER`: 数据库用户名 (默认: postgres)
- `TEST_DB_PASSWORD`: 数据库密码 (默认: postgres)
- `TEST_DB_NAME`: 数据库名称 (默认: thinking_map_test)

### Redis 配置
- `TEST_REDIS_ADDR`: Redis 地址 (默认: localhost:6379)
- `TEST_REDIS_PASSWORD`: Redis 密码 (可选)
- `TEST_REDIS_DB`: Redis 数据库编号 (默认: 1)

### JWT 配置
- `TEST_JWT_SECRET`: JWT 密钥 (默认: testsecret)
- `TEST_JWT_ACCESS_TTL`: 访问令牌过期时间 (默认: 10m)
- `TEST_JWT_REFRESH_TTL`: 刷新令牌过期时间 (默认: 20m)

## 配置文件

测试会按以下顺序查找配置文件：

1. `configs/config.test.yaml` (测试专用配置)
2. `configs/config.yaml` (默认配置)

如果使用测试配置文件，可以在其中使用环境变量引用：

```yaml
database:
  host: ${TEST_DB_HOST:-localhost}
  port: ${TEST_DB_PORT:-5432}
  username: ${TEST_DB_USER:-postgres}
  password: ${TEST_DB_PASSWORD:-postgres}
  dbname: ${TEST_DB_NAME:-thinking_map_test}
```

## 运行测试

### 使用默认配置
```bash
cd server
go test ./internal/service -v
```

### 使用自定义环境变量
```bash
cd server
TEST_DB_HOST=localhost \
TEST_DB_PORT=5432 \
TEST_DB_USER=postgres \
TEST_DB_PASSWORD=mypassword \
TEST_DB_NAME=thinking_map_test \
TEST_REDIS_ADDR=localhost:6379 \
go test ./internal/service -v
```

### 使用 .env 文件
```bash
cd server
# 创建 .env.test 文件
echo "TEST_DB_HOST=localhost" > .env.test
echo "TEST_DB_PASSWORD=mypassword" >> .env.test
echo "TEST_DB_NAME=thinking_map_test" >> .env.test

# 运行测试
go test ./internal/service -v
```

## 注意事项

1. **测试数据库**: 测试会使用独立的测试数据库，避免影响生产数据
2. **数据清理**: 每次测试运行后会自动清理测试数据
3. **Redis 隔离**: 测试使用 Redis DB 1，避免与其他环境冲突
4. **连接测试**: 测试启动时会验证数据库和 Redis 连接

## 故障排除

### 数据库连接失败
- 确保 PostgreSQL 服务正在运行
- 检查数据库连接参数是否正确
- 确保测试数据库已创建

### Redis 连接失败
- 确保 Redis 服务正在运行
- 检查 Redis 地址和端口是否正确

### 权限问题
- 确保数据库用户有足够的权限创建和删除表
- 确保 Redis 用户可以访问指定的数据库 