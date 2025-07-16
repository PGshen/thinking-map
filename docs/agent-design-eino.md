# ThinkingMap Agent技术设计方案 (基于Eino框架)

## 1. 设计概述

### 1.1 设计目标
基于cloudwego/eino框架重新设计ThinkingMap的Agent系统，实现：
- **模块化组件设计**：利用Eino的组件抽象，构建可复用的AI组件
- **灵活的编排能力**：通过Chain和Graph编排实现复杂的思维流程
- **类型安全保证**：编译时类型检查，确保组件间数据流的正确性
- **流式处理支持**：原生支持LLM流式响应的处理和传递
- **可观测性**：内置的回调机制提供完整的执行追踪

### 1.2 核心价值
- **简化开发**：通过Eino的高级抽象减少样板代码
- **提升可靠性**：类型安全和内置错误处理机制
- **增强可维护性**：组件化设计便于测试和扩展
- **优化性能**：Eino的并发管理和流处理优化

## 2. 架构设计

### 2.1 整体架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                        Web Client                              │
│                   (React + TypeScript)                         │
└─────────────────────┬───────────────────────────────────────────┘
                      │ HTTP/SSE
                      │
┌─────────────────────┴───────────────────────────────────────────┐
│                    API Gateway                                 │
│              (Gin + Middleware)                                │
└─────────────────────┬───────────────────────────────────────────┘
                      │
┌─────────────────────┴───────────────────────────────────────────┐
│                  Service Layer                                 │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐  │
│  │ Node Service│  │Think Service│  │     User Service        │  │
│  │             │  │             │  │                         │  │
│  │ • CRUD Ops  │  │ • Agent Orch│  │ • Authentication        │  │
│  │ • Graph Mgmt│  │ • Flow Ctrl │  │ • Authorization         │  │
│  │ • SSE Events│  │ • Context   │  │ • Session Management    │  │
│  └─────────────┘  └─────────────┘  └─────────────────────────┘  │
└─────────────────────┬───────────────────────────────────────────┘
                      │
┌─────────────────────┴───────────────────────────────────────────┐
│                  Eino Agent Layer                              │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                 Core Components                         │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────────┐  │   │
│  │  │ ChatModel   │  │ ChatTemplate│  │   ToolsNode     │  │   │
│  │  │             │  │             │  │                 │  │   │
│  │  │ • OpenAI    │  │ • Prompt    │  │ • Search Tools  │  │   │
│  │  │ • Claude    │  │ • Variables │  │ • Custom Tools  │  │   │
│  │  │ • Streaming │  │ • Templates │  │ • Tool Calling  │  │   │
│  │  └─────────────┘  └─────────────┘  └─────────────────┘  │   │
│  │                                                         │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────────┐  │   │
│  │  │ Retriever   │  │ Embedding   │  │   Lambda        │  │   │
│  │  │             │  │             │  │                 │  │   │
│  │  │ • Vector DB │  │ • OpenAI    │  │ • Custom Logic  │  │   │
│  │  │ • Semantic  │  │ • Text Embed│  │ • Data Process  │  │   │
│  │  │ • Context   │  │ • Similarity│  │ • Validation    │  │   │
│  │  └─────────────┘  └─────────────┘  └─────────────────┘  │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                 Orchestration Layer                     │   │
│  │                                                         │   │
│  │  ┌─────────────────────────────────────────────────┐   │   │
│  │  │              Chain Orchestration                │   │   │
│  │  │                                                 │   │   │
│  │  │  Intent → Template → Model → Response          │   │   │
│  │  │     ↓         ↓        ↓        ↓             │   │   │
│  │  │  Simple linear workflows for basic tasks       │   │   │
│  │  └─────────────────────────────────────────────────┘   │   │
│  │                                                         │   │
│  │  ┌─────────────────────────────────────────────────┐   │   │
│  │  │              Graph Orchestration                │   │   │
│  │  │                                                 │   │   │
│  │  │    ┌─────────┐    ┌─────────┐    ┌─────────┐   │   │   │
│  │  │    │ Intent  │    │Decompose│    │Analysis │   │   │   │
│  │  │    │ Agent   │───▶│ Agent   │───▶│ Agent   │   │   │   │
│  │  │    └─────────┘    └─────────┘    └─────────┘   │   │   │
│  │  │         │             │              │         │   │   │
│  │  │         ▼             ▼              ▼         │   │   │
│  │  │    ┌─────────┐    ┌─────────┐    ┌─────────┐   │   │   │
│  │  │    │Context  │    │Reasoning│    │Synthesis│   │   │   │
│  │  │    │Manager  │    │ Agent   │    │ Agent   │   │   │   │
│  │  │    └─────────┘    └─────────┘    └─────────┘   │   │   │
│  │  │                                                 │   │   │
│  │  │  Complex workflows with branching and loops     │   │   │
│  │  └─────────────────────────────────────────────────┘   │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                 Flow Implementations                    │   │
│  │                                                         │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────────┐  │   │
│  │  │ Intent Flow │  │Decompose Flow│  │ Synthesis Flow  │  │   │
│  │  │             │  │             │  │                 │  │   │
│  │  │ • User Input│  │ • Problem   │  │ • Evidence      │  │   │
│  │  │ • Intent    │  │   Analysis  │  │   Integration   │  │   │
│  │  │ • Context   │  │ • Sub-tasks │  │ • Conclusion    │  │   │
│  │  │ • Routing   │  │ • Dependencies│  │ • Validation   │  │   │
│  │  └─────────────┘  └─────────────┘  └─────────────────┘  │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────┬───────────────────────────────────────────┘
                      │
