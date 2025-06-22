## 1. 项目概述
### 1.1 项目简介
ThinkingMap 后端服务是一个智能化问题解决可视化系统，提供思维导图节点管理、AI 问题拆解、结论生成等核心功能。系统采用微服务架构思想，支持实时通信和智能体交互。

### 1.2 核心功能
+ 思维导图节点的 CRUD 操作
+ 基于 AI 的问题智能拆解
+ 实时结论生成和推理
+ SSE 长连接实时通信
+ 用户认证和权限管理
+ 会话状态管理

### 1.3 性能目标
+ 支持并发用户数：1000+
+ API 响应时间：< 200ms
+ SSE 连接稳定性：99.9%
+ 数据库查询响应：< 100ms

## 2. 技术选型
### 2.1 核心技术栈
| 技术组件 | 选型 | 版本要求 | 选型理由 |
| --- | --- | --- | --- |
| 编程语言 | Golang | 1.21+ | 高性能、并发友好、生态成熟 |
| HTTP框架 | cloudwego/hertz | v0.7.0+ | 高性能、中间件丰富、字节跳动开源 |
| 智能体框架 | cloudwego/eino | v0.1.0+ | 专业AI Agent框架，支持复杂推理链 |
| 数据库 | PostgreSQL | 14+ | JSONB支持、事务完整性、扩展性好 |
| 缓存 | Redis | 7.0+ | 高性能缓存、发布订阅、数据结构丰富 |
| ORM | GORM | v1.25+ | 功能完善、社区活跃、PostgreSQL支持好 |
| 认证 | JWT | - | 无状态、跨域友好、标准化 |
| 配置管理 | Viper | v1.16+ | 多格式支持、环境变量集成 |
| 日志 | Zap | v1.25+ | 高性能、结构化日志、字段丰富 |
| API文档 | Swagger | v1.16+ | 标准化、自动生成、调试友好 |


### 2.2 依赖服务
+ **PostgreSQL**: 主数据存储
+ **Redis**: 缓存和会话存储
+ **AI服务**: 大语言模型接口（OpenAI/Claude等）

## 3. 系统架构
## 3. 系统架构

### 3.1 整体架构图
```
┌─────────────────────────────────────────────────────────────┐
│                      Web Client (PC)                       │
│                    React + TypeScript                      │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      │ HTTPS/WSS
                      │
┌─────────────────────┴───────────────────────────────────────┐
│                    API Gateway                             │
│              (Authentication & Routing)                    │
└─────────────────────┬───────────────────────────────────────┘
                      │
          ┌───────────┼───────────┐
          │           │           │
┌─────────┴─────────┐ │ ┌─────────┴─────────┐ ┌─────────────────┐
│   Node Service    │ │ │  Thinking Service │ │   User Service  │
│                   │ │ │                   │ │                 │
│ • Node CRUD       │ │ │ • Session Manage  │ │ • Authentication│
│ • Graph Manage    │ │ │ • Context Handle  │ │ • User Profile  │
│ • SSE Broadcast   │ │ │ • Agent Orchestr  │ │ • Permission    │
│ • Dependency Chk  │ │ │ • Result Process  │ │ • Session Mgmt  │
└─────────┬─────────┘ │ └─────────┬─────────┘ └─────────┬───────┘
          │           │           │                     │
          └───────────┼───────────┼─────────────────────┘
                      │           │
                      │           │
┌─────────────────────┴───────────┴───────────────────────────┐
│                      Agent Layer                           │
│                                                            │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────┐ │
│  │ Decompose Agent │  │ Analysis Agent  │  │ Intent Agent│ │
│  │                 │  │                 │  │             │ │
│  │ • Problem Split │  │ • RAG Query     │  │ • User Intent│ │
│  │ • Sub-task Gen  │  │ • Knowledge Ret │  │ • Context Ana│ │
│  │ • Dependency    │  │ • Conclusion    │  │ • Action Rec │ │
│  │ • Strategy Plan │  │ • Evidence Syn  │  │ • Flow Ctrl  │ │
│  │ • Conclusion   │  │ • Conclusion    │  │ • Conclusion │ │
│  │ • Strategy Plan │  │ • Evidence Syn  │  │ • Conclusion │ │
│  │ • Conclusion   │  │ • Conclusion    │  │ • Conclusion │ │
│  └─────────────────┘  └─────────────────┘  └─────────────┘ │
│                                                            │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────┐ │
│  │ Reasoning Agent │  │ Validation Agent│  │ Context Mgr │ │
│  │                 │  │                 │  │             │ │
│  │ • Logic Chain   │  │ • Result Check  │  │ • Memory Mgmt│ │
│  │ • Inference     │  │ • Quality Assur │  │ • State Track│ │
│  │ • Deduction     │  │ • Error Detect  │  │ • History   │ │
│  │ • Hypothesis    │  │ • Consistency   │  │ • Context   │ │
│  └─────────────────┘  └─────────────────┘  └─────────────┘ │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────┴───────────────────────────────────────┐
│                      Data Layer                            │
│                                                            │
│  ┌─────────────────────────────────────────────────────┐   │
│  │                  PostgreSQL                         │   │
│  │                                                     │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  │   │
│  │  │    Users    │  │ Thinking    │  │   Nodes     │  │   │
│  │  │             │  │    Maps     │  │             │  │   │
│  │  │ • Profile   │  │ • Metadata  │  │ • Content   │  │   │
│  │  │ • Auth Info │  │ • Status    │  │ • Relations │  │   │
│  │  │ • Sessions  │  │ • History   │  │ • Position  │  │   │
│  │  └─────────────┘  └─────────────┘  └─────────────┘  │   │
│  │                                                     │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  │   │
│  │  │  Sessions   │  │   History   │  │  Knowledge  │  │   │
│  │  │             │  │             │  │             │  │   │
│  │  │ • Tokens    │  │ • Operations│  │ • Embeddings│  │   │
│  │  │ • Activity  │  │ • AI Logs   │  │ • Documents │  │   │
│  │  │ • Context   │  │ • Reasoning │  │ • Vectors   │  │   │
│  │  └─────────────┘  └─────────────┘  └─────────────┘  │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                            │
│  ┌─────────────────────────────────────────────────────┐   │
│  │                     Redis                           │   │
│  │                                                     │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  │   │
│  │  │    Cache    │  │ SSE Channels│  │ Rate Limit  │  │   │
│  │  │             │  │             │  │             │  │   │
│  │  │ • Node Data │  │ • User Conn │  │ • API Limit │  │   │
│  │  │ • User Info │  │ • Event Que │  │ • User Limit│  │   │
│  │  │ • Sessions  │  │ • Broadcast │  │ • IP Limit  │  │   │
│  │  └─────────────┘  └─────────────┘  └─────────────┘  │   │
│  │                                                     │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  │   │
│  │  │Agent State  │  │ Temp Data   │  │ Pub/Sub     │  │   │
│  │  │             │  │             │  │             │  │   │
│  │  │ • Execution │  │ • Processing│  │ • Events    │  │   │
│  │  │ • Context   │  │ • Buffers   │  │ • Messages  │  │   │
│  │  │ • Memory    │  │ • Cache     │  │ • Broadcast │  │   │
│  │  └─────────────┘  └─────────────┘  └─────────────┘  │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

### 3.2 服务分层架构
```
┌─────────────────────────────────────────────────────────────┐
│                  Presentation Layer                         │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │ REST API    │  │ SSE Handler │  │  GraphQL Endpoint   │  │
│  │             │  │             │  │     (Optional)      │  │
│  │ • CRUD Ops  │  │ • Real-time │  │ • Complex Queries   │  │
│  │ • Auth      │  │ • Events    │  │ • Batch Operations  │  │
│  │ • Validation│  │ • Broadcast │  │ • Subscriptions     │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│                    Business Layer                           │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │ Node Service│  │Think Service│  │   User Service      │  │
│  │             │  │             │  │                     │  │
│  │ • Graph Mgmt│  │ • Agent Orch│  │ • Authentication    │  │
│  │ • CRUD Logic│  │ • Session   │  │ • Authorization     │  │
│  │ • Validation│  │ • Context   │  │ • Profile Mgmt      │  │
│  │ • Events    │  │ • Workflow  │  │ • Session Control   │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│                      Agent Layer                            │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │ Cognitive   │  │ Execution   │  │   Support           │  │
│  │ Agents      │  │ Agents      │  │   Agents            │  │
│  │             │  │             │  │                     │  │
│  │ • Decompose │  │ • Analysis  │  │ • Intent Recognition│  │
│  │ • Reasoning │  │ • Synthesis │  │ • Context Manager   │  │
│  │ • Planning  │  │ • Validation│  │ • Quality Assurance │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│                       Data Layer                            │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │ Repository  │  │    Cache    │  │   Message Queue     │  │
│  │             │  │             │  │                     │  │
│  │ • PostgreSQL│  │ • Redis     │  │ • Event Streaming   │  │
│  │ • ORM       │  │ • Memory    │  │ • Pub/Sub           │  │
│  │ • Migrations│  │ • Sessions  │  │ • Task Queue        │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### 3.3 Agent层详细设计

