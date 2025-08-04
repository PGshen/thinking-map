# Enhanced Planning MultiAgent

## 概述

增强版规划多智能体（Enhanced Planning MultiAgent）是基于CloudWeGo/Eino框架开发的智能代理系统，具备自主规划、迭代执行和结果汇总的能力。该系统通过hostAgent的规划能力，协调多个专家智能体完成复杂任务。

## 核心特性

### 1. 规划能力
- **自主规划**：hostAgent能够根据用户输入自动生成执行计划
- **动态调整**：根据执行结果动态更新和优化计划
- **步骤分解**：将复杂任务分解为可执行的步骤

### 2. 专家协作
- **专家选择**：根据任务需求自动选择合适的专家智能体
- **并行执行**：支持多个专家智能体并行处理不同任务
- **结果反馈**：专家执行结果实时反馈给hostAgent

### 3. 迭代执行
- **状态跟踪**：实时跟踪任务执行状态
- **错误处理**：自动处理执行过程中的错误和异常
- **重试机制**：支持失败步骤的重试

### 4. 结果汇总
- **智能汇总**：自动汇总所有专家的执行结果
- **格式化输出**：提供结构化的最终结果

## 架构设计

### 核心组件

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   PlannerAgent  │    │   Specialists   │    │   Summarizer    │
│                 │    │                 │    │                 │
│ - 规划生成      │    │ - 任务执行      │    │ - 结果汇总      │
│ - 计划更新      │    │ - 专业能力      │    │ - 格式化输出    │
│ - 状态管理      │    │ - 并行处理      │    │ - 质量评估      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌─────────────────┐
                    │ PlanningState   │
                    │                 │
                    │ - 执行计划      │
                    │ - 当前状态      │
                    │ - 迭代计数      │
                    │ - 结果存储      │
                    └─────────────────┘
```

### 执行流程

```
用户输入 → 规划生成 → 专家选择 → 任务执行 → 状态更新 → 完成检查
    ↑                                                      ↓
    └──────────── 计划更新 ←──── 迭代控制 ←──── 未完成 ←────┘
                                    ↓
                                 已完成
                                    ↓
                              结果汇总 → 最终输出
```

## 使用示例

### 基本使用

```go
package main

import (
    "context"
    "fmt"
    
    "github.com/cloudwego/eino/flow/agent/multiagent/enhanced"
    "github.com/cloudwego/eino/schema"
)

func main() {
    // 创建配置
    config := &enhanced.PlanningMultiAgentConfig{
        PlannerAgent: &enhanced.PlannerAgent{
            ChatModel:       yourPlannerModel,
            PlanningPrompt:  "请为以下任务制定执行计划：{{query}}",
            UpdatePrompt:    "根据执行结果更新计划：{{results}}",
        },
        Specialists: []*enhanced.Specialist{
            {
                AgentMeta: &enhanced.AgentMeta{
                    Name: "DataAnalyst",
                },
                ChatModel:    yourSpecialistModel,
                SystemPrompt: "你是一个数据分析专家",
            },
        },
        MaxIterations: 5,
        Summarizer: &enhanced.Summarizer{
            ChatModel:     yourSummarizerModel,
            SummaryPrompt: "请汇总执行结果：{{execution_results}}",
        },
    }
    
    // 创建智能体
    agent, err := enhanced.NewPlanningMultiAgent(config)
    if err != nil {
        panic(err)
    }
    
    // 执行任务
    ctx := context.Background()
    input := schema.UserMessage("分析销售数据并生成报告")
    
    response, err := agent.Generate(ctx, input)
    if err != nil {
        panic(err)
    }
    
    fmt.Println("执行结果:", response.Content)
}
```

### 带回调的使用

```go
// 自定义回调
type MyCallback struct{}

func (c *MyCallback) OnPlanGenerated(info *enhanced.PlanGeneratedInfo) {
    fmt.Printf("计划生成完成，包含 %d 个步骤\n", len(info.Plan.Steps))
}