┌─────────────────────┴───────────────────────────────────────────┐
│                    Data Layer                                  │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐  │
│  │ PostgreSQL  │  │    Redis    │  │     Vector Store        │  │
│  │             │  │             │  │                         │  │
│  │ • Nodes     │  │ • Cache     │  │ • Embeddings            │  │
│  │ • Users     │  │ • Sessions  │  │ • Knowledge Base        │  │
│  │ • History   │  │ • State     │  │ • Semantic Search       │  │
│  └─────────────┘  └─────────────┘  └─────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

### 2.2 Eino组件映射

#### 2.2.1 核心组件定义

```go
// 核心组件接口定义
type ThinkingComponents struct {
    // 基础组件
    ChatModel    eino.ChatModel
    ChatTemplate eino.ChatTemplate
    ToolsNode    eino.ToolsNode
    Retriever    eino.Retriever
    Embedding    eino.Embedding
    
    // 自定义Lambda组件
    IntentClassifier    eino.Lambda[*schema.Message, *IntentResult]
    ProblemAnalyzer     eino.Lambda[*ProblemInput, *AnalysisResult]
    TaskDecomposer      eino.Lambda[*AnalysisResult, *DecomposeResult]
    EvidenceIntegrator  eino.Lambda[*EvidenceInput, *SynthesisResult]
    QualityValidator    eino.Lambda[*ValidationInput, *ValidationResult]
}

// 数据流类型定义
type IntentResult struct {
    Intent     string            `json:"intent"`
    Confidence float64          `json:"confidence"`
    Entities   map[string]string `json:"entities"`
    Context    map[string]any    `json:"context"`
}

type ProblemInput struct {
    Question    string            `json:"question"`
    Target      string            `json:"target"`
    Context     []string          `json:"context"`
    Constraints map[string]any    `json:"constraints"`
}

type AnalysisResult struct {
    Complexity   string   `json:"complexity"`
    Domain       string   `json:"domain"`
    KeyConcepts  []string `json:"key_concepts"`
    NeedsDecompose bool   `json:"needs_decompose"`
    Strategy     string   `json:"strategy"`
}

type DecomposeResult struct {
    SubTasks     []SubTask         `json:"sub_tasks"`
    Dependencies []Dependency      `json:"dependencies"`
    ExecutionPlan map[string]any   `json:"execution_plan"`
}

type SubTask struct {
    ID          string            `json:"id"`
    Question    string            `json:"question"`
    Target      string            `json:"target"`
    Priority    int               `json:"priority"`
    EstimatedTime string          `json:"estimated_time"`
    RequiredTools []string        `json:"required_tools"`
}
```

#### 2.2.2 组件实现策略

**ChatModel组件**
```go
// 支持多种LLM提供商
func NewChatModel(config *Config) (eino.ChatModel, error) {
    switch config.Provider {
    case "openai":
        return openai.NewChatModel(ctx, &openai.Config{
            APIKey: config.OpenAI.APIKey,
            Model:  config.OpenAI.Model,
        })
    case "claude":
        return claude.NewChatModel(ctx, &claude.Config{
            APIKey: config.Claude.APIKey,
            Model:  config.Claude.Model,
        })
    default:
        return nil, fmt.Errorf("unsupported provider: %s", config.Provider)
    }
}
```

