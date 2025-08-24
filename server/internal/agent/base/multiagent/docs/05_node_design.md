# Eino Enhanced MultiAgent 节点设计与实现

## 节点设计与类型对齐策略

### 核心设计原则

#### 1. 灵活的类型对齐策略
- **首节点输入**: 第一个节点从上层接收 `[]*schema.Message` 类型的消息
- **LLM 节点**: 依赖大语言模型服务的节点，输入为 `[]*schema.Message`，输出为 `*schema.Message`
- **Lambda 节点**: 根据具体业务逻辑需求设定输入输出类型，无需强制统一
- **状态传递**: 通过 `StatePreHandler` 和 `StatePostHandler` 处理状态，主要通过 PreHandler 设置入参，PostHandler 解析结果并保存到全局状态

### 具体节点类型定义

#### Host Think Node
- **输入**: `[]*schema.Message` (从上层接收的对话消息)
- **输出**: `*schema.Message` (LLM 生成的思考结果)

#### Direct Answer Node
- **输入**: `[]*schema.Message` (从上层接收的对话消息)
- **输出**: `*schema.Message` (直接回答结果)

#### Complexity Branch
- **输入**: `TaskComplexity` (从 EnhancedState 中获取的复杂度评估)
- **输出**: `bool` (复杂度判断结果，通过 Lambda 节点实现)

#### Plan Creation Node
- **输入**: `[]*schema.Message` (包含任务描述和上下文的消息)
- **输出**: `*schema.Message` (包含 JSON 格式规划的消息)

#### Plan Execution Node
- **输入**: `*TaskPlan` (从 EnhancedState 中获取的当前执行计划)
- **输出**: `map[string]*SpecialistResult` (专家执行结果映射，通过 Lambda 节点实现)

#### Specialist Nodes
- **输入**: `[]*schema.Message` (专家任务描述和上下文)
- **输出**: `*schema.Message` (专家执行结果)

### 7. Specialist Handler

```go
// SpecialistHandler 专家节点处理器
type SpecialistHandler struct {
    config       *EnhancedMultiAgentConfig
    specialistName string
}

// PreHandle 准备专家执行输入
func (h *SpecialistHandler) PreHandle(ctx context.Context, state *EnhancedState, input []*schema.Message) error {
    if state.CurrentExecution == nil || state.CurrentExecution.PlanStep == nil {
        return errors.New("no current execution step available")
    }
    
    step := state.CurrentExecution.PlanStep
    
    // 构建专家执行提示
    specialistPrompt := h.buildSpecialistPrompt(step, state)
    
    // 更新输入消息
    if len(input) > 0 {
        input[0] = specialistPrompt
    } else {
        input = append(input, specialistPrompt)
    }
    
    return nil
}

// PostHandle 处理专家执行结果
func (h *SpecialistHandler) PostHandle(ctx context.Context, state *EnhancedState, output *schema.Message) error {
    if state.CurrentExecution == nil || state.CurrentExecution.PlanStep == nil {
        return errors.New("no current execution step available")
    }
    
    step := state.CurrentExecution.PlanStep
    
    // 解析专家执行结果
    specialistResult := &SpecialistResult{
        StepID:         step.ID,
        SpecialistName: step.AssignedTo,
        Content:        output.Content,
        Success:        true, // 根据实际结果判断
        Confidence:     0.8,  // 根据实际结果计算
        ExecutionTime:  time.Since(*step.StartTime),
        Timestamp:      time.Now(),
    }
    
    // 更新状态
    if state.SpecialistResults == nil {
        state.SpecialistResults = make(map[string]*SpecialistResult)
    }
    state.SpecialistResults[step.ID] = specialistResult
    step.Status = StepStatusCompleted
    endTime := time.Now()
    step.EndTime = &endTime
    state.LastUpdateTime = time.Now()
    
    return nil
}

// buildSpecialistPrompt 构建专家执行提示
func (h *SpecialistHandler) buildSpecialistPrompt(step *PlanStep, state *EnhancedState) *schema.Message {
    var promptBuilder strings.Builder
    
    // 获取专家配置
    specialist, exists := h.config.Specialists[step.AssignedTo]
    if !exists {
        specialist = &SpecialistConfig{
            SystemPrompt: "你是一个专业的AI助手，请根据任务要求提供帮助。",
        }
    }
    
    // 专家系统提示
    promptBuilder.WriteString(specialist.SystemPrompt)
    promptBuilder.WriteString("\n\n")
    
    // 任务步骤信息
    promptBuilder.WriteString("任务步骤信息：\n")
    promptBuilder.WriteString(fmt.Sprintf("- 步骤ID: %s\n", step.ID))
    promptBuilder.WriteString(fmt.Sprintf("- 步骤名称: %s\n", step.Name))
    promptBuilder.WriteString(fmt.Sprintf("- 步骤描述: %s\n", step.Description))
    promptBuilder.WriteString(fmt.Sprintf("- 优先级: %d\n", step.Priority))
    
    // 步骤参数
    if len(step.Parameters) > 0 {
        promptBuilder.WriteString("- 参数:\n")
        for key, value := range step.Parameters {
            promptBuilder.WriteString(fmt.Sprintf("  - %s: %v\n", key, value))
        }
    }
    
    promptBuilder.WriteString("\n")
    
    // 上下文信息
    if state.ConversationContext != nil {
        promptBuilder.WriteString("对话上下文：\n")
        promptBuilder.WriteString(fmt.Sprintf("- 用户意图: %s\n", state.ConversationContext.UserIntent))
        if state.ConversationContext.ContextSummary != "" {
            promptBuilder.WriteString(fmt.Sprintf("- 上下文摘要: %s\n", state.ConversationContext.ContextSummary))
        }
        promptBuilder.WriteString("\n")
    }
    
    // 原始消息
    if len(state.OriginalMessages) > 0 {
        promptBuilder.WriteString("原始用户请求：\n")
        for i, msg := range state.OriginalMessages {
            role := "用户"
            if msg.Role == schema.RoleAssistant {
                role = "助手"
            }
            promptBuilder.WriteString(fmt.Sprintf("%d. %s: %s\n", i+1, role, msg.Content))
        }
        promptBuilder.WriteString("\n")
    }
    
    // 执行要求
    promptBuilder.WriteString("请根据以上信息执行分配给你的任务步骤，并提供详细的执行结果。")
    
    return &schema.Message{
        Role:    schema.RoleUser,
        Content: promptBuilder.String(),
    }
}
```

### 8. Result Collector Lambda