#### 3.3.1 Agent分类和职责

**认知类Agent (Cognitive Agents)**
```
┌─────────────────────────────────────────────────────────────┐
│                    Decompose Agent                          │
│  职责: 问题分解和任务规划                                    │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────┐  │
│  │ Problem Analysis│  │ Strategy Planning│  │ Task Split  │  │
│  │ • 问题理解       │  │ • 分解策略选择   │  │ • 子任务生成│  │
│  │ • 复杂度评估     │  │ • 依赖关系分析   │  │ • 优先级排序│  │
│  │ • 领域识别       │  │ • 执行路径规划   │  │ • 资源分配  │  │
│  └─────────────────┘  └─────────────────┘  └─────────────┘  │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                    Reasoning Agent                          │
│  职责: 逻辑推理和知识推导                                    │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────┐  │
│  │ Logic Reasoning │  │ Knowledge Infer │  │ Hypothesis  │  │
│  │ • 演绎推理       │  │ • 知识图谱推理   │  │ • 假设生成  │  │
│  │ • 归纳推理       │  │ • 规则应用       │  │ • 验证设计  │  │
│  │ • 类比推理       │  │ • 因果分析       │  │ • 结论导出  │  │
│  └─────────────────┘  └─────────────────┘  └─────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

**执行类Agent (Execution Agents)**
```
┌─────────────────────────────────────────────────────────────┐
│                    Analysis Agent                           │
│  职责: 数据分析和信息处理                                    │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────┐  │
│  │ RAG Query       │  │ Knowledge Ret   │  │ Data Process│  │
│  │ • 向量检索       │  │ • 文档召回       │  │ • 信息提取  │  │
│  │ • 语义搜索       │  │ • 知识融合       │  │ • 数据清洗  │  │
│  │ • 相关性排序     │  │ • 上下文构建     │  │ • 结构化    │  │
│  └─────────────────┘  └─────────────────┘  └─────────────┘  │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                   Synthesis Agent                           │
│  职责: 信息综合和结论生成                                    │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────┐  │
│  │ Evidence Fusion │  │ Conclusion Gen  │  │ Report Gen  │  │
│  │ • 证据整合       │  │ • 结论推导       │  │ • 报告生成  │  │
│  │ • 冲突解决       │  │ • 置信度评估     │  │ • 格式化    │  │
│  │ • 权重分配       │  │ • 不确定性处理   │  │ • 可视化    │  │
│  └─────────────────┘  └─────────────────┘  └─────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

**支持类Agent (Support Agents)**
```
┌─────────────────────────────────────────────────────────────┐
│                    Intent Agent                             │
│  职责: 用户意图识别和交互管理                                │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────┐  │
│  │ Intent Classify │  │ Context Analysis│  │ Action Plan │  │
│  │ • 意图分类       │  │ • 上下文理解     │  │ • 动作规划  │  │
│  │ • 实体识别       │  │ • 对话状态跟踪   │  │ • 响应生成  │  │
│  │ • 情感分析       │  │ • 历史关联       │  │ • 流程控制  │  │
│  └─────────────────┘  └─────────────────┘  └─────────────┘  │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                   Context Manager                           │
│  职责: 上下文管理和状态维护                                  │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────┐  │
│  │ Memory Mgmt     │  │ State Tracking  │  │ History Mgmt│  │
│  │ • 工作记忆       │  │ • 执行状态       │  │ • 操作历史  │  │
│  │ • 长期记忆       │  │ • 上下文切换     │  │ • 版本控制  │  │
│  │ • 记忆检索       │  │ • 状态恢复       │  │ • 回滚支持  │  │
│  └─────────────────┘  └─────────────────┘  └─────────────┘  │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                  Validation Agent                           │
│  职责: 质量保证和结果验证                                    │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────┐  │
│  │ Quality Check   │  │ Consistency Val │  │ Error Detect│  │
│  │ • 结果质量评估   │  │ • 逻辑一致性     │  │ • 错误检测  │  │
│  │ • 完整性检查     │  │ • 数据一致性     │  │ • 异常处理  │  │
│  │ • 准确性验证     │  │ • 语义一致性     │  │ • 修复建议  │  │
│  └─────────────────┘  └─────────────────┘  └─────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### 3.4 数据流设计

#### 3.4.1 问题拆解流程
```
User Request → Intent Agent → Decompose Agent → Analysis Agent → Node Service
     ↓              ↓              ↓               ↓              ↓
