# Enhanced MultiAgent 消息捕获与回调系统

## 概述

Enhanced MultiAgent 系统采用基于 Eino 框架的消息捕获机制，通过统一的 `MessageHandler` 接口实现对各个节点消息的监控和处理。这种设计简化了回调系统的复杂性，专注于核心的消息捕获功能。

## 核心接口设计

### MessageHandler - 统一消息处理器接口

```go
// MessageHandler 通用消息处理器接口
// 所有节点的消息处理器都实现此接口
type MessageHandler interface {
    // OnMessage 处理普通消息
    OnMessage(ctx context.Context, message *schema.Message) (context.Context, error)
    
    // OnStreamMessage 处理流式消息
    OnStreamMessage(ctx context.Context, message *schema.StreamReader[*schema.Message]) (context.Context, error)
}
```

### 内部实现机制

```go
// createMessageHandlerOption 创建通用的消息处理器option
// 这是内部函数，用于统一处理所有节点的消息捕获逻辑
func createMessageHandlerOption(handler MessageHandler, nodeKey string) agent.AgentOption {
    cmHandler := &ub.ModelCallbackHandler{
        OnEnd: func(ctx context.Context, runInfo *callbacks.RunInfo, output *model.CallbackOutput) context.Context {
            ctx, _ = handler.OnMessage(ctx, output.Message)
            return ctx
        },
        OnEndWithStreamOutput: func(ctx context.Context, runInfo *callbacks.RunInfo, output *schema.StreamReader[*model.CallbackOutput]) context.Context {
            c := func(output *model.CallbackOutput) (*schema.Message, error) {
                return output.Message, nil
            }
            s := schema.StreamReaderWithConvert(output, c)
            ctx, _ = handler.OnStreamMessage(ctx, s)
            return ctx
        },
    }
    cb := ub.NewHandlerHelper().ChatModel(cmHandler).Handler()
    option := agent.WithComposeOptions(compose.WithCallbacks(cb).DesignateNode(nodeKey))
    return option
}
```

## 可用的消息处理器选项

系统提供了以下预定义的消息处理器选项，每个选项对应一个特定的节点：

### 核心节点消息处理器

```go
// WithConversationAnalyzer 对话分析器消息处理器
func WithConversationAnalyzer(handler MessageHandler) agent.AgentOption {
    return createMessageHandlerOption(handler, "conversation_analyzer")
}

// WithDirectAnswerHandler 直接回答处理器消息处理器
func WithDirectAnswerHandler(handler MessageHandler) agent.AgentOption {
    return createMessageHandlerOption(handler, "direct_answer")
}

// WithPlanCreationHandler 计划创建处理器消息处理器
func WithPlanCreationHandler(handler MessageHandler) agent.AgentOption {
    return createMessageHandlerOption(handler, "plan_creation")
}

// WithFeedbackProcessorHandler 反馈处理器消息处理器
func WithFeedbackProcessorHandler(handler MessageHandler) agent.AgentOption {
    return createMessageHandlerOption(handler, "feedback_processor")
}

// WithPlanUpdateHandler 计划更新处理器消息处理器
func WithPlanUpdateHandler(handler MessageHandler) agent.AgentOption {
    return createMessageHandlerOption(handler, "plan_update")
}

// WithFinalAnswerHandler 最终答案处理器消息处理器
func WithFinalAnswerHandler(handler MessageHandler) agent.AgentOption {
    return createMessageHandlerOption(handler, "final_answer")
}

// WithPlanExecutionHandler 计划执行处理器消息处理器
func WithPlanExecutionHandler(handler MessageHandler) agent.AgentOption {
    return createMessageHandlerOption(handler, "plan_execution")
}

// WithResultCollectorHandler 结果收集器消息处理器
func WithResultCollectorHandler(handler MessageHandler) agent.AgentOption {
    return createMessageHandlerOption(handler, "result_collector")
}

// WithSpecialistHandler 专家处理器消息处理器
func WithSpecialistHandler(handler MessageHandler) agent.AgentOption {
    return createMessageHandlerOption(handler, "specialist")
}
```

### 使用示例

```go
// 自定义消息处理器实现
type CustomMessageHandler struct {
    logger *zap.Logger
}

func (h *CustomMessageHandler) OnMessage(ctx context.Context, message *schema.Message) (context.Context, error) {
    h.logger.Info("Received message",
        zap.String("role", string(message.Role)),
        zap.String("content", message.Content),
    )
    return ctx, nil
}

func (h *CustomMessageHandler) OnStreamMessage(ctx context.Context, message *schema.StreamReader[*schema.Message]) (context.Context, error) {
    h.logger.Info("Received stream message")
    return ctx, nil
}

// 在创建系统时使用消息处理器
handler := &CustomMessageHandler{logger: logger}
system, err := NewEnhancedMultiAgent(
    config,
    WithConversationAnalyzer(handler),
    WithPlanCreationHandler(handler),
    WithSpecialistHandler(handler),
)
```