```go
// ResultCollectorLambda 结果收集Lambda节点
func ResultCollectorLambda(ctx context.Context, state *EnhancedState) (*CollectedResults, error) {
    results := state.SpecialistResults
    
    if len(results) == 0 {
        return nil, errors.New("no specialist results available")
    }
    
    collected := &CollectedResults{
        Results:           results,
        OverallConfidence: calculateConfidenceScore(results),
        QualityScore:      calculateQualityScore(results),
        Summary:           generateResultSummary(results),
        Timestamp:         time.Now(),
    }
    
    // 更新状态
    state.CollectedResults = collected
    state.LastUpdateTime = time.Now()
    
    return collected, nil
}

// calculateConfidenceScore 计算整体置信度
func calculateConfidenceScore(results map[string]*SpecialistResult) float64 {
    if len(results) == 0 {
        return 0.0
    }
    
    var totalConfidence float64
    for _, result := range results {
        totalConfidence += result.Confidence
    }
    
    return totalConfidence / float64(len(results))
}

// calculateQualityScore 计算质量分数
func calculateQualityScore(results map[string]*SpecialistResult) float64 {
    if len(results) == 0 {
        return 0.0
    }
    
    var successCount int
    for _, result := range results {
        if result.Success {
            successCount++
        }
    }
    
    return float64(successCount) / float64(len(results))
}

// generateResultSummary 生成结果摘要
func generateResultSummary(results map[string]*SpecialistResult) string {
    var summaryBuilder strings.Builder
    
    summaryBuilder.WriteString(fmt.Sprintf("执行了 %d 个专家任务：\n", len(results)))
    
    for stepID, result := range results {
        status := "成功"
        if !result.Success {
            status = "失败"
        }
        summaryBuilder.WriteString(fmt.Sprintf("- %s (%s): %s (置信度: %.2f)\n", 
            stepID, result.SpecialistName, status, result.Confidence))
    }
    
    return summaryBuilder.String()
}
```

### 9. Feedback Processor Handler

```go
// FeedbackProcessorHandler 反馈处理节点处理器
type FeedbackProcessorHandler struct {
    config *EnhancedMultiAgentConfig
}

// PreHandle 准备反馈分析输入
func (h *FeedbackProcessorHandler) PreHandle(ctx context.Context, state *EnhancedState, input []*schema.Message) error {
    if state.CollectedResults == nil {
        return errors.New("no collected results available for feedback analysis")
    }
    
    feedbackPrompt := h.buildFeedbackPrompt(state)
    
    if len(input) > 0 {
        input[0] = feedbackPrompt
    } else {
        input = append(input, feedbackPrompt)
    }
    
    return nil
}

// PostHandle 处理反馈分析结果
func (h *FeedbackProcessorHandler) PostHandle(ctx context.Context, state *EnhancedState, output *schema.Message) error {
    // 解析反馈结果
    feedback, err := h.parseFeedbackResult(output)
    if err != nil {
        return fmt.Errorf("failed to parse feedback result: %w", err)
    }
    
    // 更新状态
    state.FeedbackResult = feedback
    state.LastUpdateTime = time.Now()
    
    return nil
}

// buildFeedbackPrompt 构建反馈分析提示
func (h *FeedbackProcessorHandler) buildFeedbackPrompt(state *EnhancedState) *schema.Message {
    var promptBuilder strings.Builder
    
    promptBuilder.WriteString("你是一个任务执行质量分析专家，需要分析当前的执行结果并提供反馈。\n\n")
    
    // 收集的结果信息
    collected := state.CollectedResults
    promptBuilder.WriteString("执行结果分析：\n")
    promptBuilder.WriteString(fmt.Sprintf("- 整体置信度: %.2f\n", collected.OverallConfidence))
    promptBuilder.WriteString(fmt.Sprintf("- 质量分数: %.2f\n", collected.QualityScore))
    promptBuilder.WriteString(fmt.Sprintf("- 结果摘要: %s\n", collected.Summary))
    promptBuilder.WriteString("\n")
    
    // 详细结果
    promptBuilder.WriteString("详细执行结果：\n")
    for stepID, result := range collected.Results {
        promptBuilder.WriteString(fmt.Sprintf("步骤 %s (%s):\n", stepID, result.SpecialistName))
        promptBuilder.WriteString(fmt.Sprintf("  - 成功: %t\n", result.Success))
        promptBuilder.WriteString(fmt.Sprintf("  - 置信度: %.2f\n", result.Confidence))
        promptBuilder.WriteString(fmt.Sprintf("  - 执行时间: %v\n", result.ExecutionTime))
        promptBuilder.WriteString(fmt.Sprintf("  - 内容: %s\n", result.Content))
        promptBuilder.WriteString("\n")
    }
    
    // 当前计划信息
    if state.CurrentPlan != nil {
        promptBuilder.WriteString("当前计划状态：\n")
        promptBuilder.WriteString(fmt.Sprintf("- 计划名称: %s\n", state.CurrentPlan.Name))
        promptBuilder.WriteString(fmt.Sprintf("- 总步骤数: %d\n", state.CurrentPlan.TotalSteps))
        promptBuilder.WriteString(fmt.Sprintf("- 当前步骤: %d\n", state.CurrentPlan.CurrentStep))
        promptBuilder.WriteString("\n")
    }
    
    promptBuilder.WriteString("请分析执行结果并以JSON格式提供反馈：\n")
    promptBuilder.WriteString("```json\n")
    promptBuilder.WriteString("{\n")
    promptBuilder.WriteString("  \"overall_quality\": \"excellent|good|fair|poor\",\n")
    promptBuilder.WriteString("  \"should_continue\": true|false,\n")
    promptBuilder.WriteString("  \"issues_identified\": [\"问题描述\"],\n")
    promptBuilder.WriteString("  \"improvement_suggestions\": [\"改进建议\"],\n")
    promptBuilder.WriteString("  \"confidence_in_results\": 0.0-1.0,\n")
    promptBuilder.WriteString("  \"next_actions\": [\"下一步行动\"]\n")
    promptBuilder.WriteString("}\n")
    promptBuilder.WriteString("```")
    
    return &schema.Message{
        Role:    schema.RoleUser,
        Content: promptBuilder.String(),
    }
}

// parseFeedbackResult 解析反馈结果
func (h *FeedbackProcessorHandler) parseFeedbackResult(output *schema.Message) (*FeedbackResult, error) {
    content := output.Content
    start := strings.Index(content, "```json")
    end := strings.LastIndex(content, "```")
    
    if start == -1 || end == -1 || start >= end {
        return nil, errors.New("no valid JSON found in feedback result")
    }
    
    jsonStr := content[start+7 : end]
    
    var feedbackData struct {
        OverallQuality          string   `json:"overall_quality"`
        ShouldContinue          bool     `json:"should_continue"`
        IssuesIdentified        []string `json:"issues_identified"`
        ImprovementSuggestions  []string `json:"improvement_suggestions"`
        ConfidenceInResults     float64  `json:"confidence_in_results"`
        NextActions             []string `json:"next_actions"`
    }
    
    if err := json.Unmarshal([]byte(jsonStr), &feedbackData); err != nil {
        return nil, fmt.Errorf("failed to parse feedback JSON: %w", err)
    }
    
    return &FeedbackResult{
        OverallQuality:         feedbackData.OverallQuality,
        ShouldContinue:         feedbackData.ShouldContinue,
        IssuesIdentified:       feedbackData.IssuesIdentified,
        ImprovementSuggestions: feedbackData.ImprovementSuggestions,
        ConfidenceInResults:    feedbackData.ConfidenceInResults,
        NextActions:            feedbackData.NextActions,
        Timestamp:              time.Now(),
    }, nil
}
```

### 10. Plan Update Handler

```go
// PlanUpdateHandler 规划更新节点处理器
type PlanUpdateHandler struct {
    config *EnhancedMultiAgentConfig
}