**ChatTemplate组件**
```go
// 预定义的提示模板
var (
    IntentClassificationTemplate = `
你是一个意图识别专家。请分析用户输入，识别其意图类型。

用户输入：{{.query}}
上下文：{{.context}}

请返回JSON格式的结果：
{
  "intent": "问题类型(research/creative/analysis/planning)",
  "confidence": 0.95,
  "entities": {"关键实体": "值"},
  "context": {"上下文信息": "值"}
}
`

    ProblemAnalysisTemplate = `
你是一个问题分析专家。请分析以下问题的复杂度和特征。

问题：{{.question}}
目标：{{.target}}
约束条件：{{.constraints}}

请分析并返回JSON格式：
{
  "complexity": "simple/medium/complex",
  "domain": "问题领域",
  "key_concepts": ["关键概念1", "关键概念2"],
  "needs_decompose": true/false,
  "strategy": "分析策略描述"
}
`

    TaskDecomposeTemplate = `
你是一个任务分解专家。请将复杂问题分解为可执行的子任务。

原问题：{{.question}}
分析结果：{{.analysis}}

请返回JSON格式的分解结果：
{
  "sub_tasks": [
    {
      "id": "task_1",
      "question": "子任务问题",
      "target": "子任务目标",
      "priority": 1,
      "estimated_time": "预估时间",
      "required_tools": ["需要的工具"]
    }
  ],
  "dependencies": [
    {"from": "task_1", "to": "task_2", "type": "prerequisite"}
  ],
  "execution_plan": {"strategy": "执行策略"}
}
`
)
```

**ToolsNode组件**
```go
// 工具节点配置
func NewToolsNode() (eino.ToolsNode, error) {
    tools := []*schema.ToolInfo{
        {
            Name: "web_search",
            Description: "搜索网络信息",
            Parameters: schema.NewParameters(
                schema.NewParameter("query", schema.String, "搜索查询", true),
                schema.NewParameter("num_results", schema.Integer, "结果数量", false),
            ),
        },
        {
            Name: "knowledge_retrieve",
            Description: "检索知识库信息",
            Parameters: schema.NewParameters(
                schema.NewParameter("query", schema.String, "检索查询", true),
                schema.NewParameter("top_k", schema.Integer, "返回数量", false),
            ),
        },
    }
    
    return toolsnode.NewToolsNode(tools, map[string]eino.Tool{
        "web_search": duckduckgo.NewTool(),
        "knowledge_retrieve": NewKnowledgeRetriever(),
    })
}
```

## 3. 核心流程设计

### 3.1 意图识别流程 (Chain)

```go
// 简单的意图识别链
func NewIntentChain(components *ThinkingComponents) (*eino.CompiledChain, error) {
    chain, err := eino.NewChain[map[string]any, *IntentResult]().
        AppendChatTemplate(components.ChatTemplate).
        AppendChatModel(components.ChatModel).
        AppendLambda(components.IntentClassifier).
        Compile(ctx)
    
    if err != nil {
        return nil, fmt.Errorf("failed to compile intent chain: %w", err)
    }
    
    return chain, nil
}

// 使用示例
func (s *ThinkingService) ClassifyIntent(ctx context.Context, input map[string]any) (*IntentResult, error) {
    result, err := s.intentChain.Invoke(ctx, input)
    if err != nil {
        return nil, fmt.Errorf("intent classification failed: %w", err)
    }
    return result, nil
}
```

### 3.2 问题分解流程 (Graph)

