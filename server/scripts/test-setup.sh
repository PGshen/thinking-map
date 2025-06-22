#!/bin/bash
###
 # @Date: 2025-06-22 10:12:03
 # @LastEditors: peng pgs1108pgs@gmail.com
 # @LastEditTime: 2025-06-22 10:35:17
 # @FilePath: /thinking-map/server/scripts/test-setup.sh
### 

# 测试环境设置脚本

set -e

echo "🚀 设置测试环境..."

# 检查必要的服务是否运行
echo "📋 检查服务状态..."

# 检查 PostgreSQL
if ! pg_isready -h localhost -p 5432 > /dev/null 2>&1; then
    echo "❌ PostgreSQL 服务未运行，请启动 PostgreSQL"
    exit 1
fi
echo "✅ PostgreSQL 服务正常"

# 检查 Redis
if ! redis-cli ping > /dev/null 2>&1; then
    echo "❌ Redis 服务未运行，请启动 Redis"
    exit 1
fi
echo "✅ Redis 服务正常"

# 创建测试数据库（如果不存在）
echo "🗄️  检查测试数据库..."
DB_NAME=${DB_NAME:-thinking_map_test}
DB_USER=${DB_USER:-postgres}

if ! psql -h localhost -U $DB_USER -d postgres -c "SELECT 1 FROM pg_database WHERE datname='$DB_NAME'" | grep -q 1; then
    echo "📝 创建测试数据库: $DB_NAME"
    psql -h localhost -U $DB_USER -d postgres -c "CREATE DATABASE $DB_NAME;"
    echo "✅ 测试数据库已创建: $DB_NAME"
else
    echo "✅ 测试数据库已存在: $DB_NAME"
fi

# 设置环境变量
echo "🔧 设置环境变量..."
export DB_HOST=${DB_HOST:-localhost}
export DB_PORT=${DB_PORT:-5432}
export DB_USER=${DB_USER:-postgres}
export DB_PASSWORD=${DB_PASSWORD:-postgres}
export DB_NAME=${DB_NAME:-thinking_map_test}
export REDIS_ADDR=${REDIS_ADDR:-localhost:6379}
export REDIS_DB=${REDIS_DB:-2}

echo "✅ 测试环境设置完成！"
echo ""
echo "📋 当前配置:"
echo "  数据库: $DB_HOST:$DB_PORT/$DB_NAME"
echo "  Redis: $REDIS_ADDR (DB: $REDIS_DB)"
echo ""
echo "🧪 运行测试:"
echo "  go test ./internal/service -v" 