// PreHandle 准备规划更新输入
func (h *PlanUpdateHandler) PreHandle(ctx context.Context, state *EnhancedState, input []*schema.Message) error {
    if state.FeedbackResult == nil {
        return errors.New("no feedback result available for plan update")
    }
    
    updatePrompt := h.buildPlanUpdatePrompt(state)
    
    if len(input) > 0 {
        input[0] = updatePrompt
    } else {
        input = append(input, updatePrompt)
    }
    
    return nil
}

// PostHandle 处理规划更新结果
func (h *PlanUpdateHandler) PostHandle(ctx context.Context, state *EnhancedState, output *schema.Message) error {
    // 解析更新后的规划
    updatedPlan, err := h.parsePlanFromMessage(output)
    if err != nil {
        return fmt.Errorf("failed to parse updated plan: %w", err)
    }
    
    // 创建规划更新记录
    update := &PlanUpdate{
        ID:          generateUpdateID(),
        PlanID:      state.CurrentPlan.ID,
        UpdateType:  PlanUpdateTypeModification,
        Description: "根据反馈结果更新规划",
        Changes:     h.calculatePlanChanges(state.CurrentPlan, updatedPlan),
        Timestamp:   time.Now(),
    }
    
    // 更新状态
    if state.PlanHistory == nil {
        state.PlanHistory = make([]*TaskPlan, 0)
    }
    if state.PlanUpdates == nil {
        state.PlanUpdates = make([]*PlanUpdate, 0)
    }
    
    state.PlanHistory = append(state.PlanHistory, state.CurrentPlan)
    state.CurrentPlan = updatedPlan
    state.PlanUpdates = append(state.PlanUpdates, update)
    state.LastUpdateTime = time.Now()
    
    return nil
}

// buildPlanUpdatePrompt 构建规划更新提示
func (h *PlanUpdateHandler) buildPlanUpdatePrompt(state *EnhancedState) *schema.Message {
    var promptBuilder strings.Builder
    
    promptBuilder.WriteString("你是一个任务规划优化专家，需要根据反馈结果更新任务规划。\n\n")
    
    // 反馈结果
    feedback := state.FeedbackResult
    promptBuilder.WriteString("反馈分析结果：\n")
    promptBuilder.WriteString(fmt.Sprintf("- 整体质量: %s\n", feedback.OverallQuality))
    promptBuilder.WriteString(fmt.Sprintf("- 是否继续: %t\n", feedback.ShouldContinue))
    promptBuilder.WriteString(fmt.Sprintf("- 结果置信度: %.2f\n", feedback.ConfidenceInResults))
    
    if len(feedback.IssuesIdentified) > 0 {
        promptBuilder.WriteString("- 发现的问题:\n")
        for _, issue := range feedback.IssuesIdentified {
            promptBuilder.WriteString(fmt.Sprintf("  - %s\n", issue))
        }
    }
    
    if len(feedback.ImprovementSuggestions) > 0 {
        promptBuilder.WriteString("- 改进建议:\n")
        for _, suggestion := range feedback.ImprovementSuggestions {
            promptBuilder.WriteString(fmt.Sprintf("  - %s\n", suggestion))
        }
    }
    
    promptBuilder.WriteString("\n")
    
    // 当前规划
    currentPlan := state.CurrentPlan
    promptBuilder.WriteString("当前规划：\n")
    promptBuilder.WriteString(fmt.Sprintf("- 名称: %s\n", currentPlan.Name))
    promptBuilder.WriteString(fmt.Sprintf("- 描述: %s\n", currentPlan.Description))
    promptBuilder.WriteString("- 步骤:\n")
    for i, step := range currentPlan.Steps {
        promptBuilder.WriteString(fmt.Sprintf("  %d. %s (%s) - %s\n", 
            i+1, step.Name, step.AssignedTo, step.Status.String()))
    }
    promptBuilder.WriteString("\n")
    
    promptBuilder.WriteString("请根据反馈结果优化规划，以JSON格式输出更新后的规划：\n")
    promptBuilder.WriteString("```json\n")
    promptBuilder.WriteString("{\n")
    promptBuilder.WriteString("  \"name\": \"更新后的计划名称\",\n")
    promptBuilder.WriteString("  \"description\": \"更新后的计划描述\",\n")
    promptBuilder.WriteString("  \"steps\": [\n")
    promptBuilder.WriteString("    {\n")
    promptBuilder.WriteString("      \"id\": \"stepID\",\n")
    promptBuilder.WriteString("      \"name\": \"步骤名称\",\n")
    promptBuilder.WriteString("      \"description\": \"步骤描述\",\n")
    promptBuilder.WriteString("      \"assigned_to\": \"专家名称\",\n")
    promptBuilder.WriteString("      \"priority\": 1,\n")
    promptBuilder.WriteString("      \"dependencies\": [],\n")
    promptBuilder.WriteString("      \"parameters\": {}\n")
    promptBuilder.WriteString("    }\n")
    promptBuilder.WriteString("  ]\n")
    promptBuilder.WriteString("}\n")
    promptBuilder.WriteString("```")
    
    return &schema.Message{
        Role:    schema.RoleUser,
        Content: promptBuilder.String(),
    }
}

// parsePlanFromMessage 从消息中解析规划（复用之前的实现）
func (h *PlanUpdateHandler) parsePlanFromMessage(output *schema.Message) (*TaskPlan, error) {
    // 实现与 PlanCreationHandler 中的 parsePlanFromMessage 相同
    // 这里省略具体实现，参考前面的代码
    return nil, nil // 占位符
}