func (c *MyCallback) OnStepStarted(info *enhanced.StepStartedInfo) {
    fmt.Printf("开始执行步骤：%s\n", info.Step.Description)
}

func (c *MyCallback) OnStepCompleted(info *enhanced.StepCompletedInfo) {
    fmt.Printf("步骤完成：%s\n", info.Step.Description)
}

func (c *MyCallback) OnIterationCompleted(info *enhanced.IterationCompletedInfo) {
    fmt.Printf("第 %d 次迭代完成\n", info.Iteration)
}

func (c *MyCallback) OnTaskCompleted(info *enhanced.TaskCompletedInfo) {
    fmt.Println("任务执行完成")
}

func (c *MyCallback) OnError(info *enhanced.ErrorInfo) {
    fmt.Printf("执行错误：%s\n", info.Error.Error())
}

// 使用回调
callback := &MyCallback{}
agent, err := enhanced.NewPlanningMultiAgent(config, 
    enhanced.WithCallbacks(callback),
    enhanced.WithMaxRetries(3),
    enhanced.WithLogging(true),
)
```

## 配置选项

### PlanningMultiAgentConfig

- **PlannerAgent**: 规划智能体配置
- **Specialists**: 专家智能体列表
- **MaxIterations**: 最大迭代次数
- **Summarizer**: 结果汇总器配置

### 可选配置

- **WithCallbacks**: 设置回调函数
- **WithMaxRetries**: 设置最大重试次数
- **WithLogging**: 启用日志记录

## 扩展开发

### 自定义专家智能体

```go
type CustomSpecialist struct {
    *enhanced.Specialist
    customField string
}

func (s *CustomSpecialist) Execute(ctx context.Context, step *enhanced.ExecutionStep) (*schema.Message, error) {
    // 自定义执行逻辑
    return schema.AssistantMessage("执行结果", nil), nil
}
```

### 自定义汇总器

```go
type CustomSummarizer struct {
    *enhanced.Summarizer
}

func (s *CustomSummarizer) Summarize(ctx context.Context, plan *enhanced.ExecutionPlan) (*schema.Message, error) {
    // 自定义汇总逻辑
    return schema.AssistantMessage("汇总结果", nil), nil
}
```

## 性能优化

### 并行执行
- 支持多个专家智能体并行处理独立任务
- 自动管理并发控制和资源分配

### 缓存机制
- 计划缓存：避免重复生成相同的执行计划
- 结果缓存：缓存专家执行结果，提高重试效率

### 错误恢复
- 自动重试失败的步骤
- 智能降级策略
- 部分失败容错机制

## 最佳实践

1. **合理设置迭代次数**：根据任务复杂度设置合适的最大迭代次数
2. **专家能力匹配**：确保专家智能体的能力与任务需求匹配
3. **提示词优化**：精心设计规划、专家和汇总的提示词
4. **回调监控**：使用回调函数监控执行过程和性能指标
5. **错误处理**：实现完善的错误处理和恢复机制

## 故障排除

### 常见问题

1. **计划生成失败**
   - 检查规划模型配置
   - 优化规划提示词
   - 增加重试次数

2. **专家执行超时**
   - 调整超时设置
   - 优化专家模型性能
   - 检查网络连接

3. **迭代次数过多**
   - 检查完成条件设置
   - 优化任务分解粒度
   - 调整专家能力匹配

## 版本历史

- **v1.0.0**: 初始版本，支持基本的规划和执行功能
- 后续版本将持续优化性能和扩展功能

## 贡献指南

欢迎提交Issue和Pull Request来改进这个项目。请确保：

1. 遵循代码规范
2. 添加适当的测试
3. 更新相关文档
4. 保持向后兼容性

## 许可证

本项目采用Apache License 2.0许可证。详见LICENSE文件。