Context Mgr ← Validation ← Reasoning ← Knowledge ← SSE Broadcast
     ↓         Agent        Agent       Retrieval       ↓
PostgreSQL ←                                      → Web Client
     ↑                                                   ↑
Redis Cache ←─────────────────────────────────────────────┘
```

#### 3.4.2 结论生成流程
```
Node Data → Context Manager → Analysis Agent → Reasoning Agent
    ↓             ↓                ↓               ↓
Evidence ← Memory Retrieval ← RAG Query ← Logic Inference
    ↓             ↓                ↓               ↓
Synthesis Agent ← Validation ← Quality Check ← Conclusion
    ↓             Agent           ↓               ↓
Result Storage → SSE Event → Web Client ← User Feedback
```

### 3.5 核心组件说明

**API Gateway**
- 统一入口和路由分发
- JWT认证和授权验证
- 请求限流和熔断保护
- 请求日志和监控统计

**Service Layer**
- **Node Service**: 节点生命周期管理、图结构维护、依赖关系检查
- **Thinking Service**: Agent编排、会话管理、执行流程控制
- **User Service**: 用户认证授权、配置管理、权限控制

**Agent Layer**
- 基于cloudwego/eino框架构建
- 支持Agent间消息传递和协作
- 提供统一的执行上下文和状态管理
- 实现可插拔的Agent扩展机制

**Data Layer**
- **PostgreSQL**: 结构化数据持久化存储
- **Redis**: 缓存、会话状态、实时通信支持
- 支持读写分离和数据分片扩展

## 4. 项目目录结构
```plain
server/
├── cmd/
│   └── server/
│       └── main.go                 # 应用入口
├── internal/
│   ├── config/
│   │   ├── config.go              # 配置结构定义
│   │   └── loader.go              # 配置加载逻辑
│   ├── handler/
│   │   ├── auth.go                # 认证相关接口
│   │   ├── node.go                # 节点相关接口
│   │   ├── thinking.go            # 思考相关接口
│   │   ├── sse.go                 # SSE处理器
│   │   └── middleware/
│   │       ├── auth.go            # 认证中间件
│   │       ├── cors.go            # 跨域中间件
│   │       ├── logger.go          # 日志中间件
│   │       └── ratelimit.go       # 限流中间件
│   ├── service/
│   │   ├── auth.go                # 认证服务
│   │   ├── node.go                # 节点服务
│   │   ├── thinking.go            # 思考服务
│   │   ├── sse.go                 # SSE服务
│   │   └── ai/
│   │       ├── agent.go           # AI智能体
│   │       ├── rag.go             # RAG检索
│   │       └── llm.go             # 大语言模型接口
│   ├── repository/
│   │   ├── interfaces.go          # 仓储接口定义
│   │   ├── node.go                # 节点数据访问
│   │   ├── user.go                # 用户数据访问
│   ├── model/
│   │   ├── node.go                # 节点数据模型
│   │   ├── user.go                # 用户数据模型
│   │   ├── session.go             # 会话数据模型
│   │   └── dto/
│   │       ├── request.go         # 请求DTO
│   │       └── response.go        # 响应DTO
│   ├── pkg/
│   │   ├── database/
│   │   │   ├── postgres.go        # PostgreSQL连接
│   │   │   └── redis.go           # Redis连接
│   │   ├── logger/
│   │   │   └── zap.go             # 日志配置
│   │   ├── jwt/
│   │   │   └── token.go           # JWT处理
│   │   ├── validator/
│   │   │   └── validator.go       # 参数验证
│   │   └── utils/
│   │       ├── crypto.go          # 加密工具
│   │       ├── time.go            # 时间工具
│   │       └── uuid.go            # UUID生成
│   └── router/
│       └── router.go              # 路由配置
├── api/
│   └── swagger/
│       ├── docs.go                # Swagger文档
│       ├── swagger.json           # API规范
│       └── swagger.yaml           # API规范
├── scripts/
│   ├── migration/
│   │   ├── 001_init.sql          # 数据库初始化
│   │   ├── 002_nodes.sql         # 节点表创建
│   │   └── 003_indexes.sql       # 索引创建
│   ├── docker/
│   │   ├── Dockerfile            # 应用镜像
│   │   └── docker-compose.yml    # 开发环境
│   └── deploy/
│       ├── k8s/                  # Kubernetes配置
│       └── helm/                 # Helm Charts
├── test/
│   ├── integration/              # 集成测试
│   ├── unit/                     # 单元测试
│   └── fixtures/                 # 测试数据
├── configs/
│   ├── config.yaml               # 默认配置
│   ├── config.dev.yaml           # 开发环境配置
│   └── config.prod.yaml          # 生产环境配置
├── docs/
│   ├── api.md                    # API文档
│   ├── deployment.md             # 部署文档
│   └── development.md            # 开发文档
├── .env.example                  # 环境变量示例
├── .gitignore                    # Git忽略文件
├── go.mod                        # Go模块定义
├── go.sum                        # Go模块校验
├── Makefile                      # 构建脚本
└── README.md                     # 项目说明
```

## 5. 数据库设计
### 5.1 数据库表结构
#### 5.1.1 用户表 (users)
```sql
CREATE TABLE "users" (
    id UUID PRIMARY KEY,
    username VARCHAR(32) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    full_name VARCHAR(100) NOT NULL,
    status SMALLINT NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);
CREATE INDEX idx_users_deleted_at ON "users"(deleted_at);
```

#### 5.1.2 思维导图表 (thinking_maps)
```sql
CREATE TABLE "thinking_maps" (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    problem TEXT NOT NULL,
    problem_type VARCHAR(50),
    target TEXT,
    key_points JSONB,
    constraints JSONB,
    conclusion TEXT,
    status INT NOT NULL DEFAULT 1,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);