// calculatePlanChanges 计算规划变更
func (h *PlanUpdateHandler) calculatePlanChanges(oldPlan, newPlan *TaskPlan) []string {
    var changes []string
    
    if oldPlan.Name != newPlan.Name {
        changes = append(changes, fmt.Sprintf("名称变更: %s -> %s", oldPlan.Name, newPlan.Name))
    }
    
    if oldPlan.Description != newPlan.Description {
        changes = append(changes, "描述已更新")
    }
    
    if len(oldPlan.Steps) != len(newPlan.Steps) {
        changes = append(changes, fmt.Sprintf("步骤数量变更: %d -> %d", len(oldPlan.Steps), len(newPlan.Steps)))
    }
    
    return changes
}
```

### 11. Reflection Branch Lambda

```go
// ReflectionBranchLambda 反思决策Lambda节点
func ReflectionBranchLambda(ctx context.Context, state *EnhancedState) (bool, error) {
    if state.FeedbackResult == nil {
        return false, errors.New("no feedback result available for reflection")
    }
    
    feedback := state.FeedbackResult
    
    // 决策逻辑
    shouldContinue := feedback.ShouldContinue && 
                     state.CurrentRound < state.MaxRounds &&
                     !state.IsCompleted
    
    // 更新轮次
    if shouldContinue {
        state.CurrentRound++
    }
    
    return shouldContinue, nil
}
```

### 12. Final Answer Handler

```go
// FinalAnswerHandler 最终答案节点处理器
type FinalAnswerHandler struct {
    config *EnhancedMultiAgentConfig
}

// PreHandle 准备最终答案生成输入
func (h *FinalAnswerHandler) PreHandle(ctx context.Context, state *EnhancedState, input []*schema.Message) error {
    finalPrompt := h.buildFinalAnswerPrompt(state)
    
    if len(input) > 0 {
        input[0] = finalPrompt
    } else {
        input = append(input, finalPrompt)
    }
    
    return nil
}

// PostHandle 处理最终答案结果
func (h *FinalAnswerHandler) PostHandle(ctx context.Context, state *EnhancedState, output *schema.Message) error {
    // 更新最终状态
    state.FinalAnswer = output
    state.IsCompleted = true
    state.EndTime = time.Now()
    state.LastUpdateTime = time.Now()
    
    return nil
}

// buildFinalAnswerPrompt 构建最终答案生成提示
func (h *FinalAnswerHandler) buildFinalAnswerPrompt(state *EnhancedState) *schema.Message {
    var promptBuilder strings.Builder
    
    promptBuilder.WriteString("你是一个智能助手，需要基于所有执行结果为用户生成最终的回答。\n\n")
    
    // 原始用户请求
    if len(state.OriginalMessages) > 0 {
        promptBuilder.WriteString("用户原始请求：\n")
        for i, msg := range state.OriginalMessages {
            if msg.Role == schema.RoleUser {
                promptBuilder.WriteString(fmt.Sprintf("%d. %s\n", i+1, msg.Content))
            }
        }
        promptBuilder.WriteString("\n")
    }
    
    // 执行结果摘要
    if state.CollectedResults != nil {
        promptBuilder.WriteString("执行结果摘要：\n")
        promptBuilder.WriteString(state.CollectedResults.Summary)
        promptBuilder.WriteString("\n")
        
        // 详细结果
        promptBuilder.WriteString("详细执行结果：\n")
        for stepID, result := range state.CollectedResults.Results {
            promptBuilder.WriteString(fmt.Sprintf("- %s (%s): %s\n", 
                stepID, result.SpecialistName, result.Content))
        }
        promptBuilder.WriteString("\n")
    }
    
    // 对话上下文
    if state.ConversationContext != nil && state.ConversationContext.IsContinuation {
        promptBuilder.WriteString("注意：这是延续对话，请确保回答与之前的对话保持连贯。\n\n")
    }
    
    promptBuilder.WriteString("请基于以上信息为用户生成清晰、准确、有帮助的最终回答。")
    
    return &schema.Message{
        Role:    schema.RoleUser,
        Content: promptBuilder.String(),
    }
}
```

## 节点类型对齐策略

### 类型对齐原则

根据 [Eino Chain & Graph 编排规范](https://www.cloudwego.io/zh/docs/eino/core_modules/chain_and_graph_orchestration/)，我们采用以下类型对齐策略：

#### 1. LLM 节点类型
- **适用节点**: Host Think Node, Direct Answer Node, Plan Creation Node, Specialist Nodes, Feedback Processor Node, Plan Update Node, Final Answer Node
- **输入类型**: `[]*schema.Message`
- **输出类型**: `*schema.Message`
- **特点**: 直接与大语言模型交互，无需额外转换

#### 2. Lambda 节点类型
- **适用节点**: Complexity Branch, Plan Execution Node, Result Collector Node, Reflection Branch
- **输入类型**: 根据业务逻辑需求自定义
- **输出类型**: 根据业务逻辑需求自定义
- **特点**: 纯逻辑处理，通过 PreHandler 和 PostHandler 与全局状态交互

#### 3. 混合节点类型
- **适用节点**: Plan Execution Node (包含多个 Specialist 子节点)
- **输入类型**: 自定义（如 `*TaskPlan`）
- **输出类型**: 自定义（如 `map[string]*SpecialistResult`）
- **特点**: 内部包含多个子节点，通过编排实现复杂逻辑

### 编排示例

```go
// 创建 Enhanced MultiAgent Graph
func CreateEnhancedMultiAgentGraph(config *EnhancedMultiAgentConfig) (*graph.Graph[*schema.Message, *schema.Message], error) {
    g := graph.New[*schema.Message, *schema.Message]()
    
    // 1. Host Think Node (LLM节点)
    hostThinkNode := chatmodel.NewChatModel(config.Host.ThinkingModel)
    g.AddChatModelNode("host_think", hostThinkNode,
        graph.WithNodePreHandler(NewHostThinkHandler(config).PreHandle),
        graph.WithNodePostHandler(NewHostThinkHandler(config).PostHandle),
    )
    
    // 2. Complexity Branch (Lambda节点)
    g.AddLambdaNode("complexity_branch", 
        func(ctx context.Context, state *EnhancedState) (bool, error) {
            return ComplexityBranchLambda(ctx, state)
        },
        graph.WithNodePreHandler(func(ctx context.Context, state *EnhancedState, input []*schema.Message) error {
            // 从状态中获取复杂度评估结果
            return nil
        }),
    )
    
    // 3. Direct Answer Node (LLM节点)
    directAnswerNode := chatmodel.NewChatModel(config.Host.ChatModel)
    g.AddChatModelNode("direct_answer", directAnswerNode,
        graph.WithNodePreHandler(NewDirectAnswerHandler(config).PreHandle),
        graph.WithNodePostHandler(NewDirectAnswerHandler(config).PostHandle),
    )
    
    // 4. Plan Creation Node (LLM节点)
    planCreationNode := chatmodel.NewChatModel(config.Host.ChatModel)
    g.AddChatModelNode("plan_creation", planCreationNode,
        graph.WithNodePreHandler(NewPlanCreationHandler(config).PreHandle),
        graph.WithNodePostHandler(NewPlanCreationHandler(config).PostHandle),
    )
    
    // 5. Plan Execution Node (Lambda节点，内部包含多个Specialist)
    g.AddLambdaNode("plan_execution",
        func(ctx context.Context, state *EnhancedState) (map[string]*SpecialistResult, error) {
            return PlanExecutionLambda(ctx, state, config)
        },
        graph.WithNodePreHandler(func(ctx context.Context, state *EnhancedState, input []*schema.Message) error {
            // 从状态中获取当前执行计划
            return nil
        }),
    )
    
    // 6. Result Collector Node (Lambda节点)
    g.AddLambdaNode("result_collector",
        func(ctx context.Context, state *EnhancedState) (*CollectedResults, error) {
            return ResultCollectorLambda(ctx, state)
        },
    )
    
    // 7. Feedback Processor Node (LLM节点)
    feedbackNode := chatmodel.NewChatModel(config.Host.ChatModel)
    g.AddChatModelNode("feedback_processor", feedbackNode,
        graph.WithNodePreHandler(NewFeedbackProcessorHandler(config).PreHandle),
        graph.WithNodePostHandler(NewFeedbackProcessorHandler(config).PostHandle),
    )
    
    // 8. Plan Update Node (LLM节点)
    planUpdateNode := chatmodel.NewChatModel(config.Host.ChatModel)
    g.AddChatModelNode("plan_update", planUpdateNode,
        graph.WithNodePreHandler(NewPlanUpdateHandler(config).PreHandle),
        graph.WithNodePostHandler(NewPlanUpdateHandler(config).PostHandle),
    )
    
    // 9. Reflection Branch (Lambda节点)
    g.AddLambdaNode("reflection_branch",
        func(ctx context.Context, state *EnhancedState) (bool, error) {
            return ReflectionBranchLambda(ctx, state)
        },
    )
    
    // 10. Final Answer Node (LLM节点)
    finalAnswerNode := chatmodel.NewChatModel(config.Host.ChatModel)
    g.AddChatModelNode("final_answer", finalAnswerNode,
        graph.WithNodePreHandler(NewFinalAnswerHandler(config).PreHandle),
        graph.WithNodePostHandler(NewFinalAnswerHandler(config).PostHandle),
    )
    
    // 添加边连接
    g.AddEdge(graph.START, "host_think")
    g.AddEdge("host_think", "complexity_branch")
    
    // 复杂度分支
    g.AddConditionalEdges("complexity_branch", map[bool]string{
        false: "direct_answer",  // 简单任务直接回答
        true:  "plan_creation",  // 复杂任务创建规划
    })
    
    g.AddEdge("direct_answer", graph.END)
    g.AddEdge("plan_creation", "plan_execution")
    g.AddEdge("plan_execution", "result_collector")
    g.AddEdge("result_collector", "feedback_processor")
    g.AddEdge("feedback_processor", "reflection_branch")
    
    // 反思分支
    g.AddConditionalEdges("reflection_branch", map[bool]string{
        true:  "plan_update",    // 继续执行，更新规划
        false: "final_answer",   // 结束执行，生成最终答案
    })
    
    g.AddEdge("plan_update", "plan_execution")  // 形成循环
    g.AddEdge("final_answer", graph.END)
    
    return g, nil
}
```

### 类型转换最小化

通过以上设计，我们实现了：

1. **减少转换节点**: 避免为了类型对齐而创建不必要的转换节点
2. **状态驱动**: 主要通过全局状态 `EnhancedState` 传递数据
3. **灵活类型**: 根据节点实际需求设定输入输出类型
4. **编排优化**: 遵循 Eino 框架的最佳实践

## 辅助函数

```go
// generatePlanID 生成规划ID
func generatePlanID() string {
    return fmt.Sprintf("plan_%d", time.Now().UnixNano())
}

