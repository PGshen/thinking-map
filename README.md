# Thinking Map

一个基于思维导图的智能思考辅助系统。

## 项目结构

```
thinking-map/
├── server/          # 后端服务 (Go)
├── web/            # 前端应用 (React)
├── docs/           # 项目文档
└── logs/           # 日志文件
```

## 快速开始

### 后端服务

1. 进入后端目录：
```bash
cd server
```

2. 安装依赖：
```bash
go mod download
```

3. 配置环境变量：
```bash
# 复制配置文件
cp configs/config.yaml configs/config.local.yaml
# 编辑配置文件，设置数据库和 Redis 连接信息
```

4. 运行服务：
```bash
go run cmd/server/main.go
```

### 测试

1. 设置测试环境：
```bash
cd server
./scripts/test-setup.sh
```

2. 运行测试：
```bash
go test ./internal/service -v
```

详细测试配置说明请参考 [TESTING.md](server/TESTING.md)。

## 开发

### 环境要求

- Go 1.23+
- PostgreSQL 14+
- Redis 7+
- Node.js 18+ (前端开发)

### 数据库设置

1. 创建数据库：
```sql
CREATE DATABASE thinking_map;
CREATE DATABASE thinking_map_test;
```

2. 运行迁移：
```bash
cd server
go run cmd/migrate/main.go
```

## 文档

- [后端文档](docs/backend.md)
- [API 文档](docs/api.md)
- [部署文档](docs/deployment.md)

## 许可证

MIT License 