CREATE INDEX idx_thinking_maps_user_id ON "thinking_maps"(user_id);
CREATE INDEX idx_thinking_maps_deleted_at ON "thinking_maps"(deleted_at);
```

#### 5.1.3 节点表 (thinking_nodes)
```sql
CREATE TABLE "thinking_nodes" (
    id UUID PRIMARY KEY,
    map_id UUID NOT NULL,
    parent_id UUID,
    node_type VARCHAR(50) NOT NULL,
    question TEXT NOT NULL,
    target TEXT,
    context TEXT DEFAULT '[]',
    conclusion TEXT,
    status INT DEFAULT 0,
    position JSONB DEFAULT '{"x":0,"y":0}',
    metadata JSONB DEFAULT '{}',
    dependencies JSONB DEFAULT '[]',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);
CREATE INDEX idx_thinking_nodes_map_id ON "thinking_nodes"(map_id);
CREATE INDEX idx_thinking_nodes_parent_id ON "thinking_nodes"(parent_id);
CREATE INDEX idx_thinking_nodes_deleted_at ON "thinking_nodes"(deleted_at);
```
> position 字段结构
```json
{
  "x": 100,
  "y": 200,
  "width": 300,
  "height": 150
}
```
> dependencies 字段结构
```json
[
  {
    "node_id": "uuid",
    "dependency_type": "prerequisite",
    "required": true
  }
]
```

#### 5.1.4 节点详情表 (node_details)
```sql
CREATE TABLE "node_details" (
    id UUID PRIMARY KEY,
    node_id UUID NOT NULL,
    detail_type VARCHAR(50) NOT NULL,
    content JSONB NOT NULL DEFAULT '{}',
    status INT NOT NULL DEFAULT 1,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);
CREATE INDEX idx_node_details_node_id ON "node_details"(node_id);
CREATE INDEX idx_node_details_deleted_at ON "node_details"(deleted_at);
```
> tab_content 字段结构
// info tab
```json
{
  "context": [
    {
      "type": "父节点问题",
      "question": "",
      "target": "",
      "conclusion": ""
    },{
      "type": "子节点问题",
      "question": "",
      "target": "",
      "conclusion": ""
    },
  ],
  "question": "",
  "target": ""
}
```
// decompose tab
```json
{
  "message": [
    [1,2,3],
    [4]
  ],
  "decompose_result": [
    {
      "question": "",
      "target": ""
    }
  ]
}
```
// conclusion tab
```json
{
  "message": [
    [1,2,3],
    [4]
  ],
  "conclusion": ""
}
```

#### 5.1.5 消息表 (messages)
```sql
CREATE TABLE "messages" (
    id UUID PRIMARY KEY,
    node_id UUID NOT NULL,
    parent_id UUID,
    message_type VARCHAR(20) NOT NULL DEFAULT '1',
    content JSONB NOT NULL,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);
CREATE INDEX idx_messages_node_id ON "messages"(node_id);
CREATE INDEX idx_messages_parent_id ON "messages"(parent_id);
CREATE INDEX idx_messages_deleted_at ON "messages"(deleted_at);
```
> content 字段结构
```json
{
  "text": "",
  "rag": ["rag_id"],
  "notice": [
    {
      "type": "",
      "content": ""
    }
  ]
}
```
#### 5.1.6 RAG检索记录表 (rag_record)
```sql
CREATE TABLE "rag_records" (
    id UUID PRIMARY KEY,
    query TEXT NOT NULL,
    answer TEXT NOT NULL,
    sources JSONB NOT NULL DEFAULT '[]',
    status INT NOT NULL DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);
CREATE INDEX idx_rag_records_deleted_at ON "rag_records"(deleted_at);
```

## 6. API设计
### 6.1 API 规范
+ **协议**: HTTP/1.1, HTTP/2
+ **数据格式**: JSON
+ **认证方式**: Bearer Token (JWT)
+ **版本控制**: URL路径版本 `/api/v1/`
+ **错误处理**: 标准HTTP状态码 + 错误详情

### 6.2 通用响应格式
```json
{
  "code": 200,
  "message": "success",
  "data": {},
  "timestamp": "2024-01-01T00:00:00Z",
  "request_id": "uuid"
}
```

### 6.3 API 接口设计
#### 6.3.1 认证相关接口
```yaml
# 用户注册
POST /api/v1/auth/register
Content-Type: application/json

Request:
{
  "username": "string",
  "email": "string", 
  "password": "string",
  "full_name": "string"
}

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": {
    "user_id": "uuid",
    "username": "string",
    "email": "string",
    "full_name": "string",
    "access_token": "string",
    "refresh_token": "string",
    "expires_in": 900
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "request_id": "uuid"
}

Response 400 Bad Request:
{
  "code": 400,
  "message": "invalid request parameters",
  "data": {
    "field": "username",
    "error": "username already exists"
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "request_id": "uuid"
}

# 用户登录
POST /api/v1/auth/login
Content-Type: application/json

Request:
{
  "username": "string",
  "password": "string"
}

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": {
    "user_id": "uuid",
    "username": "string",
    "email": "string",
    "full_name": "string",
    "access_token": "string",
    "refresh_token": "string",
    "expires_in": 900
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "request_id": "uuid"
}

Response 401 Unauthorized:
{
  "code": 401,
  "message": "invalid credentials",
  "data": null,
  "timestamp": "2024-01-01T00:00:00Z",
  "request_id": "uuid"
}

# 刷新Token
POST /api/v1/auth/refresh
Authorization: Bearer <refresh_token>

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": {
    "access_token": "string",
    "refresh_token": "string",
    "expires_in": 900
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "request_id": "uuid"
}

Response 401 Unauthorized:
{
  "code": 401,
  "message": "invalid refresh token",
  "data": null,
  "timestamp": "2024-01-01T00:00:00Z",
  "request_id": "uuid"
}

# 登出
POST /api/v1/auth/logout
Authorization: Bearer <access_token>

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": null,
  "timestamp": "2024-01-01T00:00:00Z",
  "request_id": "uuid"
}
```

#### 6.3.2 思考图接口
```yaml
# 创建思维导图
POST /api/v1/maps
Authorization: Bearer <token>
Content-Type: application/json

Request:
{
  "title": "string",
  "description": "string",
  "root_question": "string"
}

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "uuid",
    "title": "string",
    "description": "string",
    "root_question": "string",
    "root_node_id": "uuid",
    "status": 1,
    "metadata": {},
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "request_id": "uuid"
}

# 获取用户的思维导图列表
GET /api/v1/maps?page=1&limit=20&status=1
Authorization: Bearer <token>

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": {
    "total": 100,
    "page": 1,
    "limit": 20,
    "items": [
      {
        "id": "uuid",
        "title": "string",
        "description": "string",
        "root_question": "string",
        "status": 1,
        "node_count": 10,
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z"
      }
    ]
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "request_id": "uuid"
}