// generateUpdateID 生成更新ID
func generateUpdateID() string {
    return fmt.Sprintf("update_%d", time.Now().UnixNano())
}

// PlanExecutionLambda 规划执行Lambda节点
func PlanExecutionLambda(ctx context.Context, state *EnhancedState, config *EnhancedMultiAgentConfig) (map[string]*SpecialistResult, error) {
    if state.CurrentPlan == nil {
        return nil, errors.New("no current plan available")
    }
    
    results := make(map[string]*SpecialistResult)
    
    // 并行执行所有待执行的步骤
    for _, step := range state.CurrentPlan.Steps {
        if step.Status == StepStatusPending {
            // 创建专家子图执行
            specialist, exists := config.Specialists[step.AssignedTo]
            if !exists {
                continue
            }
            
            // 执行专家任务
            result, err := executeSpecialistStep(ctx, step, specialist, state)
            if err != nil {
                continue
            }
            
            results[step.ID] = result
        }
    }
    
    return results, nil
}

// executeSpecialistStep 执行专家步骤
func executeSpecialistStep(ctx context.Context, step *PlanStep, specialist *EnhancedSpecialist, state *EnhancedState) (*SpecialistResult, error) {
    // 构建专家执行消息
    messages := []*schema.Message{
        {
            Role:    schema.RoleUser,
            Content: buildSpecialistPrompt(step, state.ConversationContext, state.OriginalMessages),
        },
    }
    
    // 调用专家模型
    response, err := specialist.ChatModel.Generate(ctx, messages)
    if err != nil {
        return nil, err
    }
    
    // 构建结果
    result := &SpecialistResult{
        StepID:         step.ID,
        SpecialistName: step.AssignedTo,
        Content:        response.Content,
        Success:        true,
        Confidence:     0.8,
        ExecutionTime:  time.Since(*step.StartTime),
        Timestamp:      time.Now(),
    }
    
    return result, nil
}
```

#### 2. 状态驱动
- **状态读取**: 节点通过 `StatePreHandler` 从 `EnhancedState` 读取所需数据，根据节点类型转换为合适的输入格式
- **状态更新**: 节点通过 `StatePostHandler` 更新 `EnhancedState`
- **数据转换**: 在处理器中完成状态数据与消息数据的转换
- **类型灵活性**: 避免为了类型对齐而创建不必要的转换节点

#### 3. 处理器模式
```go
// StatePreHandler 状态预处理器接口
type StatePreHandler interface {
    PreHandle(ctx context.Context, state *EnhancedState, input []*schema.Message) error
}

// StatePostHandler 状态后处理器接口
type StatePostHandler interface {
    PostHandle(ctx context.Context, state *EnhancedState, output *schema.Message) error
}

// 组合处理器
type StateHandler interface {
    StatePreHandler
    StatePostHandler
}
```

#### 4. 简化架构
- **Lambda节点**: 使用Lambda节点包装处理逻辑，根据节点实际需求设定类型
- **分支节点**: 使用条件分支控制流程
- **状态隔离**: 节点逻辑与状态管理分离
- **编排参考**: 遵循 [Eino Chain & Graph 编排规范](https://www.cloudwego.io/zh/docs/eino/core_modules/chain_and_graph_orchestration/)

## 核心节点实现

### 1. Host Think Node

```go
// HostThinkHandler Host思考节点处理器
type HostThinkHandler struct {
    config *EnhancedMultiAgentConfig
}

