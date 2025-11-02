# ThinkingMap 后端技术文档

## 1. 项目概述

### 1.1 项目简介
ThinkingMap 后端服务是一个智能化思维导图系统，提供思维导图节点管理、AI 问题拆解、结论生成等核心功能。系统采用现代化架构设计，支持实时通信和智能体交互。

### 1.2 核心功能
- 思维导图节点的 CRUD 操作
- 基于 AI 的问题智能拆解
- 实时结论生成和推理
- SSE 长连接实时通信
- 用户认证和权限管理
- 会话状态管理

### 1.3 性能目标
- 支持并发用户数：1000+
- API 响应时间：< 200ms
- SSE 连接稳定性：99.9%
- 数据库查询响应：< 100ms

## 2. 技术选型

### 2.1 核心技术栈
| 技术组件 | 选型 | 版本要求 | 选型理由 |
| --- | --- | --- | --- |
| 编程语言 | Golang | 1.24+ | 高性能、并发友好、生态成熟 |
| HTTP框架 | Gin | v1.10+ | 高性能、中间件丰富、社区活跃 |
| 智能体框架 | cloudwego/eino | v0.3+ | 专业AI Agent框架，支持复杂推理链 |
| 数据库 | PostgreSQL | 14+ | JSONB支持、事务完整性、扩展性好 |
| 缓存 | Redis | 7.0+ | 高性能缓存、发布订阅、数据结构丰富 |
| ORM | GORM | v1.25+ | 功能完善、社区活跃、PostgreSQL支持好 |
| 认证 | JWT | v5.2+ | 无状态、跨域友好、标准化 |
| 配置管理 | Viper | v1.16+ | 多格式支持、环境变量集成 |
| 日志 | Zap | v1.27+ | 高性能、结构化日志、字段丰富 |

### 2.2 依赖服务
- **PostgreSQL**: 主数据存储
- **Redis**: 缓存和会话存储
- **AI服务**: 大语言模型接口（OpenAI/Claude/Deepseek等）

## 3. 系统架构

### 3.1 整体架构

ThinkingMap 后端采用分层架构，主要包括：

- **表示层**：处理HTTP请求和SSE通信
- **业务层**：实现核心业务逻辑
- **智能体层**：提供AI推理和问题解决能力
- **数据层**：负责数据持久化和缓存

```
┌─────────────────────────────────────────────────────────────┐
│                  表示层 (Presentation Layer)                 │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │ REST API    │  │ SSE Handler │  │  中间件             │  │
│  │             │  │             │  │                     │  │
│  │ • CRUD 操作  │  │ • 实时通信  │  │ • 认证             │  │
│  │ • 认证      │  │ • 事件广播  │  │ • 权限检查          │  │
│  │ • 数据验证  │  │ • 心跳保持  │  │ • 日志记录          │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│                    业务层 (Business Layer)                   │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │ 节点服务    │  │ 思考服务    │  │   用户服务          │  │
│  │             │  │             │  │                     │  │
│  │ • 图管理    │  │ • 智能体编排│  │ • 认证              │  │
│  │ • CRUD 逻辑 │  │ • 会话管理  │  │ • 授权              │  │
│  │ • 数据验证  │  │ • 上下文管理│  │ • 用户资料管理      │  │
│  │ • 事件处理  │  │ • 工作流    │  │ • 会话控制          │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│                      智能体层 (Agent Layer)                  │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │ 理解智能体  │  │ 分解智能体  │  │   结论智能体        │  │
│  │             │  │             │  │                     │  │
│  │ • 问题理解  │  │ • 问题分解  │  │ • 结论生成          │  │
│  │ • 意图识别  │  │ • 子任务生成│  │ • 证据合成          │  │
│  │ • 上下文分析│  │ • 依赖分析  │  │ • 质量保证          │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│                       数据层 (Data Layer)                    │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │ 仓储层      │  │    缓存     │  │   消息队列          │  │
│  │             │  │             │  │                     │  │
│  │ • PostgreSQL│  │ • Redis     │  │ • 事件流            │  │
│  │ • ORM       │  │ • 内存      │  │ • 发布/订阅         │  │
│  │ • 迁移      │  │ • 会话      │  │ • 任务队列          │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### 3.2 目录结构

```
server/
├── cmd/                 # 应用入口点
│   └── server/          # 主服务器入口
├── configs/             # 配置文件
│   ├── config.yaml      # 主配置
│   └── config.test.yaml # 测试配置
├── internal/            # 内部应用代码
│   ├── agent/           # 智能体实现
│   │   ├── base/        # 基础智能体
│   │   ├── callback/    # 回调处理
│   │   ├── conclusion/  # 结论生成
│   │   ├── decomposition/ # 问题分解
│   │   ├── llmmodel/    # LLM模型接口
│   │   ├── tool/        # 智能体工具
│   │   └── understanding/ # 问题理解
│   ├── config/          # 配置加载
│   ├── global/          # 全局组件
│   ├── handler/         # HTTP处理器
│   │   └── thinking/    # 思考相关处理器
│   ├── middleware/      # 中间件
│   ├── model/           # 数据模型
│   │   └── dto/         # 数据传输对象
│   ├── pkg/             # 通用包
│   │   ├── comm/        # 通信
│   │   ├── database/    # 数据库
│   │   ├── jwt/         # JWT
│   │   ├── logger/      # 日志
│   │   ├── sse/         # SSE
│   │   ├── utils/       # 工具函数
│   │   └── validator/   # 验证器
│   ├── repository/      # 数据仓储
│   ├── router/          # 路由
│   └── service/         # 业务服务
└── scripts/             # 脚本
    ├── db.sql           # 数据库脚本
    └── test-setup.sh    # 测试设置