# 获取思维导图详情
GET /api/v1/maps/{mapId}
Authorization: Bearer <token>

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "uuid",
    "title": "string",
    "description": "string",
    "root_question": "string",
    "root_node_id": "uuid",
    "status": 1,
    "metadata": {},
    "node_count": 10,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "request_id": "uuid"
}

# 更新思维导图
PUT /api/v1/maps/{mapId}
Authorization: Bearer <token>
Content-Type: application/json

Request:
{
  "title": "string",
  "description": "string",
  "status": 1
}

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "uuid",
    "title": "string",
    "description": "string",
    "status": 1,
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "request_id": "uuid"
}

# 删除思维导图
DELETE /api/v1/maps/{mapId}
Authorization: Bearer <token>

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": null,
  "timestamp": "2024-01-01T00:00:00Z",
  "request_id": "uuid"
}
```

#### 6.3.3 节点管理接口
```yaml
# 获取思维导图的所有节点
GET /api/v1/maps/{mapId}/nodes
Authorization: Bearer <token>

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": {
    "nodes": [
      {
        "id": "uuid",
        "parent_id": "uuid",
        "node_type": "analysis",
        "question": "string",
        "target": "string",
        "status": 0,
        "position": {
          "x": 100,
          "y": 200
        }
      }
    ]
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "request_id": "uuid"
}

# 创建节点
POST /api/v1/maps/{mapId}/nodes
Authorization: Bearer <token>
Content-Type: application/json

Request:
{
  "parent_id": "uuid",
  "node_type": "analysis",
  "question": "string",
  "target": "string",
  "context": "string",
  "position": {"x": 100, "y": 200}
}

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "uuid",
    "map_id": "uuid",
    "parent_id": "uuid",
    "node_type": "analysis",
    "question": "string",
    "target": "string",
    "context": "string",
    "status": 0,
    "position": {
      "x": 100,
      "y": 200
    },
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "request_id": "uuid"
}

# 更新节点
PUT /api/v1/nodes/{nodeId}
Authorization: Bearer <token>
Content-Type: application/json

Request:
{
  "question": "string",
  "target": "string",
  "context": "string",
  "position": {"x": 100, "y": 200}
}

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "uuid",
    "question": "string",
    "target": "string",
    "context": "string",
    "position": {
      "x": 100,
      "y": 200
    },
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "request_id": "uuid"
}

# 删除节点
DELETE /api/v1/nodes/{nodeId}
Authorization: Bearer <token>

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": null,
  "timestamp": "2024-01-01T00:00:00Z",
  "request_id": "uuid"
}

# 检查节点依赖
GET /api/v1/nodes/{nodeId}/dependencies
Authorization: Bearer <token>

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": {
    "dependencies": [
      {
        "node_id": "uuid",
        "dependency_type": "prerequisite",
        "required": true,
        "status": 2
      }
    ],
    "dependent_nodes": [
      {
        "node_id": "uuid",
        "dependency_type": "dependent",
        "required": true,
        "status": 0
      }
    ]
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "request_id": "uuid"
}
```

#### 6.3.4 AI思考接口
```yaml
# 开始问题分析
POST /api/v1/thinking/analyze
Authorization: Bearer <token>
Content-Type: application/json

Request:
{
  "node_id": "uuid",
  "context": "string",
  "options": {
    "model": "gpt-4",
    "temperature": 0.7
  }
}

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": {
    "task_id": "uuid",
    "node_id": "uuid",
    "status": "processing",
    "estimated_time": 30
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "request_id": "uuid"
}

# 开始问题拆解
POST /api/v1/thinking/decompose
Authorization: Bearer <token>
Content-Type: application/json

Request:
{
  "node_id": "uuid",
  "decompose_strategy": "breadth_first",
  "max_depth": 3
}

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": {
    "task_id": "uuid",
    "node_id": "uuid",
    "status": "processing",
    "estimated_time": 60
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "request_id": "uuid"
}

# 生成结论
POST /api/v1/thinking/conclude
Authorization: Bearer <token>
Content-Type: application/json

Request:
{
  "node_id": "uuid",
  "evidence": ["string"],
  "reasoning_type": "deductive"
}

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": {
    "task_id": "uuid",
    "node_id": "uuid",
    "status": "processing",
    "estimated_time": 45
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "request_id": "uuid"
}

# 对话交互
POST /api/v1/thinking/chat
Authorization: Bearer <token>
Content-Type: application/json

Request:
{
  "node_id": "uuid",
  "message": "string",
  "context": "decompose" // decompose | conclude
}

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": {
    "message_id": "uuid",
    "content": "string",
    "role": "assistant",
    "created_at": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "request_id": "uuid"
}
```

#### 6.3.5 SSE接口
```yaml
# 建立SSE连接
GET /api/v1/sse/connect/{mapId}
Authorization: Bearer <token>
Accept: text/event-stream
Cache-Control: no-cache
Connection: keep-alive

# 连接成功响应
event: connected
data: {
  "connection_id": "uuid",
  "map_id": "uuid",
  "timestamp": "2024-01-01T00:00:00Z"
}