// PreHandle 准备思考输入
func (h *HostThinkHandler) PreHandle(ctx context.Context, state *EnhancedState, input []*schema.Message) error {
    // 构建思考提示，融合对话上下文
    thinkingPrompt := h.buildConversationalThinkingPrompt(state)
    
    // 更新输入消息
    if len(input) > 0 {
        input[0] = thinkingPrompt
    } else {
        input = append(input, thinkingPrompt)
    }
    
    return nil
}

// PostHandle 处理思考结果
func (h *HostThinkHandler) PostHandle(ctx context.Context, state *EnhancedState, output *schema.Message) error {
    // 解析思考结果
    thinkingResult, err := h.parseThinkingResult(output)
    if err != nil {
        return fmt.Errorf("failed to parse thinking result: %w", err)
    }
    
    // 更新状态
    state.CurrentThinkingResult = thinkingResult
    state.ThinkingHistory = append(state.ThinkingHistory, thinkingResult)
    state.LastUpdateTime = time.Now()
    
    return nil
}

// buildConversationalThinkingPrompt 构建对话感知的思考提示
func (h *HostThinkHandler) buildConversationalThinkingPrompt(state *EnhancedState) *schema.Message {
    var promptBuilder strings.Builder
    
    // 基础思考提示
    promptBuilder.WriteString("你是一个智能助手，需要分析用户的问题并制定解决方案。\n\n")
    
    // 对话上下文信息
    if state.ConversationContext != nil {
        ctx := state.ConversationContext
        promptBuilder.WriteString(fmt.Sprintf("对话上下文分析：\n"))
        promptBuilder.WriteString(fmt.Sprintf("- 对话轮次: %d/%d\n", ctx.CurrentTurn, ctx.TotalTurns))
        promptBuilder.WriteString(fmt.Sprintf("- 是否首次对话: %t\n", ctx.IsFirstTurn))
        promptBuilder.WriteString(fmt.Sprintf("- 是否延续对话: %t\n", ctx.IsContinuation))
        promptBuilder.WriteString(fmt.Sprintf("- 用户意图: %s (置信度: %.2f)\n", ctx.UserIntent, ctx.IntentConfidence))
        
        if ctx.ContextSummary != "" {
            promptBuilder.WriteString(fmt.Sprintf("- 上下文摘要: %s\n", ctx.ContextSummary))
        }
        
        if len(ctx.KeyTopics) > 0 {
            promptBuilder.WriteString(fmt.Sprintf("- 关键话题: %s\n", strings.Join(ctx.KeyTopics, ", ")))
        }
        
        promptBuilder.WriteString("\n")
    }
    
    // 对话历史
    if len(state.OriginalMessages) > 0 {
        promptBuilder.WriteString("对话历史：\n")
        for i, msg := range state.OriginalMessages {
            role := "用户"
            if msg.Role == schema.RoleAssistant {
                role = "助手"
            }
            promptBuilder.WriteString(fmt.Sprintf("%d. %s: %s\n", i+1, role, msg.Content))
        }
        promptBuilder.WriteString("\n")
    }
    
    // 思考指导
    promptBuilder.WriteString("请按照以下步骤进行思考：\n")
    promptBuilder.WriteString("1. 理解问题：分析用户的具体需求和意图\n")
    promptBuilder.WriteString("2. 评估复杂度：判断问题的复杂程度（low/medium/high/very_high）\n")
    promptBuilder.WriteString("3. 制定策略：确定解决问题的最佳方法\n")
    promptBuilder.WriteString("4. 考虑上下文：结合对话历史提供连贯的回应\n\n")
    
    promptBuilder.WriteString("请以JSON格式输出你的思考结果：\n")
    promptBuilder.WriteString("```json\n")
    promptBuilder.WriteString("{\n")
    promptBuilder.WriteString("  \"understanding\": \"对问题的理解\",\n")
    promptBuilder.WriteString("  \"complexity\": \"low|medium|high|very_high\",\n")
    promptBuilder.WriteString("  \"strategy\": \"解决策略\",\n")
    promptBuilder.WriteString("  \"reasoning\": \"推理过程\",\n")
    promptBuilder.WriteString("  \"context_consideration\": \"上下文考虑\"\n")
    promptBuilder.WriteString("}\n")
    promptBuilder.WriteString("```")
    
    return &schema.Message{
        Role:    schema.RoleUser,
        Content: promptBuilder.String(),
    }
}

// parseThinkingResult 解析思考结果
func (h *HostThinkHandler) parseThinkingResult(output *schema.Message) (*ThinkingResult, error) {
    // 提取JSON内容
    content := output.Content
    start := strings.Index(content, "```json")
    end := strings.LastIndex(content, "```")
    
    if start == -1 || end == -1 || start >= end {
        return nil, errors.New("no valid JSON found in thinking result")
    }
    
    jsonStr := content[start+7 : end]
    
    // 解析JSON
    var result struct {
        Understanding        string `json:"understanding"`
        Complexity          string `json:"complexity"`
        Strategy            string `json:"strategy"`
        Reasoning           string `json:"reasoning"`
        ContextConsideration string `json:"context_consideration"`
    }
    
    if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
        return nil, fmt.Errorf("failed to parse thinking JSON: %w", err)
    }
    
    // 转换复杂度
    complexity := parseComplexity(result.Complexity)
    
    return &ThinkingResult{
        Understanding:        result.Understanding,
        Complexity:          complexity,
        Strategy:            result.Strategy,
        Reasoning:           result.Reasoning,
        ContextConsideration: result.ContextConsideration,
        Timestamp:           time.Now(),
    }, nil
}

// parseComplexity 解析复杂度字符串
func parseComplexity(complexityStr string) TaskComplexity {
    switch strings.ToLower(complexityStr) {
    case "low":
        return TaskComplexityLow
    case "medium":
        return TaskComplexityMedium
    case "high":
        return TaskComplexityHigh
    case "very_high":
        return TaskComplexityVeryHigh
    default:
        return TaskComplexityUnknown
    }
}
```

### 2. Direct Answer Node

```go
// DirectAnswerHandler 直接回答节点处理器
type DirectAnswerHandler struct {
    config *EnhancedMultiAgentConfig
}

// PreHandle 准备直接回答输入
func (h *DirectAnswerHandler) PreHandle(ctx context.Context, state *EnhancedState, input []*schema.Message) error {
    // 构建直接回答提示
    answerPrompt := h.buildDirectAnswerPrompt(state)
    
    // 更新输入消息
    if len(input) > 0 {
        input[0] = answerPrompt
    } else {
        input = append(input, answerPrompt)
    }
    
    return nil
}

// PostHandle 处理直接回答结果
func (h *DirectAnswerHandler) PostHandle(ctx context.Context, state *EnhancedState, output *schema.Message) error {
    // 标记为简单任务已完成
    state.IsSimpleTask = true
    state.IsCompleted = true
    state.FinalAnswer = output
    state.LastUpdateTime = time.Now()
    
    return nil
}