## 节点映射说明

每个消息处理器选项对应系统中的特定节点，以下是详细的节点映射关系：

| 处理器选项 | 节点键值 | 功能描述 | 触发时机 |
|-----------|---------|----------|----------|
| `WithConversationAnalyzer` | `conversation_analyzer` | 对话分析器 | 分析用户输入和对话上下文时 |
| `WithDirectAnswerHandler` | `direct_answer` | 直接回答处理器 | 系统判断可以直接回答时 |
| `WithPlanCreationHandler` | `plan_creation` | 计划创建处理器 | 创建任务执行计划时 |
| `WithFeedbackProcessorHandler` | `feedback_processor` | 反馈处理器 | 处理执行结果反馈时 |
| `WithPlanUpdateHandler` | `plan_update` | 计划更新处理器 | 更新或修改执行计划时 |
| `WithFinalAnswerHandler` | `final_answer` | 最终答案处理器 | 生成最终回答时 |
| `WithPlanExecutionHandler` | `plan_execution` | 计划执行处理器 | 执行计划步骤时 |
| `WithResultCollectorHandler` | `result_collector` | 结果收集器 | 收集各专家执行结果时 |
| `WithSpecialistHandler` | `specialist` | 专家处理器 | 专家执行具体任务时 |

## 最佳实践

### 1. 选择性监控

```go
// 只监控关键节点，避免过度监控影响性能
system, err := NewEnhancedMultiAgent(
    config,
    WithPlanCreationHandler(planHandler),    // 监控计划创建
    WithSpecialistHandler(specialistHandler), // 监控专家执行
    WithFinalAnswerHandler(answerHandler),   // 监控最终答案
)
```

### 2. 异步处理

```go
type AsyncMessageHandler struct {
    msgChan chan *schema.Message
    logger  *zap.Logger
}

func (h *AsyncMessageHandler) OnMessage(ctx context.Context, message *schema.Message) (context.Context, error) {
    // 异步处理，避免阻塞主流程
    select {
    case h.msgChan <- message:
    default:
        h.logger.Warn("Message channel full, dropping message")
    }
    return ctx, nil
}
```

### 3. 错误处理

```go
type RobustMessageHandler struct {
    logger *zap.Logger
}

func (h *RobustMessageHandler) OnMessage(ctx context.Context, message *schema.Message) (context.Context, error) {
    defer func() {
        if r := recover(); r != nil {
            h.logger.Error("Message handler panic recovered", zap.Any("panic", r))
        }
    }()
    
    // 处理逻辑
    return ctx, nil
}
```

### 4. 性能监控

```go
type PerformanceMessageHandler struct {
    metrics map[string]time.Duration
    mutex   sync.RWMutex
}

func (h *PerformanceMessageHandler) OnMessage(ctx context.Context, message *schema.Message) (context.Context, error) {
    start := time.Now()
    defer func() {
        duration := time.Since(start)
        h.mutex.Lock()
        h.metrics["message_processing"] = duration
        h.mutex.Unlock()
    }()
    
    // 处理逻辑
    return ctx, nil
}
```

## 注意事项

1. **性能影响**: 消息处理器会在每次节点执行时被调用，应避免在处理器中执行耗时操作
2. **错误处理**: 处理器中的错误不会中断主流程，但会被记录
3. **线程安全**: 如果处理器需要维护状态，必须确保线程安全
4. **资源管理**: 及时释放处理器中使用的资源，避免内存泄漏
5. **日志记录**: 建议在处理器中添加适当的日志记录，便于调试和监控

## 总结

Enhanced MultiAgent 的消息捕获系统通过统一的 `MessageHandler` 接口和预定义的选项函数，提供了简洁而强大的节点消息监控能力。这种设计：

- **简化了接口**: 统一的 `MessageHandler` 接口替代了复杂的回调系统
- **提高了可维护性**: 通过 `createMessageHandlerOption` 函数消除了代码重复
- **增强了灵活性**: 支持选择性监控特定节点
- **保证了性能**: 轻量级的消息处理机制，最小化对主流程的影响

通过合理使用这些消息处理器，可以实现对 Enhanced MultiAgent 系统的全面监控和分析，为系统优化和问题诊断提供有力支持。