```go
// 复杂的问题分解图
func NewDecomposeGraph(components *ThinkingComponents) (*eino.CompiledGraph, error) {
    graph := eino.NewGraph[*ProblemInput, *DecomposeResult]()
    
    // 添加节点
    err := graph.AddChatTemplateNode("analysis_template", components.ChatTemplate)
    if err != nil {
        return nil, err
    }
    
    err = graph.AddChatModelNode("analysis_model", components.ChatModel)
    if err != nil {
        return nil, err
    }
    
    err = graph.AddLambdaNode("problem_analyzer", components.ProblemAnalyzer)
    if err != nil {
        return nil, err
    }
    
    err = graph.AddLambdaNode("task_decomposer", components.TaskDecomposer)
    if err != nil {
        return nil, err
    }
    
    err = graph.AddToolsNode("search_tools", components.ToolsNode)
    if err != nil {
        return nil, err
    }
    
    err = graph.AddLambdaNode("quality_validator", components.QualityValidator)
    if err != nil {
        return nil, err
    }
    
    // 添加边
    graph.AddEdge(eino.START, "analysis_template")
    graph.AddEdge("analysis_template", "analysis_model")
    graph.AddEdge("analysis_model", "problem_analyzer")
    
    // 条件分支：是否需要搜索
    graph.AddBranch("problem_analyzer", func(ctx context.Context, result *AnalysisResult) (string, error) {
        if result.NeedsDecompose {
            return "search_tools", nil
        }
        return "task_decomposer", nil
    })
    
    graph.AddEdge("search_tools", "task_decomposer")
    graph.AddEdge("task_decomposer", "quality_validator")
    graph.AddEdge("quality_validator", eino.END)
    
    return graph.Compile(ctx)
}
```

### 3.3 结论生成流程 (Graph)

```go
// 结论生成图
func NewSynthesisGraph(components *ThinkingComponents) (*eino.CompiledGraph, error) {
    graph := eino.NewGraph[*EvidenceInput, *SynthesisResult]()
    
    // 证据收集节点
    graph.AddRetrieverNode("evidence_retriever", components.Retriever)
    
    // 证据整合节点
    graph.AddLambdaNode("evidence_integrator", components.EvidenceIntegrator)
    
    // 结论生成节点
    graph.AddChatTemplateNode("synthesis_template", components.ChatTemplate)
    graph.AddChatModelNode("synthesis_model", components.ChatModel)
    
    // 质量验证节点
    graph.AddLambdaNode("quality_validator", components.QualityValidator)
    
    // 构建执行流程
    graph.AddEdge(eino.START, "evidence_retriever")
    graph.AddEdge("evidence_retriever", "evidence_integrator")
    graph.AddEdge("evidence_integrator", "synthesis_template")
    graph.AddEdge("synthesis_template", "synthesis_model")
    graph.AddEdge("synthesis_model", "quality_validator")
    
    // 质量检查循环
    graph.AddBranch("quality_validator", func(ctx context.Context, result *ValidationResult) (string, error) {
        if result.IsValid {
            return eino.END, nil
        }
        return "synthesis_template", nil // 重新生成
    })
    
    return graph.Compile(ctx)
}
```

## 4. 高级特性实现

### 4.1 流式处理支持

```go
// 流式响应处理
func (s *ThinkingService) StreamDecompose(ctx context.Context, input *ProblemInput, callback func(*StreamEvent)) error {
    // 创建流式处理的回调处理器
    handler := eino.NewHandlerBuilder().
        OnStartFn(func(ctx context.Context, info *eino.RunInfo, input eino.CallbackInput) context.Context {
            callback(&StreamEvent{
                Type: "start",
                Data: map[string]any{"node": info.NodeName},
            })
            return ctx
        }).
        OnEndFn(func(ctx context.Context, info *eino.RunInfo, output eino.CallbackOutput) context.Context {
            callback(&StreamEvent{
                Type: "complete",
                Data: map[string]any{
                    "node": info.NodeName,
                    "result": output,
                },
            })
            return ctx
        }).
        OnStreamFn(func(ctx context.Context, info *eino.RunInfo, chunk eino.CallbackStreamChunk) context.Context {
            callback(&StreamEvent{
                Type: "stream",
                Data: map[string]any{
                    "node": info.NodeName,
                    "chunk": chunk,
                },
            })
            return ctx
        }).
        Build()
    
    // 执行图并处理流式响应
    _, err := s.decomposeGraph.Invoke(ctx, input, eino.WithCallbacks(handler))
    return err
}

type StreamEvent struct {
    Type string         `json:"type"`
    Data map[string]any `json:"data"`
}
```

### 4.2 上下文管理