// buildDirectAnswerPrompt 构建直接回答提示
func (h *DirectAnswerHandler) buildDirectAnswerPrompt(state *EnhancedState) *schema.Message {
    var promptBuilder strings.Builder
    
    promptBuilder.WriteString("你是一个智能助手，请直接回答用户的问题。\n\n")
    
    // 添加对话上下文
    if state.ConversationContext != nil && state.ConversationContext.IsContinuation {
        promptBuilder.WriteString("注意：这是一个延续对话，请考虑之前的对话内容。\n\n")
    }
    
    // 添加思考结果作为参考
    if state.CurrentThinkingResult != nil {
        promptBuilder.WriteString("参考分析：\n")
        promptBuilder.WriteString(fmt.Sprintf("- 问题理解: %s\n", state.CurrentThinkingResult.Understanding))
        promptBuilder.WriteString(fmt.Sprintf("- 解决策略: %s\n", state.CurrentThinkingResult.Strategy))
        if state.CurrentThinkingResult.ContextConsideration != "" {
            promptBuilder.WriteString(fmt.Sprintf("- 上下文考虑: %s\n", state.CurrentThinkingResult.ContextConsideration))
        }
        promptBuilder.WriteString("\n")
    }
    
    // 添加对话历史
    if len(state.OriginalMessages) > 0 {
        promptBuilder.WriteString("对话历史：\n")
        for i, msg := range state.OriginalMessages {
            role := "用户"
            if msg.Role == schema.RoleAssistant {
                role = "助手"
            }
            promptBuilder.WriteString(fmt.Sprintf("%d. %s: %s\n", i+1, role, msg.Content))
        }
        promptBuilder.WriteString("\n")
    }
    
    promptBuilder.WriteString("请提供清晰、准确、有帮助的回答。")
    
    return &schema.Message{
        Role:    schema.RoleUser,
        Content: promptBuilder.String(),
    }
}
```

### 3. Complexity Branch

```go
// ComplexityBranch 复杂度判断分支
type ComplexityBranch struct{}

// Decide 决定分支路径
func (b *ComplexityBranch) Decide(ctx context.Context, state *EnhancedState) (string, error) {
    if state.CurrentThinkingResult == nil {
        return "", errors.New("thinking result not available for complexity decision")
    }
    
    complexity := state.CurrentThinkingResult.Complexity
    
    // 考虑对话上下文
    if state.ConversationContext != nil && state.ConversationContext.IsContinuation {
        // 延续对话可能需要更复杂的处理
        if complexity >= TaskComplexityMedium {
            return "complex", nil
        }
    }
    
    // 基于复杂度判断
    if complexity >= TaskComplexityHigh {
        return "complex", nil
    }
    
    return "simple", nil
}
```

### 4. Plan Creation Node

```go
// PlanCreationHandler 规划创建节点处理器
type PlanCreationHandler struct {
    config *EnhancedMultiAgentConfig
}

// PreHandle 准备规划创建输入
func (h *PlanCreationHandler) PreHandle(ctx context.Context, state *EnhancedState, input []*schema.Message) error {
    planPrompt := h.buildPlanCreationPrompt(state)
    
    if len(input) > 0 {
        input[0] = planPrompt
    } else {
        input = append(input, planPrompt)
    }
    
    return nil
}

// PostHandle 处理规划创建结果
func (h *PlanCreationHandler) PostHandle(ctx context.Context, state *EnhancedState, output *schema.Message) error {
    // 解析规划结果
    plan, err := h.parsePlanFromMessage(output)
    if err != nil {
        return fmt.Errorf("failed to parse plan: %w", err)
    }
    
    // 设置规划属性
    plan.ID = generatePlanID()
    plan.Version = 1
    plan.CreatedAt = time.Now()
    plan.UpdatedAt = time.Now()
    plan.Status = PlanStatusActive
    
    // 基于对话上下文调整规划
    if state.ConversationContext != nil && state.ConversationContext.IsContinuation {
        h.adjustPlanForContinuation(plan, state.ConversationContext)
    }
    
    // 更新状态
    state.CurrentPlan = plan
    state.IsSimpleTask = false
    state.LastUpdateTime = time.Now()
    
    return nil
}

// buildPlanCreationPrompt 构建规划创建提示
func (h *PlanCreationHandler) buildPlanCreationPrompt(state *EnhancedState) *schema.Message {
    var promptBuilder strings.Builder
    
    promptBuilder.WriteString("你是一个任务规划专家，需要为复杂任务制定详细的执行计划。\n\n")
    
    // 添加思考结果
    if state.CurrentThinkingResult != nil {
        promptBuilder.WriteString("基于以下分析制定计划：\n")
        promptBuilder.WriteString(fmt.Sprintf("- 问题理解: %s\n", state.CurrentThinkingResult.Understanding))
        promptBuilder.WriteString(fmt.Sprintf("- 复杂度: %s\n", state.CurrentThinkingResult.Complexity.String()))
        promptBuilder.WriteString(fmt.Sprintf("- 策略: %s\n", state.CurrentThinkingResult.Strategy))
        promptBuilder.WriteString(fmt.Sprintf("- 推理: %s\n", state.CurrentThinkingResult.Reasoning))
        promptBuilder.WriteString("\n")
    }
    
    // 添加可用专家信息
    if len(h.config.Specialists) > 0 {
        promptBuilder.WriteString("可用专家：\n")
        for name, specialist := range h.config.Specialists {
            promptBuilder.WriteString(fmt.Sprintf("- %s: %s\n", name, specialist.SystemPrompt))
        }
        promptBuilder.WriteString("\n")
    }
    
    // 添加对话上下文
    if state.ConversationContext != nil && state.ConversationContext.IsContinuation {
        promptBuilder.WriteString("注意：这是延续对话，请考虑上下文连贯性。\n\n")
    }
    
    promptBuilder.WriteString("请制定详细的执行计划，以JSON格式输出：\n")
    promptBuilder.WriteString("```json\n")
    promptBuilder.WriteString("{\n")
    promptBuilder.WriteString("  \"name\": \"计划名称\",\n")
    promptBuilder.WriteString("  \"description\": \"计划描述\",\n")
    promptBuilder.WriteString("  \"steps\": [\n")
    promptBuilder.WriteString("    {\n")
    promptBuilder.WriteString("      \"id\": \"step_1\",\n")
    promptBuilder.WriteString("      \"name\": \"步骤名称\",\n")
    promptBuilder.WriteString("      \"description\": \"步骤描述\",\n")
    promptBuilder.WriteString("      \"assigned_to\": \"专家名称\",\n")
    promptBuilder.WriteString("      \"priority\": 1,\n")
    promptBuilder.WriteString("      \"dependencies\": [],\n")
    promptBuilder.WriteString("      \"parameters\": {}\n")
    promptBuilder.WriteString("    }\n")
    promptBuilder.WriteString("  ]\n")
    promptBuilder.WriteString("}\n")
    promptBuilder.WriteString("```")
    
    return &schema.Message{
        Role:    schema.RoleUser,
        Content: promptBuilder.String(),
    }
}