```

## 4. 核心模块

### 4.1 用户认证模块

基于JWT的认证系统，支持注册、登录、刷新令牌和登出功能。

**主要组件**:
- `AuthService`: 处理用户认证逻辑
- `AuthHandler`: 处理认证相关HTTP请求
- `AuthMiddleware`: 验证请求中的JWT令牌

### 4.2 思维导图模块

管理思维导图及其元数据。

**主要组件**:
- `MapService`: 处理思维导图业务逻辑
- `MapHandler`: 处理思维导图相关HTTP请求
- `MapOwnershipMiddleware`: 验证用户对思维导图的所有权

### 4.3 节点模块

管理思维导图中的节点及其关系。

**主要组件**:
- `NodeService`: 处理节点业务逻辑
- `NodeHandler`: 处理节点相关HTTP请求
- `NodeOwnershipMiddleware`: 验证用户对节点的所有权

### 4.4 智能体模块

ThinkingMap 系统的核心是其智能体系统，基于 cloudwego/eino 框架实现，提供 AI 驱动的问题解决能力。

#### 4.4.1 智能体架构

智能体系统采用多智能体协作架构，主要包括以下几类智能体：

**理解智能体 (Understanding Agent)**
- **职责**：理解用户输入的问题，分析问题意图和上下文
- **实现**：基于 React 模式的智能体，使用 LLM 进行问题理解
- **关键组件**：
  - `understanding.go`: 实现问题理解的核心逻辑
  - `prompt.go`: 定义与 LLM 交互的提示模板

**分解智能体 (Decomposition Agent)**
- **职责**：将复杂问题分解为子任务，建立任务依赖关系
- **实现**：基于 React 模式的智能体，使用 LLM 进行问题分解
- **关键组件**：
  - `decomposition.go`: 实现问题分解的核心逻辑
  - `analysis.go`: 分析问题结构和复杂度
  - `prompt.go`: 定义问题分解的提示模板

**结论智能体 (Conclusion Agent)**
- **职责**：基于子任务结果生成最终结论
- **实现**：多版本迭代，当前使用 v3 版本
- **关键组件**：
  - `generation.go`: 实现结论生成的核心逻辑
  - `optimization.go`: 优化生成的结论质量
  - `prompt.go`: 定义结论生成的提示模板

#### 4.4.2 智能体工具

智能体系统提供了一系列工具，增强智能体的能力：

- **节点操作工具**：允许智能体创建、更新和查询节点
- **消息工具**：用于智能体间通信和消息传递
- **搜索工具**：支持智能体进行信息检索

#### 4.4.3 LLM 模型集成

系统支持多种 LLM 模型，通过统一接口进行调用：

- **OpenAI 模型**：支持 GPT-3.5/GPT-4 系列
- **Claude 模型**：支持 Claude 系列模型
- **自定义模型**：可扩展支持其他 LLM 模型

#### 4.4.4 智能体回调系统

智能体执行过程中的事件通过回调系统进行处理：

- **日志回调**：记录智能体执行过程
- **进度回调**：通过 SSE 实时更新执行进度
- **结果回调**：处理智能体执行结果

### 4.5 上下文管理模块

上下文管理是 ThinkingMap 系统的关键组件，负责管理思维导图的上下文信息，为智能体提供必要的背景知识。

#### 4.5.1 上下文管理器

`ContextManager` 是上下文管理的核心组件，主要职责包括：

- **上下文收集**：基于导图结构自动收集相关上下文
- **上下文组织**：将收集的上下文按照不同类型进行组织
- **上下文提供**：为智能体提供结构化的上下文信息

#### 4.5.2 上下文类型

系统支持多种类型的上下文：

- **祖先上下文 (AncestorsContext)**：节点的父节点和祖先节点信息
- **依赖上下文 (DependencyContext)**：节点依赖的其他节点信息
- **子节点上下文 (ChildrenContext)**：节点的子节点信息
- **对话上下文 (ConversationContext)**：与节点相关的对话历史
- **用户上下文 (UserContext)**：用户相关的上下文信息

#### 4.5.3 上下文获取方法

`ContextManager` 提供多种方法获取上下文：

- **GetNodeContext**：获取节点的基本上下文
- **GetNodeContextWithConversation**：获取包含对话历史的节点上下文
- **GetNodeContextWithDependencies**：获取包含依赖关系的节点上下文
- **GetFullNodeContext**：获取节点的完整上下文

#### 4.5.4 上下文应用

上下文信息在系统中的主要应用：

- **问题理解**：为理解智能体提供问题背景
- **问题分解**：为分解智能体提供问题结构
- **结论生成**：为结论智能体提供依赖信息
- **依赖检查**：用于检查节点间的依赖关系

### 4.6 实时通信模块

ThinkingMap 系统使用 Server-Sent Events (SSE) 实现实时通信，支持服务器向客户端推送事件和数据。

#### 4.6.1 SSE 架构

SSE 通信系统的核心组件：

- **Broker**：管理所有客户端连接和事件分发
- **EventBus**：处理事件的发布和订阅
- **ConnectionManager**：管理客户端连接的生命周期

#### 4.6.2 客户端管理

系统支持多种客户端连接管理功能：

- **客户端注册**：处理新客户端的连接请求
- **心跳机制**：通过定期发送 ping 事件保持连接活跃
- **超时处理**：自动关闭不活跃的客户端连接
- **资源清理**：确保客户端断开连接时释放资源

#### 4.6.3 事件类型

系统支持多种类型的事件：

- **节点事件**：节点创建、更新、删除等
- **状态事件**：节点状态变化、处理进度等
- **结果事件**：智能体处理结果、结论生成等
- **系统事件**：心跳、错误通知等

#### 4.6.4 事件广播

系统支持多种广播模式：

- **全局广播**：向所有连接的客户端发送事件
- **地图广播**：向特定思维导图的客户端发送事件
- **用户广播**：向特定用户的客户端发送事件
- **会话广播**：向特定会话的客户端发送事件

## 5. 数据模型

### 5.1 用户模型 (User)

```go
type User struct {
    ID        string         `gorm:"type:uuid;primaryKey"`
    Username  string         `gorm:"uniqueIndex;size:50;not null"`
    Email     string         `gorm:"uniqueIndex;size:100;not null"`
    Password  string         `gorm:"size:100;not null"`
    Role      string         `gorm:"size:20;default:'user'"`
    Status    string         `gorm:"size:20;default:'active'"`
    Metadata  datatypes.JSON `gorm:"type:jsonb;default:'{}'"`
    CreatedAt time.Time      `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
    UpdatedAt time.Time      `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
```

### 5.2 思维导图模型 (ThinkingMap)

```go
type ThinkingMap struct {
    SerialID    int64          `gorm:"primaryKey;autoIncrement;column:serial_id" json:"-"`
    ID          string         `gorm:"type:uuid;uniqueIndex"`
    UserID      string         `gorm:"type:uuid;not null;index"`
    Title       string         `gorm:"size:100;not null"`
    Description string         `gorm:"size:500"`
    Status      string         `gorm:"size:20;default:'active'"`
    Metadata    datatypes.JSON `gorm:"type:jsonb;default:'{}'"`
    CreatedAt   time.Time      `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
    UpdatedAt   time.Time      `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
    DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
```

### 5.3 节点模型 (ThinkingNode)

```go
type ThinkingNode struct {
    SerialID      int64            `gorm:"primaryKey;autoIncrement;column:serial_id" json:"-"`
    ID            string           `gorm:"type:uuid;uniqueIndex"`
    MapID         string           `gorm:"type:uuid;not null;index"`
    ParentID      string           `gorm:"type:uuid;index"`
    NodeType      string           `gorm:"type:varchar(50);not null"` // root, analysis, conclusion, custom
    Question      string           `gorm:"type:text;not null"`
    Target        string           `gorm:"type:text"`
    Context       DependentContext `gorm:"type:text;default:'{}'"` // 上下文
    Decomposition Decomposition    `gorm:"type:jsonb;default:'{}'"`
    Conclusion    Conclusion       `gorm:"type:jsonb;default:'{}'"`
    Status        string           `gorm:"type:varchar(50);default:'initial'"` // initial, pending, running, completed, error
    Position      Position         `gorm:"type:jsonb;default:'{\"x\":0,\"y\":0}'"`
    Metadata      datatypes.JSON   `gorm:"type:jsonb;default:'{}'"`
    Dependencies  Dependencies     `gorm:"type:jsonb;default:'[]'"`
    CreatedAt     time.Time        `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
    UpdatedAt     time.Time        `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
    DeletedAt     gorm.DeletedAt   `gorm:"index" json:"-"`
}
```

### 5.4 消息模型 (Message)

```go
type Message struct {
    SerialID  int64          `gorm:"primaryKey;autoIncrement;column:serial_id" json:"-"`
    ID        string         `gorm:"type:uuid;uniqueIndex"`
    NodeID    string         `gorm:"type:uuid;not null;index"`
    Role      string         `gorm:"type:varchar(20);not null"` // user, assistant, system
    Content   string         `gorm:"type:text;not null"`
    Type      string         `gorm:"type:varchar(20);default:'text'"` // text, image, file
    Metadata  datatypes.JSON `gorm:"type:jsonb;default:'{}'"`
    CreatedAt time.Time      `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
    UpdatedAt time.Time      `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
```

### 5.5 依赖上下文模型 (DependentContext)

```go
type DependentContext struct {
    Ancestors  []string `json:"ancestors,omitempty"`  // 祖先节点ID列表
    Siblings   []string `json:"siblings,omitempty"`   // 兄弟节点ID列表
    Children   []string `json:"children,omitempty"`   // 子节点ID列表
    References []string `json:"references,omitempty"` // 引用节点ID列表
}
```

### 5.6 分解模型 (Decomposition)

```go
type Decomposition struct {
    Strategy   string   `json:"strategy,omitempty"`   // 分解策略
    SubTasks   []string `json:"subTasks,omitempty"`   // 子任务列表
    Reasoning  string   `json:"reasoning,omitempty"`  // 推理过程
    Complexity string   `json:"complexity,omitempty"` // 复杂度评估
}
```

### 5.7 结论模型 (Conclusion)

```go
type Conclusion struct {
    Summary    string   `json:"summary,omitempty"`    // 结论摘要
    Content    string   `json:"content,omitempty"`    // 结论内容
    Evidence   []string `json:"evidence,omitempty"`   // 证据列表
    Confidence float64  `json:"confidence,omitempty"` // 置信度
    References []string `json:"references,omitempty"` // 引用列表
}
```

## 6. 部署指南

### 6.1 环境要求
- Go 1.24+
- PostgreSQL 14+
- Redis 7.0+

### 6.2 配置说明
系统使用环境变量和配置文件进行配置，主要配置项包括：

- 服务器配置：端口、运行模式
- 数据库配置：连接信息
- Redis配置：连接信息
- JWT配置：密钥、过期时间
- LLM配置：API密钥、模型选择
- SSE配置：心跳间隔、超时时间
- 日志配置：级别、文件路径

### 6.3 部署步骤

1. 克隆代码库
   ```bash
   git clone https://github.com/username/thinking-map.git
   cd thinking-map/server
   ```

2. 安装依赖
   ```bash
   go mod download
   ```

3. 配置环境变量
   ```bash
   cp .env.example .env
   # 编辑.env文件，填入必要的配置
   ```

4. 初始化数据库
   ```bash
   psql -U postgres -f scripts/db.sql
   ```

5. 构建应用
   ```bash
   go build -o thinking-map-server ./cmd/server
   ```

6. 运行服务
   ```bash
   ./thinking-map-server
   ```

## 7. 开发指南

### 7.1 开发环境设置
1. 安装Go 1.24+
2. 安装PostgreSQL 14+
3. 安装Redis 7.0+
4. 配置开发环境变量

### 7.2 代码规范
- 遵循Go标准代码规范
- 使用gofmt格式化代码
- 编写单元测试和集成测试
- 使用有意义的变量和函数名
- 遵循适当的错误处理模式

### 7.3 测试
- 单元测试：`go test ./internal/service`
- 集成测试：`go test ./internal/...`
- 测试覆盖率：`go test -cover ./internal/...`

### 7.4 常见问题
1. **数据库连接问题**
   - 检查数据库配置
   - 确保PostgreSQL服务正在运行

2. **Redis连接问题**
   - 检查Redis配置
   - 确保Redis服务正在运行

3. **JWT认证问题**
   - 检查JWT密钥配置
   - 确保令牌未过期

4. **LLM API问题**
   - 检查API密钥配置
   - 确保网络连接正常

## 8. 未来规划

### 8.1 短期计划
- 优化智能体性能
- 增强实时通信稳定性
- 改进用户体验

### 8.2 中期计划
- 添加更多智能体类型
- 实现知识库集成
- 支持团队协作

### 8.3 长期计划
- 多语言支持
- 移动端API
- 高级分析功能