```go
// 上下文管理器
type ContextManager struct {
    memory    map[string]any
    history   []HistoryItem
    state     map[string]any
    mutex     sync.RWMutex
}

func (cm *ContextManager) UpdateContext(key string, value any) {
    cm.mutex.Lock()
    defer cm.mutex.Unlock()
    
    cm.memory[key] = value
    cm.history = append(cm.history, HistoryItem{
        Timestamp: time.Now(),
        Action:    "update",
        Key:       key,
        Value:     value,
    })
}

func (cm *ContextManager) GetContext(key string) (any, bool) {
    cm.mutex.RLock()
    defer cm.mutex.RUnlock()
    
    value, exists := cm.memory[key]
    return value, exists
}

// 在Eino组件中使用上下文
func NewContextAwareLambda(cm *ContextManager) eino.Lambda[*ProblemInput, *AnalysisResult] {
    return eino.NewLambda(func(ctx context.Context, input *ProblemInput) (*AnalysisResult, error) {
        // 获取历史上下文
        if history, exists := cm.GetContext("analysis_history"); exists {
            input.Context = append(input.Context, fmt.Sprintf("历史分析: %v", history))
        }
        
        // 执行分析逻辑
        result := &AnalysisResult{
            // ... 分析逻辑
        }
        
        // 更新上下文
        cm.UpdateContext("last_analysis", result)
        
        return result, nil
    })
}
```

### 4.3 错误处理和重试

```go
// 带重试机制的组件包装
func WithRetry(component eino.ChatModel, maxRetries int) eino.ChatModel {
    return &RetryWrapper{
        component:   component,
        maxRetries:  maxRetries,
    }
}

type RetryWrapper struct {
    component  eino.ChatModel
    maxRetries int
}

func (r *RetryWrapper) Generate(ctx context.Context, input []*schema.Message, opts ...eino.Option) (*schema.Message, error) {
    var lastErr error
    
    for i := 0; i <= r.maxRetries; i++ {
        result, err := r.component.Generate(ctx, input, opts...)
        if err == nil {
            return result, nil
        }
        
        lastErr = err
        
        // 指数退避
        if i < r.maxRetries {
            backoff := time.Duration(math.Pow(2, float64(i))) * time.Second
            select {
            case <-ctx.Done():
                return nil, ctx.Err()
            case <-time.After(backoff):
                continue
            }
        }
    }
    
    return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}
```

## 5. 部署和配置

### 5.1 组件配置

```yaml
# config.yaml
eino:
  components:
    chat_model:
      provider: "openai"
      config:
        api_key: "${OPENAI_API_KEY}"
        model: "gpt-4"
        temperature: 0.7
        max_tokens: 2048
    
    tools:
      search:
        provider: "duckduckgo"
        config:
          max_results: 10
      
      knowledge:
        provider: "custom"
        config:
          vector_store: "postgresql"
          embedding_model: "text-embedding-ada-002"
    
    retriever:
      provider: "vector"
      config:
        top_k: 5
        similarity_threshold: 0.8
  
  flows:
    intent_classification:
      timeout: "30s"
      retry_count: 3
    
    problem_decompose:
      timeout: "120s"
      retry_count: 2
      max_subtasks: 10
    
    synthesis:
      timeout: "60s"
      retry_count: 2
      quality_threshold: 0.8
```

### 5.2 组件初始化

```go
// 组件工厂
type ComponentFactory struct {
    config *Config
}

func NewComponentFactory(config *Config) *ComponentFactory {
    return &ComponentFactory{config: config}
}

func (f *ComponentFactory) CreateComponents(ctx context.Context) (*ThinkingComponents, error) {
    // 创建ChatModel
    chatModel, err := f.createChatModel(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to create chat model: %w", err)
    }
    
    // 创建ChatTemplate
    chatTemplate := f.createChatTemplate()
    
    // 创建ToolsNode
    toolsNode, err := f.createToolsNode(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to create tools node: %w", err)
    }
    
    // 创建Retriever
    retriever, err := f.createRetriever(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to create retriever: %w", err)
    }
    
    // 创建自定义Lambda组件
    intentClassifier := f.createIntentClassifier()
    problemAnalyzer := f.createProblemAnalyzer()
    taskDecomposer := f.createTaskDecomposer()
    evidenceIntegrator := f.createEvidenceIntegrator()
    qualityValidator := f.createQualityValidator()
    
    return &ThinkingComponents{
        ChatModel:           chatModel,
        ChatTemplate:        chatTemplate,
        ToolsNode:          toolsNode,
        Retriever:          retriever,
        IntentClassifier:   intentClassifier,
        ProblemAnalyzer:    problemAnalyzer,
        TaskDecomposer:     taskDecomposer,
        EvidenceIntegrator: evidenceIntegrator,
        QualityValidator:   qualityValidator,
    }, nil
}
```