// parsePlanFromMessage 从消息中解析规划
func (h *PlanCreationHandler) parsePlanFromMessage(output *schema.Message) (*TaskPlan, error) {
    content := output.Content
    start := strings.Index(content, "```json")
    end := strings.LastIndex(content, "```")
    
    if start == -1 || end == -1 || start >= end {
        return nil, errors.New("no valid JSON found in plan result")
    }
    
    jsonStr := content[start+7 : end]
    
    var planData struct {
        Name        string `json:"name"`
        Description string `json:"description"`
        Steps       []struct {
            ID           string                 `json:"id"`
            Name         string                 `json:"name"`
            Description  string                 `json:"description"`
            AssignedTo   string                 `json:"assigned_to"`
            Priority     int                    `json:"priority"`
            Dependencies []string               `json:"dependencies"`
            Parameters   map[string]interface{} `json:"parameters"`
        } `json:"steps"`
    }
    
    if err := json.Unmarshal([]byte(jsonStr), &planData); err != nil {
        return nil, fmt.Errorf("failed to parse plan JSON: %w", err)
    }
    
    // 构建TaskPlan
    plan := &TaskPlan{
        Name:        planData.Name,
        Description: planData.Description,
        Steps:       make([]*PlanStep, len(planData.Steps)),
        TotalSteps:  len(planData.Steps),
        CurrentStep: 0,
        Dependencies: make(map[string][]string),
        ResourceAllocation: make(map[string]string),
        Metadata:    make(map[string]interface{}),
    }
    
    // 构建步骤
    for i, stepData := range planData.Steps {
        step := &PlanStep{
            ID:           stepData.ID,
            Name:         stepData.Name,
            Description:  stepData.Description,
            AssignedTo:   stepData.AssignedTo,
            Priority:     stepData.Priority,
            Status:       StepStatusPending,
            Dependencies: stepData.Dependencies,
            Parameters:   stepData.Parameters,
            Metadata:     make(map[string]interface{}),
        }
        
        plan.Steps[i] = step
        plan.Dependencies[step.ID] = step.Dependencies
        plan.ResourceAllocation[step.ID] = step.AssignedTo
    }
    
    return plan, nil
}

// adjustPlanForContinuation 为延续对话调整规划
func (h *PlanCreationHandler) adjustPlanForContinuation(plan *TaskPlan, ctx *ConversationContext) {
    // 添加上下文相关的元数据
    plan.Metadata["conversation_context"] = ctx
    plan.Metadata["is_continuation"] = true
    
    // 可以根据上下文调整步骤优先级或添加额外步骤
    if ctx.RequiresClarification {
        // 添加澄清步骤
        clarificationStep := &PlanStep{
            ID:          "clarification",
            Name:        "澄清用户意图",
            Description: "基于对话历史澄清用户的具体需求",
            AssignedTo:  "host",
            Priority:    10, // 最高优先级
            Status:      StepStatusPending,
            Parameters:  map[string]interface{}{"context": ctx},
            Metadata:    map[string]interface{}{"auto_generated": true},
        }
        
        // 插入到计划开头
        plan.Steps = append([]*PlanStep{clarificationStep}, plan.Steps...)
        plan.TotalSteps++
    }
}
```

### 5. Plan Execution Node

```go
// PlanExecutionHandler 规划执行节点处理器
type PlanExecutionHandler struct {
    config *EnhancedMultiAgentConfig
}

// PreHandle 准备规划执行
func (h *PlanExecutionHandler) PreHandle(ctx context.Context, state *EnhancedState, input []*schema.Message) error {
    if state.CurrentPlan == nil {
        return errors.New("no current plan available for execution")
    }
    
    // 找到下一个待执行的步骤
    nextStep, err := h.findNextExecutableStep(state.CurrentPlan)
    if err != nil {
        return fmt.Errorf("failed to find next executable step: %w", err)
    }
    
    // 创建执行上下文
    executionCtx := &ExecutionContext{
        PlanID:      state.CurrentPlan.ID,
        PlanVersion: state.CurrentPlan.Version,
        PlanStep:    nextStep,
        Round:       state.CurrentRound,
        StartTime:   time.Now(),
    }
    
    // 更新状态
    state.CurrentExecution = executionCtx
    nextStep.Status = StepStatusExecuting
    nextStep.StartTime = &executionCtx.StartTime
    state.CurrentPlan.CurrentStep = h.findStepIndex(state.CurrentPlan, nextStep.ID)
    
    return nil
}

// PostHandle 处理执行结果
func (h *PlanExecutionHandler) PostHandle(ctx context.Context, state *EnhancedState, output *schema.Message) error {
    // 执行准备完成，等待专家执行
    state.LastUpdateTime = time.Now()
    return nil
}

// findNextExecutableStep 找到下一个可执行的步骤
func (h *PlanExecutionHandler) findNextExecutableStep(plan *TaskPlan) (*PlanStep, error) {
    for _, step := range plan.Steps {
        if step.Status == StepStatusPending && h.areDependenciesSatisfied(plan, step) {
            return step, nil
        }
    }
    
    return nil, errors.New("no executable step found")
}

// areDependenciesSatisfied 检查依赖是否满足
func (h *PlanExecutionHandler) areDependenciesSatisfied(plan *TaskPlan, step *PlanStep) bool {
    for _, depID := range step.Dependencies {
        depStep := h.findStepByID(plan, depID)
        if depStep == nil || depStep.Status != StepStatusCompleted {
            return false
        }
    }
    return true
}

// findStepByID 根据ID查找步骤
func (h *PlanExecutionHandler) findStepByID(plan *TaskPlan, stepID string) *PlanStep {
    for _, step := range plan.Steps {
        if step.ID == stepID {
            return step
        }
    }
    return nil
}

// findStepIndex 查找步骤索引
func (h *PlanExecutionHandler) findStepIndex(plan *TaskPlan, stepID string) int {
    for i, step := range plan.Steps {
        if step.ID == stepID {
            return i
        }
    }
    return -1
}
```

### 6. Specialist Multi-Branch

```go
// SpecialistMultiBranch 专家多分支节点
type SpecialistMultiBranch struct {
    config *EnhancedMultiAgentConfig
}

// GetBranches 获取分支列表
func (b *SpecialistMultiBranch) GetBranches(ctx context.Context, state *EnhancedState) ([]string, error) {
    if state.CurrentExecution == nil || state.CurrentExecution.PlanStep == nil {
        return nil, errors.New("no current execution step available")
    }
    
    assignedTo := state.CurrentExecution.PlanStep.AssignedTo
    
    // 检查专家是否存在
    if _, exists := b.config.Specialists[assignedTo]; !exists {
        return nil, fmt.Errorf("specialist %s not found", assignedTo)
    }
    
    return []string{assignedTo}, nil
}
```