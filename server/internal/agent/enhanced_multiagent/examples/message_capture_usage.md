# Enhanced MultiAgent 消息捕获使用示例

本文档展示如何使用新增的消息捕获option函数来监控和处理Enhanced MultiAgent系统中各个节点的消息流。

## 概述

参考 `WithConversationAnalyzer`，我们为Enhanced MultiAgent编排中的其他关键节点也设计了对应的option函数，方便上层业务捕获和处理消息：

- `WithDirectAnswerHandler` - 直接回答节点消息捕获
- `WithPlanCreationHandler` - 计划创建节点消息捕获
- `WithFeedbackProcessorHandler` - 反馈处理节点消息捕获
- `WithPlanUpdateHandler` - 计划更新节点消息捕获
- `WithFinalAnswerHandler` - 最终回答节点消息捕获
- `WithPlanExecutionHandler` - 计划执行节点消息捕获
- `WithResultCollectorHandler` - 结果收集节点消息捕获
- `WithSpecialistHandler` - 专家节点消息捕获

## 基本使用方法

### 1. 实现消息处理接口

每个handler都需要实现对应的接口，包含 `OnMessage` 和 `OnStreamMessage` 方法：

```go
// 直接回答处理器示例
type MyDirectAnswerHandler struct {
    logger *log.Logger
}

func (h *MyDirectAnswerHandler) OnMessage(ctx context.Context, message *schema.Message) (context.Context, error) {
    h.logger.Printf("直接回答节点消息: %s", message.Content)
    // 可以在这里进行业务处理，如保存到数据库、发送通知等
    return ctx, nil
}

func (h *MyDirectAnswerHandler) OnStreamMessage(ctx context.Context, message *schema.StreamReader[*schema.Message]) (context.Context, error) {
    h.logger.Println("直接回答节点流式消息开始")
    for {
        msg, err := message.Recv()
        if err != nil {
            break
        }
        h.logger.Printf("流式内容: %s", msg.Content)
    }
    return ctx, nil
}
```

### 2. 创建Agent并应用Option

```go
func main() {
    ctx := context.Background()
    
    // 创建配置
    config := &enhanced.EnhancedMultiAgentConfig{
        Name: "消息捕获示例",
        Host: enhanced.EnhancedHost{
            Model: chatModel,
            SystemPrompt: "你是一个智能助手",
        },
        Specialists: []*enhanced.EnhancedSpecialist{
            {
                Name: "代码专家",
                IntendedUse: "处理编程相关任务",
                ChatModel: chatModel,
            },
        },
    }
    
    // 创建Agent
    agent, err := enhanced.NewEnhancedMultiAgent(ctx, config)
    if err != nil {
        panic(err)
    }
    
    // 创建消息处理器
    directAnswerHandler := &MyDirectAnswerHandler{logger: log.New(os.Stdout, "[DirectAnswer] ", log.LstdFlags)}
    planCreationHandler := &MyPlanCreationHandler{logger: log.New(os.Stdout, "[PlanCreation] ", log.LstdFlags)}
    
    // 准备输入
    input := []*schema.Message{
        {
            Role:    schema.User,
            Content: "请帮我分析一下Go语言的并发模型",
        },
    }
    
    // 使用多个option执行生成
    result, err := agent.Generate(ctx, input,
        enhanced.WithDirectAnswerHandler(directAnswerHandler),
        enhanced.WithPlanCreationHandler(planCreationHandler),
    )
    
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("最终结果: %s\n", result.Content)
}
```

## 高级使用场景

### 1. 消息持久化

```go
type DatabaseMessageHandler struct {
    db *sql.DB
    nodeType string
}

func (h *DatabaseMessageHandler) OnMessage(ctx context.Context, message *schema.Message) (context.Context, error) {
    // 保存消息到数据库
    _, err := h.db.ExecContext(ctx, 
        "INSERT INTO agent_messages (node_type, role, content, timestamp) VALUES (?, ?, ?, ?)",
        h.nodeType, string(message.Role), message.Content, time.Now())
    return ctx, err
}

func (h *DatabaseMessageHandler) OnStreamMessage(ctx context.Context, message *schema.StreamReader[*schema.Message]) (context.Context, error) {
    var fullContent strings.Builder
    for {
        msg, err := message.Recv()
        if err != nil {
            break
        }
        fullContent.WriteString(msg.Content)
    }
    
    // 保存完整的流式消息
    _, err := h.db.ExecContext(ctx,
        "INSERT INTO agent_messages (node_type, role, content, timestamp, is_stream) VALUES (?, ?, ?, ?, ?)",
        h.nodeType, "assistant", fullContent.String(), time.Now(), true)
    return ctx, err
}
```

### 2. 实时监控和告警

```go
type MonitoringHandler struct {
    metrics *prometheus.CounterVec
    alertManager AlertManager
}

func (h *MonitoringHandler) OnMessage(ctx context.Context, message *schema.Message) (context.Context, error) {
    // 记录指标
    h.metrics.WithLabelValues("plan_creation", string(message.Role)).Inc()
    
    // 检查是否需要告警
    if strings.Contains(message.Content, "error") || strings.Contains(message.Content, "failed") {
        h.alertManager.SendAlert("计划创建节点出现错误", message.Content)
    }
    
    return ctx, nil
}
```

### 3. 消息过滤和转换

```go
type FilteringHandler struct {
    filter func(string) bool
    transformer func(string) string
    downstream MessageHandler
}

func (h *FilteringHandler) OnMessage(ctx context.Context, message *schema.Message) (context.Context, error) {
    // 过滤消息
    if !h.filter(message.Content) {
        return ctx, nil
    }
    
    // 转换消息
    transformedContent := h.transformer(message.Content)
    transformedMessage := &schema.Message{
        Role: message.Role,
        Content: transformedContent,
    }
    
    // 传递给下游处理器
    return h.downstream.OnMessage(ctx, transformedMessage)
}
```

## 专家节点特殊处理

专家节点的handler需要指定专家名称：

```go
// 为特定专家创建处理器
specialistHandler := &MySpecialistHandler{specialistName: "代码专家"}

// 应用到特定专家
result, err := agent.Generate(ctx, input,
    enhanced.WithSpecialistHandler("代码专家", specialistHandler),
)
```

## 最佳实践

1. **错误处理**: Handler中的错误不会中断Agent执行，但应该记录日志
2. **性能考虑**: 避免在Handler中执行耗时操作，考虑异步处理
3. **内存管理**: 流式消息处理时注意内存使用，避免缓存过多数据
4. **上下文传播**: 可以通过返回的context传递额外信息给后续处理
5. **组合使用**: 可以同时使用多个option来监控不同节点

## 注意事项

- Handler接口的实现必须是线程安全的
- 流式消息只能读取一次，读取后流会被消耗
- Handler中的错误会被记录但不会中断Agent执行
- 建议为每个Handler添加适当的日志记录