## 6. 监控和可观测性

### 6.1 指标收集

```go
// 指标收集器
type MetricsCollector struct {
    executionCount    prometheus.CounterVec
    executionDuration prometheus.HistogramVec
    errorCount        prometheus.CounterVec
}

func NewMetricsCollector() *MetricsCollector {
    return &MetricsCollector{
        executionCount: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "eino_component_executions_total",
                Help: "Total number of component executions",
            },
            []string{"component", "flow", "status"},
        ),
        executionDuration: prometheus.NewHistogramVec(
            prometheus.HistogramOpts{
                Name: "eino_component_duration_seconds",
                Help: "Component execution duration",
            },
            []string{"component", "flow"},
        ),
        errorCount: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "eino_component_errors_total",
                Help: "Total number of component errors",
            },
            []string{"component", "flow", "error_type"},
        ),
    }
}

// 监控回调处理器
func (mc *MetricsCollector) CreateMonitoringHandler() eino.CallbackHandler {
    return eino.NewHandlerBuilder().
        OnStartFn(func(ctx context.Context, info *eino.RunInfo, input eino.CallbackInput) context.Context {
            start := time.Now()
            return context.WithValue(ctx, "start_time", start)
        }).
        OnEndFn(func(ctx context.Context, info *eino.RunInfo, output eino.CallbackOutput) context.Context {
            if start, ok := ctx.Value("start_time").(time.Time); ok {
                duration := time.Since(start)
                mc.executionDuration.WithLabelValues(info.NodeName, info.FlowName).Observe(duration.Seconds())
            }
            mc.executionCount.WithLabelValues(info.NodeName, info.FlowName, "success").Inc()
            return ctx
        }).
        OnErrorFn(func(ctx context.Context, info *eino.RunInfo, err error) context.Context {
            mc.executionCount.WithLabelValues(info.NodeName, info.FlowName, "error").Inc()
            mc.errorCount.WithLabelValues(info.NodeName, info.FlowName, "execution_error").Inc()
            return ctx
        }).
        Build()
}
```

### 6.2 分布式追踪

```go
// OpenTelemetry集成
func CreateTracingHandler() eino.CallbackHandler {
    return eino.NewHandlerBuilder().
        OnStartFn(func(ctx context.Context, info *eino.RunInfo, input eino.CallbackInput) context.Context {
            tracer := otel.Tracer("thinking-map")
            ctx, span := tracer.Start(ctx, fmt.Sprintf("eino.%s", info.NodeName))
            
            span.SetAttributes(
                attribute.String("component.name", info.NodeName),
                attribute.String("flow.name", info.FlowName),
                attribute.String("input.type", fmt.Sprintf("%T", input)),
            )
            
            return ctx
        }).
        OnEndFn(func(ctx context.Context, info *eino.RunInfo, output eino.CallbackOutput) context.Context {
            if span := trace.SpanFromContext(ctx); span.IsRecording() {
                span.SetAttributes(
                    attribute.String("output.type", fmt.Sprintf("%T", output)),
                    attribute.String("status", "success"),
                )
                span.End()
            }
            return ctx
        }).
        OnErrorFn(func(ctx context.Context, info *eino.RunInfo, err error) context.Context {
            if span := trace.SpanFromContext(ctx); span.IsRecording() {
                span.RecordError(err)
                span.SetStatus(codes.Error, err.Error())
                span.End()
            }
            return ctx
        }).
        Build()
}
```

## 7. 测试策略

### 7.1 单元测试

```go
// 组件单元测试
func TestIntentClassifier(t *testing.T) {
    // 创建模拟组件
    mockChatModel := &MockChatModel{}
    mockTemplate := &MockChatTemplate{}
    
    // 创建测试链
    chain, err := eino.NewChain[map[string]any, *IntentResult]().
        AppendChatTemplate(mockTemplate).
        AppendChatModel(mockChatModel).
        AppendLambda(NewIntentClassifier()).
        Compile(context.Background())
    
    require.NoError(t, err)
    
    // 测试用例
    testCases := []struct {
        name     string
        input    map[string]any
        expected *IntentResult
    }{
        {
            name: "research intent",
            input: map[string]any{
                "query": "如何提高机器学习模型的准确率？",
            },
            expected: &IntentResult{
                Intent:     "research",
                Confidence: 0.9,
            },
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            result, err := chain.Invoke(context.Background(), tc.input)
            require.NoError(t, err)
            assert.Equal(t, tc.expected.Intent, result.Intent)
            assert.GreaterOrEqual(t, result.Confidence, 0.8)
        })
    }
}
```