# 连接关闭响应
event: disconnected
data: {
  "connection_id": "uuid",
  "reason": "user_disconnected",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 6.4 SSE事件格式
```javascript
// 节点创建事件
event: node_created
data: {
  "node_id": "uuid",
  "parent_id": "uuid", 
  "node_type": "analysis",
  "question": "string",
  "target": "string",
  "position": {"x": 100, "y": 200},
  "timestamp": "2024-01-01T00:00:00Z"
}

// 节点更新事件
event: node_updated
data: {
  "node_id": "uuid",
  "updates": {
    "status": 2,
    "conclusion": "string"
  },
  "timestamp": "2024-01-01T00:00:00Z"
}

// 思考进度事件
event: thinking_progress
data: {
  "node_id": "uuid",
  "stage": "analyzing",
  "progress": 50,
  "message": "正在分析问题...",
  "timestamp": "2024-01-01T00:00:00Z"
}

// 错误事件
event: error
data: {
  "node_id": "uuid",
  "error_code": "THINKING_FAILED",
  "error_message": "AI服务暂时不可用",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

## 7. 安全设计
### 7.1 认证授权机制
#### 7.1.1 JWT Token设计
```go
type Claims struct {
    UserID    string `json:"user_id"`
    Username  string `json:"username"`
    Email     string `json:"email"`
    Role      string `json:"role"`
    SessionID string `json:"session_id"`
    jwt.RegisteredClaims
}

// Token配置
const (
    AccessTokenExpiry  = 15 * time.Minute
    RefreshTokenExpiry = 7 * 24 * time.Hour
    TokenIssuer        = "thinkingmap"
)
```

#### 7.1.2 权限控制
```go
// 权限级别
type Permission int

const (
    PermissionRead Permission = iota
    PermissionWrite
    PermissionDelete
    PermissionAdmin
)

// 资源类型
type ResourceType string

const (
    ResourceMap  ResourceType = "map"
    ResourceNode ResourceType = "node"
    ResourceUser ResourceType = "user"
)
```

### 7.2 数据安全
#### 7.2.1 密码安全
```go
import "golang.org/x/crypto/bcrypt"

// 密码加密
func HashPassword(password string) (string, error) {
    cost := 12 // bcrypt cost
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
    return string(bytes), err
}

// 密码验证
func CheckPassword(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}
```

#### 7.2.2 敏感数据加密
```go
// 使用AES-256-GCM加密敏感字段
type EncryptionService interface {
    Encrypt(plaintext string) (string, error)
    Decrypt(ciphertext string) (string, error)
}
```

### 7.3 API安全
#### 7.3.1 限流策略
```go
// 限流配置
type RateLimitConfig struct {
    // 每分钟请求数限制
    RequestsPerMinute int
    // 突发请求数限制  
    BurstSize int
    // 限流键生成策略
    KeyGenerator func(*hertz.RequestContext) string
}

// 不同接口的限流策略
var RateLimits = map[string]RateLimitConfig{
    "/api/v1/auth/login":    {RequestsPerMinute: 5, BurstSize: 10},
    "/api/v1/thinking/*":    {RequestsPerMinute: 30, BurstSize: 50},
    "/api/v1/nodes/*":       {RequestsPerMinute: 100, BurstSize: 200},
}
```

#### 7.3.2 输入验证
```go
import "github.com/go-playground/validator/v10"

// 请求参数验证
type CreateNodeRequest struct {
    ParentID string `json:"parent_id" validate:"omitempty,uuid"`
    NodeType string `json:"node_type" validate:"required,oneof=root analysis conclusion custom"`
    Question string `json:"question" validate:"required,min=1,max=1000"`
    Target   string `json:"target" validate:"max=500"`
    Context  string `json:"context" validate:"max=5000"`
}
```

### 7.4 传输安全
+ **HTTPS**: 强制使用TLS 1.2+
+ **HSTS**: HTTP严格传输安全
+ **CORS**: 跨域资源共享配置
+ **CSP**: 内容安全策略

### 7.5 会话安全
```go
// 会话管理
type SessionManager interface {
    CreateSession(userID string) (*Session, error)
    ValidateSession(token string) (*Session, error)
    RefreshSession(refreshToken string) (*Session, error)
    RevokeSession(sessionID string) error
    CleanExpiredSessions() error
}

// 会话安全策略
type SessionConfig struct {
    MaxConcurrentSessions int           // 最大并发会话数
    SessionTimeout        time.Duration // 会话超时时间
    IdleTimeout          time.Duration // 空闲超时时间
    RequireIPValidation  bool          // 是否验证IP
    RequireUAValidation  bool          // 是否验证User-Agent
}
```

## 8. 日志设计
### 8.1 日志级别和分类
```go
import "go.uber.org/zap"

// 日志级别
const (
    LogLevelDebug = "debug"
    LogLevelInfo  = "info" 
    LogLevelWarn  = "warn"
    LogLevelError = "error"
    LogLevelFatal = "fatal"
)

// 日志分类
const (
    LogCategoryAPI      = "api"
    LogCategoryDB       = "database"
    LogCategoryAuth     = "auth"
    LogCategoryThinking = "thinking"
    LogCategorySSE      = "sse"
    LogCategorySystem   = "system"
)
```

### 8.2 日志格式设计
```go
// 结构化日志字段
type LogFields struct {
    RequestID   string        `json:"request_id"`
    UserID      string        `json:"user_id,omitempty"`
    Method      string        `json:"method,omitempty"`
    Path        string        `json:"path,omitempty"`
    StatusCode  int           `json:"status_code,omitempty"`
    Duration    time.Duration `json:"duration,omitempty"`
    IP          string        `json:"ip,omitempty"`
    UserAgent   string        `json:"user_agent,omitempty"`
    Category    string        `json:"category"`
    Component   string        `json:"component"`
    Action      string        `json:"action,omitempty"`
    Resource    string        `json:"resource,omitempty"`
    Error       string        `json:"error,omitempty"`
    Metadata    interface{}   `json:"metadata,omitempty"`
}
```

### 8.3 日志配置
```go
// 日志配置结构
type LogConfig struct {
    Level      string `mapstructure:"level"`
    Format     string `mapstructure:"format"`     // json | console
    Output     string `mapstructure:"output"`     // stdout | file | both
    FilePath   string `mapstructure:"file_path"`
    MaxSize    int    `mapstructure:"max_size"`    // MB
    MaxBackups int    `mapstructure:"max_backups"`
    MaxAge     int    `mapstructure:"max_age"`     // days
    Compress   bool   `mapstructure:"compress"`
}

// Zap日志器初始化
func InitLogger(config LogConfig) (*zap.Logger, error) {
    var zapConfig zap.Config
    
    if config.Format == "json" {
        zapConfig = zap.NewProductionConfig()
    } else {
        zapConfig = zap.NewDevelopmentConfig()
    }
    
    // 设置日志级别
    level, err := zap.ParseAtomicLevel(config.Level)
    if err != nil {
        return nil, err
    }
    zapConfig.Level = level
    
    // 配置输出路径
    if config.Output == "file" || config.Output == "both" {
        zapConfig.OutputPaths = append(zapConfig.OutputPaths, config.FilePath)
    }
    
    return zapConfig.Build()
}
```

### 8.4 日志中间件
```go
// HTTP请求日志中间件
func LoggerMiddleware(logger *zap.Logger) app.HandlerFunc {
    return func(ctx context.Context, c *app.RequestContext) {
        start := time.Now()
        requestID := uuid.New().String()
        
        // 设置请求ID到上下文
        c.Set("request_id", requestID)
        
        // 处理请求
        c.Next(ctx)
        
        // 记录请求日志
        duration := time.Since(start)
        
        fields := []zap.Field{
            zap.String("request_id", requestID),
            zap.String("method", string(c.Method())),
            zap.String("path", string(c.Path())),
            zap.Int("status_code", c.Response.StatusCode()),
            zap.Duration("duration", duration),
            zap.String("ip", c.ClientIP()),
            zap.String("user_agent", string(c.UserAgent())),
            zap.String("category", LogCategoryAPI),
        }
        
        // 添加用户ID（如果已认证）
        if userID, exists := c.Get("user_id"); exists {
            fields = append(fields, zap.String("user_id", userID.(string)))
        }
        
        // 根据状态码选择日志级别
        if c.Response.StatusCode() >= 500 {
            logger.Error("HTTP Request", fields...)
        } else if c.Response.StatusCode() >= 400 {
            logger.Warn("HTTP Request", fields...)
        } else {
            logger.Info("HTTP Request", fields...)
        }
    }
}
```

### 8.5 业务日志记录
```go
// 业务操作日志
type BusinessLogger struct {
    logger *zap.Logger
}

func (bl *BusinessLogger) LogNodeOperation(ctx context.Context, operation string, nodeID string, userID string, metadata interface{}) {
    requestID, _ := ctx.Value("request_id").(string)
    
    bl.logger.Info("Node Operation",
        zap.String("request_id", requestID),
        zap.String("category", LogCategoryThinking),
        zap.String("action", operation),
        zap.String("resource", "node"),
        zap.String("node_id", nodeID),
        zap.String("user_id", userID),
        zap.Any("metadata", metadata),
    )
}

func (bl *BusinessLogger) LogAIInteraction(ctx context.Context, nodeID string, model string, prompt string, response string, duration time.Duration) {
    requestID, _ := ctx.Value("request_id").(string)
    
    bl.logger.Info("AI Interaction",
        zap.String("request_id", requestID),
        zap.String("category", LogCategoryThinking),
        zap.String("action", "ai_call"),
        zap.String("node_id", nodeID),
        zap.String("model", model),
        zap.Int("prompt_length", len(prompt)),
        zap.Int("response_length", len(response)),
        zap.Duration("duration", duration),
    )
}

func (bl *BusinessLogger) LogSSEEvent(ctx context.Context, userID string, mapID string, eventType string, data interface{}) {
    requestID, _ := ctx.Value("request_id").(string)
    
    bl.logger.Info("SSE Event",
        zap.String("request_id", requestID),
        zap.String("category", LogCategorySSE),
        zap.String("action", "send_event"),
        zap.String("user_id", userID),
        zap.String("map_id", mapID),
        zap.String("event_type", eventType),
        zap.Any("data", data),
    )
}
```

### 8.6 错误日志处理
```go
// 错误日志包装器
type ErrorLogger struct {
    logger *zap.Logger
}

func (el *ErrorLogger) LogError(ctx context.Context, err error, component string, action string, metadata map[string]interface{}) {
    requestID, _ := ctx.Value("request_id").(string)
    userID, _ := ctx.Value("user_id").(string)
    
    fields := []zap.Field{
        zap.String("request_id", requestID),
        zap.String("user_id", userID),
        zap.String("category", LogCategorySystem),
        zap.String("component", component),
        zap.String("action", action),
        zap.Error(err),
    }
    
    if metadata != nil {
        fields = append(fields, zap.Any("metadata", metadata))
    }
    
    el.logger.Error("System Error", fields...)
}

// 数据库错误日志
func (el *ErrorLogger) LogDBError(ctx context.Context, err error, query string, args interface{}) {
    el.LogError(ctx, err, "database", "query_failed", map[string]interface{}{
        "query": query,
        "args":  args,
    })
}

// AI服务错误日志
func (el *ErrorLogger) LogAIError(ctx context.Context, err error, model string, prompt string) {
    el.LogError(ctx, err, "ai_service", "request_failed", map[string]interface{}{
        "model":         model,
        "prompt_length": len(prompt),
    })
}
```

### 8.7 性能监控日志
```go
// 性能监控日志
type PerformanceLogger struct {
    logger *zap.Logger
}

func (pl *PerformanceLogger) LogSlowQuery(ctx context.Context, query string, duration time.Duration, threshold time.Duration) {
    if duration > threshold {
        requestID, _ := ctx.Value("request_id").(string)
        
        pl.logger.Warn("Slow Database Query",
            zap.String("request_id", requestID),
            zap.String("category", LogCategoryDB),
            zap.String("action", "slow_query"),
            zap.String("query", query),
            zap.Duration("duration", duration),
            zap.Duration("threshold", threshold),
        )
    }
}

func (pl *PerformanceLogger) LogMemoryUsage(ctx context.Context, component string, memStats runtime.MemStats) {
    pl.logger.Info("Memory Usage",
        zap.String("category", LogCategorySystem),
        zap.String("component", component),
        zap.String("action", "memory_stats"),
        zap.Uint64("alloc", memStats.Alloc),
        zap.Uint64("total_alloc", memStats.TotalAlloc),
        zap.Uint64("sys", memStats.Sys),
        zap.Uint32("num_gc", memStats.NumGC),
    )
}
```

### 8.8 日志聚合和分析
```yaml
# 日志收集配置 (使用ELK Stack)
filebeat.inputs:
- type: log
  enabled: true
  paths:
    - /var/log/thinkingmap/*.log
  json.keys_under_root: true
  json.add_error_key: true
  fields:
    service: thinkingmap-backend
    environment: production

output.elasticsearch:
  hosts: ["elasticsearch:9200"]
  index: "thinkingmap-logs-%{+yyyy.MM.dd}"

# Logstash配置示例
input {
  beats {
    port => 5044
  }
}

filter {
  if [service] == "thinkingmap-backend" {
    # 解析日志级别
    if [level] == "error" {
      mutate {
        add_tag => ["alert"]
      }
    }
    
    # 提取慢查询
    if [category] == "database" and [action] == "slow_query" {
      mutate {
        add_tag => ["performance_issue"]
      }
    }
  }
}

output {
  elasticsearch {
    hosts => ["elasticsearch:9200"]
    index => "thinkingmap-logs-%{+YYYY.MM.dd}"
  }
}
```

### 8.9 日志告警规则
```go
// 告警规则配置
type AlertRule struct {
    Name        string        `json:"name"`
    Condition   string        `json:"condition"`
    Threshold   int           `json:"threshold"`
    TimeWindow  time.Duration `json:"time_window"`
    Severity    string        `json:"severity"`
    Channels    []string      `json:"channels"`
}

var AlertRules = []AlertRule{
    {
        Name:       "High Error Rate",
        Condition:  "level:error",
        Threshold:  10,
        TimeWindow: 5 * time.Minute,
        Severity:   "critical",
        Channels:   []string{"slack", "email"},
    },
    {
        Name:       "Slow Database Queries",
        Condition:  "category:database AND action:slow_query",
        Threshold:  5,
        TimeWindow: 1 * time.Minute,
        Severity:   "warning",
        Channels:   []string{"slack"},
    },
    {
        Name:       "AI Service Failures",
        Condition:  "component:ai_service AND level:error",
        Threshold:  3,
        TimeWindow: 1 * time.Minute,
        Severity:   "high",
        Channels:   []string{"slack", "pagerduty"},
    },
}
```

## 9. 部署和运维
### 9.1 容器化部署
```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o thinkingmap-backend ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/thinkingmap-backend .
COPY --from=builder /app/configs ./configs

EXPOSE 8080
CMD ["./thinkingmap-backend"]
```

### 9.2 Kubernetes部署配置
```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: thinkingmap-backend
  labels:
    app: thinkingmap-backend
spec:
  replicas: 3
  selector:
    matchLabels:
      app: thinkingmap-backend
  template:
    metadata:
      labels:
        app: thinkingmap-backend
    spec:
      containers:
      - name: thinkingmap-backend
        image: thinkingmap/backend:latest
        ports:
        - containerPort: 8080
        env:
        - name: DB_HOST
          valueFrom:
            secretKeyRef:
              name: thinkingmap-secrets
              key: db-host
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: thinkingmap-secrets
              key: db-password
        - name: REDIS_URL
          valueFrom:
            secretKeyRef:
              name: thinkingmap-secrets
              key: redis-url
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: thinkingmap-secrets
              key: jwt-secret
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

### 9.3 监控和告警
```yaml
# 监控配置 (Prometheus)
apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus-config
data:
  prometheus.yml: |
    global:
      scrape_interval: 15s
    
    scrape_configs:
    - job_name: 'thinkingmap-backend'
      static_configs:
      - targets: ['thinkingmap-backend:8080']
      metrics_path: /metrics
      scrape_interval: 10s
      
    rule_files:
    - "alert_rules.yml"
    
    alerting:
      alertmanagers:
      - static_configs:
        - targets: ['alertmanager:9093']

# 告警规则
alert_rules.yml: |
  groups:
  - name: thinkingmap-backend
    rules:
    - alert: HighErrorRate
      expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.1
      for: 2m
      labels:
        severity: critical
      annotations:
        summary: "High error rate detected"
        
    - alert: DatabaseConnectionFailure
      expr: up{job="postgres"} == 0
      for: 1m
      labels:
        severity: critical
      annotations:
        summary: "Database connection failed"
```

## 10. 开发和测试
### 10.1 开发环境配置
```yaml
# docker-compose.dev.yml
version: '3.8'
services:
  postgres:
    image: postgres:14
    environment:
      POSTGRES_DB: thinkingmap_dev
      POSTGRES_USER: dev
      POSTGRES_PASSWORD: devpass
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./scripts/migration:/docker-entrypoint-initdb.d
  
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    command: redis-server --appendonly yes
    volumes:
      - redis_data:/data
  
  app:
    build:
      context: .
      dockerfile: Dockerfile.dev
    ports:
      - "8080:8080"
    environment:
      - ENV=development
      - DB_HOST=postgres
      - REDIS_URL=redis://redis:6379
    volumes:
      - .:/app
    depends_on:
      - postgres
      - redis

volumes:
  postgres_data:
  redis_data:
```

### 10.2 测试策略
```go
// 单元测试示例
func TestNodeService_CreateNode(t *testing.T) {
    // 准备测试数据
    mockRepo := &MockNodeRepository{}
    mockSSE := &MockSSEService{}
    service := NewNodeService(mockRepo, mockSSE)
    
    tests := []struct {
        name    string
        input   *CreateNodeRequest
        want    *Node
        wantErr bool
    }{
        {
            name: "valid node creation",
            input: &CreateNodeRequest{
                MapID:    "map-123",
                NodeType: "analysis",
                Question: "Test question",
                Target:   "Test target",
            },
            want: &Node{
                NodeType: "analysis",
                Question: "Test question",
                Target:   "Test target",
                Status:   StatusPending,
            },
            wantErr: false,
        },
        // 更多测试用例...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := service.CreateNode(context.Background(), tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("CreateNode() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got.NodeType, tt.want.NodeType) {
                t.Errorf("CreateNode() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### 10.3 集成测试
```go
// 集成测试示例
func TestIntegration_ThinkingFlow(t *testing.T) {
    // 设置测试环境
    testDB := setupTestDB(t)
    testRedis := setupTestRedis(t)
    testServer := setupTestServer(t, testDB, testRedis)
    
    // 创建测试用户
    user := createTestUser(t, testDB)
    token := generateTestToken(t, user)
    
    // 测试完整的思考流程
    t.Run("complete thinking flow", func(t *testing.T) {
        // 1. 创建思维导图
        mapResp := createThinkingMap(t, testServer, token, &CreateMapRequest{
            Title:        "Test Map",
            RootQuestion: "How to solve this problem?",
        })
        
        // 2. 开始问题分析
        analyzeResp := startAnalysis(t, testServer, token, mapResp.RootNodeID)
        assert.Equal(t, "processing", analyzeResp.Status)
        
        // 3. 验证SSE事件
        sseEvents := collectSSEEvents(t, testServer, token, mapResp.ID, 10*time.Second)
        assert.Contains(t, sseEvents, "thinking_progress")
        assert.Contains(t, sseEvents, "node_updated")
        
        // 4. 验证最终结果
        finalNode := getNode(t, testServer, token, mapResp.RootNodeID)
        assert.Equal(t, "completed", finalNode.Status)
        assert.NotEmpty(t, finalNode.Conclusion)
    })
}
```

---

## 总结
本设计文档详细描述了 ThinkingMap 后端系统的技术架构、数据设计、API接口、安全机制和日志系统。主要特点包括：

1. **技术栈现代化**: 采用 Golang + Hertz + Eino 的高性能技术栈
2. **架构清晰**: 分层架构设计，职责明确，易于维护和扩展
3. **数据设计合理**: 使用 PostgreSQL + JSONB 支持复杂数据结构
4. **实时通信**: 基于 SSE 的实时事件推送机制
5. **安全完善**: 多层次的安全防护措施
6. **可观测性**: 完整的日志、监控和告警体系
7. **可扩展性**: 支持水平扩展和微服务拆分

该设计为 ThinkingMap 产品提供了坚实的技术基础，能够支撑产品的快速迭代和规模化发展。

