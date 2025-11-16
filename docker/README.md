# ThinkingMap Docker 部署指南

本目录包含了 ThinkingMap 项目的 Docker 部署配置文件。

## 文件说明

- `docker-compose.yml` - 主要的 Docker Compose 配置文件
- `Dockerfile.backend` - 后端服务的 Dockerfile
- `Dockerfile.frontend` - 前端服务的 Dockerfile
- `.env.example` - 环境变量示例文件
- `README.md` - 本说明文档

## 快速开始

### 1. 准备环境变量

复制环境变量示例文件并填入您的配置：

```bash
cp .env.example .env
```

编辑 `.env` 文件，填入您的 AI 服务 API 密钥：

```bash
# 编辑环境变量文件
nano .env
```

### 2. 启动服务

在 `docker` 目录下运行：

```bash
# 构建并启动所有服务
docker-compose up -d --build

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f
```

### 3. 访问应用

- 前端应用：http://localhost:6000
- 后端 API：http://localhost:8080
- PostgreSQL：localhost:5432
- Redis：localhost:6379

## 服务说明

### 前端服务 (frontend)
- 基于 Next.js 15
- 端口：3000
- 依赖：backend 服务

### 后端服务 (backend)
- 基于 Go 1.24
- 端口：8080
- 依赖：postgres, redis 服务

### 数据库服务 (postgres)
- PostgreSQL 15
- 端口：5432
- 数据库名：thinking_map
- 用户名：postgres
- 密码：pwss

### 缓存服务 (redis)
- Redis 7
- 端口：6379
- 密码：redispass

## 常用命令

```bash
# 启动服务
docker-compose up -d

# 停止服务
docker-compose down

# 重新构建并启动
docker-compose up -d --build

# 查看日志
docker-compose logs -f [service_name]

# 进入容器
docker-compose exec [service_name] sh

# 清理数据卷（注意：会删除所有数据）
docker-compose down -v
```

## 开发模式

如果您想在开发模式下运行，可以只启动数据库和缓存服务：

```bash
# 只启动数据库和缓存
docker-compose up -d postgres redis

# 然后在本地运行前后端服务
cd ../server && go run cmd/server/main.go
cd ../web && pnpm dev
```

## 故障排除

### 1. 端口冲突
如果遇到端口冲突，可以修改 `docker-compose.yml` 中的端口映射。

### 2. 权限问题
确保 Docker 有足够的权限访问项目目录。

### 3. 构建失败
检查网络连接，确保可以下载依赖包。

### 4. 数据库连接失败
等待数据库服务完全启动，通常需要几秒钟时间。

## 生产部署注意事项

1. **安全性**：
   - 修改默认密码
   - 使用 HTTPS
   - 配置防火墙

2. **性能**：
   - 调整数据库连接池大小
   - 配置 Redis 内存限制
   - 使用 CDN 加速静态资源

3. **监控**：
   - 配置日志收集
   - 设置健康检查
   - 监控资源使用情况

4. **备份**：
   - 定期备份数据库
   - 备份配置文件
   - 测试恢复流程