### 7.2 集成测试

```go
// 端到端流程测试
func TestDecomposeFlow(t *testing.T) {
    // 设置测试环境
    config := &Config{
        // 测试配置
    }
    
    factory := NewComponentFactory(config)
    components, err := factory.CreateComponents(context.Background())
    require.NoError(t, err)
    
    graph, err := NewDecomposeGraph(components)
    require.NoError(t, err)
    
    // 执行测试
    input := &ProblemInput{
        Question: "如何设计一个高可用的微服务架构？",
        Target:   "提供详细的架构设计方案",
    }
    
    result, err := graph.Invoke(context.Background(), input)
    require.NoError(t, err)
    
    // 验证结果
    assert.NotEmpty(t, result.SubTasks)
    assert.True(t, len(result.SubTasks) > 0)
    assert.NotEmpty(t, result.ExecutionPlan)
}
```

## 8. 性能优化

### 8.1 并发处理

```go
// 并发执行优化
func NewParallelDecomposeGraph(components *ThinkingComponents) (*eino.CompiledGraph, error) {
    graph := eino.NewGraph[*ProblemInput, *DecomposeResult]()
    
    // 并行分析节点
    graph.AddLambdaNode("domain_analyzer", components.DomainAnalyzer)
    graph.AddLambdaNode("complexity_analyzer", components.ComplexityAnalyzer)
    graph.AddLambdaNode("resource_analyzer", components.ResourceAnalyzer)
    
    // 汇聚节点
    graph.AddLambdaNode("analysis_merger", components.AnalysisMerger)
    
    // 并行执行
    graph.AddEdge(eino.START, "domain_analyzer")
    graph.AddEdge(eino.START, "complexity_analyzer")
    graph.AddEdge(eino.START, "resource_analyzer")
    
    // 等待所有分析完成
    graph.AddEdge("domain_analyzer", "analysis_merger")
    graph.AddEdge("complexity_analyzer", "analysis_merger")
    graph.AddEdge("resource_analyzer", "analysis_merger")
    
    return graph.Compile(ctx)
}
```

### 8.2 缓存策略

```go
// 缓存装饰器
type CachedComponent struct {
    component eino.ChatModel
    cache     cache.Cache
    ttl       time.Duration
}

func WithCache(component eino.ChatModel, cache cache.Cache, ttl time.Duration) eino.ChatModel {
    return &CachedComponent{
        component: component,
        cache:     cache,
        ttl:       ttl,
    }
}

func (c *CachedComponent) Generate(ctx context.Context, input []*schema.Message, opts ...eino.Option) (*schema.Message, error) {
    // 生成缓存键
    key := c.generateCacheKey(input, opts)
    
    // 尝试从缓存获取
    if cached, found := c.cache.Get(key); found {
        if result, ok := cached.(*schema.Message); ok {
            return result, nil
        }
    }
    
    // 缓存未命中，执行组件
    result, err := c.component.Generate(ctx, input, opts...)
    if err != nil {
        return nil, err
    }
    
    // 存储到缓存
    c.cache.Set(key, result, c.ttl)
    
    return result, nil
}
```

## 9. 总结

### 9.1 技术优势

1. **类型安全**：Eino的泛型设计确保编译时类型检查
2. **流式处理**：原生支持LLM流式响应，提升用户体验
3. **组件化**：高度模块化的设计，便于测试和维护
4. **可观测性**：内置的回调机制提供完整的执行追踪
5. **性能优化**：并发管理和流处理优化

### 9.2 实施建议

1. **渐进式迁移**：逐步将现有Agent替换为Eino组件
2. **充分测试**：建立完善的单元测试和集成测试
3. **监控完善**：部署完整的监控和告警系统
4. **文档维护**：保持技术文档和API文档的更新

### 9.3 后续规划

1. **组件扩展**：开发更多专用的业务组件
2. **性能调优**：基于实际使用情况进行性能优化
3. **功能增强**：添加更多高级特性如动态路由、自适应重试等
4. **生态集成**：与更多外部服务和工具集成

通过采用Eino框架，ThinkingMap的Agent系统将获得更好的可维护性、可扩展性和性能表现，为用户提供更优质的AI辅助思考体验。