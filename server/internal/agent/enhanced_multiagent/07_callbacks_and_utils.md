# Eino Enhanced MultiAgent 回调接口与辅助函数

## 增强版回调接口

### EnhancedMultiAgentCallback - 主回调接口

```go
// EnhancedMultiAgentCallback 增强版多智能体系统回调接口
type EnhancedMultiAgentCallback interface {
    // 系统生命周期回调
    OnSystemStart(ctx context.Context, config *EnhancedMultiAgentConfig) error
    OnSystemEnd(ctx context.Context, state *EnhancedState, result *schema.Message) error
    
    // 对话上下文回调
    OnConversationAnalysis(ctx context.Context, messages []*schema.Message, context *ConversationContext) error
    
    // 思考阶段回调
    OnThinkingStart(ctx context.Context, state *EnhancedState) error
    OnThinkingEnd(ctx context.Context, state *EnhancedState, result *ThinkingResult) error
    
    // 复杂度判断回调
    OnComplexityDecision(ctx context.Context, state *EnhancedState, complexity TaskComplexity, isComplex bool) error
    
    // 规划阶段回调
    OnPlanCreationStart(ctx context.Context, state *EnhancedState) error
    OnPlanCreationEnd(ctx context.Context, state *EnhancedState, plan *TaskPlan) error
    OnPlanUpdate(ctx context.Context, state *EnhancedState, oldPlan, newPlan *TaskPlan, update *PlanUpdate) error
    
    // 执行阶段回调
    OnExecutionStart(ctx context.Context, state *EnhancedState, step *PlanStep) error
    OnExecutionEnd(ctx context.Context, state *EnhancedState, step *PlanStep, result *SpecialistResult) error
    
    // 专家执行回调
    OnSpecialistStart(ctx context.Context, state *EnhancedState, specialistName string, step *PlanStep) error
    OnSpecialistEnd(ctx context.Context, state *EnhancedState, specialistName string, result *SpecialistResult) error
    OnSpecialistError(ctx context.Context, state *EnhancedState, specialistName string, err error) error
    
    // 结果收集回调
    OnResultCollection(ctx context.Context, state *EnhancedState, results map[string]*SpecialistResult, collected *CollectedResults) error
    
    // 反馈处理回调
    OnFeedbackStart(ctx context.Context, state *EnhancedState, collected *CollectedResults) error
    OnFeedbackEnd(ctx context.Context, state *EnhancedState, feedback *FeedbackResult) error
    
    // 反思决策回调
    OnReflectionDecision(ctx context.Context, state *EnhancedState, shouldContinue bool, reason string) error
    
    // 轮次控制回调
    OnRoundStart(ctx context.Context, state *EnhancedState, round int) error
    OnRoundEnd(ctx context.Context, state *EnhancedState, round int) error
    OnMaxRoundsReached(ctx context.Context, state *EnhancedState) error
    
    // 错误处理回调
    OnError(ctx context.Context, state *EnhancedState, err error, stage string) error
    OnRecovery(ctx context.Context, state *EnhancedState, err error, recovery string) error
    
    // 性能监控回调
    OnPerformanceMetric(ctx context.Context, metric *PerformanceMetric) error
    OnResourceUsage(ctx context.Context, usage *ResourceUsage) error
}
```

### 默认回调实现

```go
// DefaultEnhancedCallback 默认回调实现
type DefaultEnhancedCallback struct {
    logger *zap.Logger
    config *EnhancedMultiAgentConfig
}

// NewDefaultEnhancedCallback 创建默认回调实例
func NewDefaultEnhancedCallback(logger *zap.Logger, config *EnhancedMultiAgentConfig) *DefaultEnhancedCallback {
    return &DefaultEnhancedCallback{
        logger: logger,
        config: config,
    }
}

// OnSystemStart 系统启动回调
func (c *DefaultEnhancedCallback) OnSystemStart(ctx context.Context, config *EnhancedMultiAgentConfig) error {
    c.logger.Info("Enhanced MultiAgent System started",
        zap.String("system_name", config.SystemName),
        zap.String("system_version", config.SystemVersion),
        zap.Int("max_rounds", config.MaxRounds),
        zap.String("complexity_threshold", config.ComplexityThreshold.String()),
    )
    return nil
}

// OnSystemEnd 系统结束回调
func (c *DefaultEnhancedCallback) OnSystemEnd(ctx context.Context, state *EnhancedState, result *schema.Message) error {
    duration := time.Since(state.StartTime)
    c.logger.Info("Enhanced MultiAgent System completed",
        zap.String("session_id", state.SessionID),
        zap.Duration("total_duration", duration),
        zap.Int("total_rounds", state.CurrentRound),
        zap.Bool("is_completed", state.IsCompleted),
        zap.Bool("is_simple_task", state.IsSimpleTask),
    )
    return nil
}

// OnConversationAnalysis 对话分析回调
func (c *DefaultEnhancedCallback) OnConversationAnalysis(ctx context.Context, messages []*schema.Message, context *ConversationContext) error {
    c.logger.Info("Conversation context analyzed",
        zap.Int("total_turns", context.TotalTurns),
        zap.Int("current_turn", context.CurrentTurn),
        zap.Bool("is_first_turn", context.IsFirstTurn),
        zap.Bool("is_continuation", context.IsContinuation),
        zap.String("user_intent", context.UserIntent),
        zap.Float64("intent_confidence", context.IntentConfidence),
    )
    return nil
}

// OnThinkingStart 思考开始回调
func (c *DefaultEnhancedCallback) OnThinkingStart(ctx context.Context, state *EnhancedState) error {
    c.logger.Info("Host thinking started",
        zap.String("session_id", state.SessionID),
        zap.Int("thinking_round", len(state.ThinkingHistory)+1),
    )
    return nil
}

// OnThinkingEnd 思考结束回调
func (c *DefaultEnhancedCallback) OnThinkingEnd(ctx context.Context, state *EnhancedState, result *ThinkingResult) error {
    c.logger.Info("Host thinking completed",
        zap.String("session_id", state.SessionID),
        zap.String("understanding", result.Understanding),
        zap.String("complexity", result.Complexity.String()),
        zap.String("strategy", result.Strategy),
    )
    return nil
}

// OnComplexityDecision 复杂度决策回调
func (c *DefaultEnhancedCallback) OnComplexityDecision(ctx context.Context, state *EnhancedState, complexity TaskComplexity, isComplex bool) error {
    c.logger.Info("Complexity decision made",
        zap.String("session_id", state.SessionID),
        zap.String("complexity", complexity.String()),
        zap.Bool("is_complex", isComplex),
        zap.String("next_action", func() string {
            if isComplex {
                return "create_plan"
            }
            return "direct_answer"
        }()),
    )
    return nil
}

// OnPlanCreationStart 规划创建开始回调
func (c *DefaultEnhancedCallback) OnPlanCreationStart(ctx context.Context, state *EnhancedState) error {
    c.logger.Info("Plan creation started",
        zap.String("session_id", state.SessionID),
    )
    return nil
}

// OnPlanCreationEnd 规划创建结束回调
func (c *DefaultEnhancedCallback) OnPlanCreationEnd(ctx context.Context, state *EnhancedState, plan *TaskPlan) error {
    c.logger.Info("Plan creation completed",
        zap.String("session_id", state.SessionID),
        zap.String("plan_id", plan.ID),
        zap.String("plan_name", plan.Name),
        zap.Int("total_steps", plan.TotalSteps),
        zap.Int("version", plan.Version),
    )
    return nil
}

// OnPlanUpdate 规划更新回调
func (c *DefaultEnhancedCallback) OnPlanUpdate(ctx context.Context, state *EnhancedState, oldPlan, newPlan *TaskPlan, update *PlanUpdate) error {
    c.logger.Info("Plan updated",
        zap.String("session_id", state.SessionID),
        zap.String("plan_id", newPlan.ID),
        zap.Int("old_version", oldPlan.Version),
        zap.Int("new_version", newPlan.Version),
        zap.String("update_type", update.UpdateType.String()),
        zap.String("update_description", update.Description),
    )
    return nil
}

// OnExecutionStart 执行开始回调
func (c *DefaultEnhancedCallback) OnExecutionStart(ctx context.Context, state *EnhancedState, step *PlanStep) error {
    c.logger.Info("Step execution started",
        zap.String("session_id", state.SessionID),
        zap.String("step_id", step.ID),
        zap.String("step_name", step.Name),
        zap.String("assigned_to", step.AssignedTo),
        zap.Int("priority", step.Priority),
    )
    return nil
}

// OnExecutionEnd 执行结束回调
func (c *DefaultEnhancedCallback) OnExecutionEnd(ctx context.Context, state *EnhancedState, step *PlanStep, result *SpecialistResult) error {
    duration := time.Duration(0)
    if step.StartTime != nil && step.EndTime != nil {
        duration = step.EndTime.Sub(*step.StartTime)
    }
    
    c.logger.Info("Step execution completed",
        zap.String("session_id", state.SessionID),
        zap.String("step_id", step.ID),
        zap.String("step_name", step.Name),
        zap.String("assigned_to", step.AssignedTo),
        zap.String("status", step.Status.String()),
        zap.Duration("execution_duration", duration),
        zap.Bool("success", result.Success),
        zap.Float64("confidence", result.Confidence),
    )
    return nil
}

// OnSpecialistStart 专家开始回调
func (c *DefaultEnhancedCallback) OnSpecialistStart(ctx context.Context, state *EnhancedState, specialistName string, step *PlanStep) error {
    c.logger.Info("Specialist execution started",
        zap.String("session_id", state.SessionID),
        zap.String("specialist_name", specialistName),
        zap.String("step_id", step.ID),
        zap.String("step_name", step.Name),
    )
    return nil
}

// OnSpecialistEnd 专家结束回调
func (c *DefaultEnhancedCallback) OnSpecialistEnd(ctx context.Context, state *EnhancedState, specialistName string, result *SpecialistResult) error {
    c.logger.Info("Specialist execution completed",
        zap.String("session_id", state.SessionID),
        zap.String("specialist_name", specialistName),
        zap.Bool("success", result.Success),
        zap.Float64("confidence", result.Confidence),
        zap.Duration("execution_time", result.ExecutionTime),
    )
    return nil
}

// OnSpecialistError 专家错误回调
func (c *DefaultEnhancedCallback) OnSpecialistError(ctx context.Context, state *EnhancedState, specialistName string, err error) error {
    c.logger.Error("Specialist execution error",
        zap.String("session_id", state.SessionID),
        zap.String("specialist_name", specialistName),
        zap.Error(err),
    )
    return nil
}

// OnResultCollection 结果收集回调
func (c *DefaultEnhancedCallback) OnResultCollection(ctx context.Context, state *EnhancedState, results map[string]*SpecialistResult, collected *CollectedResults) error {
    c.logger.Info("Results collected",
        zap.String("session_id", state.SessionID),
        zap.Int("specialist_count", len(results)),
        zap.Float64("overall_confidence", collected.OverallConfidence),
        zap.Float64("quality_score", collected.QualityScore),
    )
    return nil
}

// OnFeedbackStart 反馈开始回调
func (c *DefaultEnhancedCallback) OnFeedbackStart(ctx context.Context, state *EnhancedState, collected *CollectedResults) error {
    c.logger.Info("Feedback processing started",
        zap.String("session_id", state.SessionID),
        zap.Float64("overall_confidence", collected.OverallConfidence),
    )
    return nil
}

// OnFeedbackEnd 反馈结束回调
func (c *DefaultEnhancedCallback) OnFeedbackEnd(ctx context.Context, state *EnhancedState, feedback *FeedbackResult) error {
    c.logger.Info("Feedback processing completed",
        zap.String("session_id", state.SessionID),
        zap.Bool("should_continue", feedback.ShouldContinue),
        zap.String("suggested_action", feedback.SuggestedAction.String()),
        zap.String("feedback_summary", feedback.FeedbackSummary),
    )
    return nil
}

// OnReflectionDecision 反思决策回调
func (c *DefaultEnhancedCallback) OnReflectionDecision(ctx context.Context, state *EnhancedState, shouldContinue bool, reason string) error {
    c.logger.Info("Reflection decision made",
        zap.String("session_id", state.SessionID),
        zap.Bool("should_continue", shouldContinue),
        zap.String("reason", reason),
        zap.Int("current_round", state.CurrentRound),
        zap.Int("max_rounds", state.MaxRounds),
    )
    return nil
}

// OnRoundStart 轮次开始回调
func (c *DefaultEnhancedCallback) OnRoundStart(ctx context.Context, state *EnhancedState, round int) error {
    c.logger.Info("Execution round started",
        zap.String("session_id", state.SessionID),
        zap.Int("round", round),
        zap.Int("max_rounds", state.MaxRounds),
    )
    return nil
}

// OnRoundEnd 轮次结束回调
func (c *DefaultEnhancedCallback) OnRoundEnd(ctx context.Context, state *EnhancedState, round int) error {
    c.logger.Info("Execution round completed",
        zap.String("session_id", state.SessionID),
        zap.Int("round", round),
        zap.Int("completed_steps", state.CurrentPlan.CompletedSteps),
        zap.Int("total_steps", state.CurrentPlan.TotalSteps),
    )
    return nil
}

// OnMaxRoundsReached 达到最大轮次回调
func (c *DefaultEnhancedCallback) OnMaxRoundsReached(ctx context.Context, state *EnhancedState) error {
    c.logger.Warn("Maximum rounds reached",
        zap.String("session_id", state.SessionID),
        zap.Int("max_rounds", state.MaxRounds),
        zap.Bool("is_completed", state.IsCompleted),
    )
    return nil
}

// OnError 错误处理回调
func (c *DefaultEnhancedCallback) OnError(ctx context.Context, state *EnhancedState, err error, stage string) error {
    c.logger.Error("System error occurred",
        zap.String("session_id", state.SessionID),
        zap.String("stage", stage),
        zap.Error(err),
    )
    return nil
}

// OnRecovery 恢复处理回调
func (c *DefaultEnhancedCallback) OnRecovery(ctx context.Context, state *EnhancedState, err error, recovery string) error {
    c.logger.Info("System recovery attempted",
        zap.String("session_id", state.SessionID),
        zap.Error(err),
        zap.String("recovery_action", recovery),
    )
    return nil
}

// OnPerformanceMetric 性能指标回调
func (c *DefaultEnhancedCallback) OnPerformanceMetric(ctx context.Context, metric *PerformanceMetric) error {
    if c.config.Logging.EnablePerformanceLog {
        c.logger.Info("Performance metric",
            zap.String("metric_name", metric.Name),
            zap.Float64("value", metric.Value),
            zap.String("unit", metric.Unit),
            zap.Time("timestamp", metric.Timestamp),
        )
    }
    return nil
}

// OnResourceUsage 资源使用回调
func (c *DefaultEnhancedCallback) OnResourceUsage(ctx context.Context, usage *ResourceUsage) error {
    if c.config.Performance.EnableMetrics {
        c.logger.Info("Resource usage",
            zap.Int64("memory_usage", usage.MemoryUsage),
            zap.Float64("cpu_usage", usage.CPUUsage),
            zap.Int("goroutine_count", usage.GoroutineCount),
            zap.Time("timestamp", usage.Timestamp),
        )
    }
    return nil
}
```

## 辅助函数

### 序列化/反序列化函数

```go
// SerializeState 序列化状态
func SerializeState(state *EnhancedState) ([]byte, error) {
    return json.Marshal(state)
}

// DeserializeState 反序列化状态
func DeserializeState(data []byte) (*EnhancedState, error) {
    var state EnhancedState
    if err := json.Unmarshal(data, &state); err != nil {
        return nil, fmt.Errorf("failed to deserialize state: %w", err)
    }
    
    // 验证状态完整性
    if err := state.validateState(); err != nil {
        return nil, fmt.Errorf("invalid state: %w", err)
    }
    
    return &state, nil
}

// SerializeConfig 序列化配置
func SerializeConfig(config *EnhancedMultiAgentConfig) ([]byte, error) {
    return yaml.Marshal(config)
}

// DeserializeConfig 反序列化配置
func DeserializeConfig(data []byte) (*EnhancedMultiAgentConfig, error) {
    var config EnhancedMultiAgentConfig
    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, fmt.Errorf("failed to deserialize config: %w", err)
    }
    
    // 验证配置完整性
    if err := validateConfig(&config); err != nil {
        return nil, fmt.Errorf("invalid config: %w", err)
    }
    
    return &config, nil
}

// CompressState 压缩状态（用于长期存储）
func CompressState(state *EnhancedState) ([]byte, error) {
    compressed, err := state.Compress()
    if err != nil {
        return nil, fmt.Errorf("failed to compress state: %w", err)
    }
    
    data, err := json.Marshal(compressed)
    if err != nil {
        return nil, fmt.Errorf("failed to serialize compressed state: %w", err)
    }
    
    return data, nil
}
```

### 提示构建函数

```go
// BuildSystemPrompt 构建系统提示
func BuildSystemPrompt(template string, params map[string]interface{}) string {
    result := template
    for key, value := range params {
        placeholder := fmt.Sprintf("{%s}", key)
        replacement := fmt.Sprintf("%v", value)
        result = strings.ReplaceAll(result, placeholder, replacement)
    }
    return result
}

// BuildConversationalPrompt 构建对话感知提示
func BuildConversationalPrompt(basePrompt string, context *ConversationContext, messages []*schema.Message) string {
    var builder strings.Builder
    
    // 添加基础提示
    builder.WriteString(basePrompt)
    builder.WriteString("\n\n")
    
    // 添加对话上下文
    if context != nil {
        builder.WriteString("对话上下文：\n")
        builder.WriteString(fmt.Sprintf("- 对话轮次: %d/%d\n", context.CurrentTurn, context.TotalTurns))
        builder.WriteString(fmt.Sprintf("- 是否延续对话: %t\n", context.IsContinuation))
        builder.WriteString(fmt.Sprintf("- 用户意图: %s\n", context.UserIntent))
        
        if context.ContextSummary != "" {
            builder.WriteString(fmt.Sprintf("- 上下文摘要: %s\n", context.ContextSummary))
        }
        
        builder.WriteString("\n")
    }
    
    // 添加相关历史
    if len(messages) > 0 {
        builder.WriteString("相关对话历史：\n")
        for i, msg := range messages {
            role := "用户"
            if msg.Role == schema.RoleAssistant {
                role = "助手"
            }
            builder.WriteString(fmt.Sprintf("%d. %s: %s\n", i+1, role, msg.Content))
        }
        builder.WriteString("\n")
    }
    
    return builder.String()
}

// BuildSpecialistPrompt 构建专家提示
func BuildSpecialistPrompt(step *PlanStep, context *ConversationContext, messages []*schema.Message) string {
    var builder strings.Builder
    
    builder.WriteString(fmt.Sprintf("执行任务步骤：%s\n", step.Name))
    builder.WriteString(fmt.Sprintf("步骤描述：%s\n\n", step.Description))
    
    // 添加执行参数
    if len(step.Parameters) > 0 {
        builder.WriteString("执行参数：\n")
        for key, value := range step.Parameters {
            builder.WriteString(fmt.Sprintf("- %s: %v\n", key, value))
        }
        builder.WriteString("\n")
    }
    
    // 添加对话上下文
    if context != nil && context.IsContinuation {
        builder.WriteString("注意：这是延续对话的一部分，请考虑上下文连贯性。\n\n")
    }
    
    // 添加相关历史
    if len(messages) > 0 {
        builder.WriteString("相关对话历史：\n")
        for i, msg := range messages {
            role := "用户"
            if msg.Role == schema.RoleAssistant {
                role = "助手"
            }
            builder.WriteString(fmt.Sprintf("%d. %s: %s\n", i+1, role, msg.Content))
        }
        builder.WriteString("\n")
    }
    
    builder.WriteString("请完成此步骤并提供详细的执行结果。")
    
    return builder.String()
}
```

### 解析函数

```go
// ParseJSONFromMessage 从消息中解析JSON
func ParseJSONFromMessage(message *schema.Message, target interface{}) error {
    content := message.Content
    
    // 查找JSON代码块
    start := strings.Index(content, "```json")
    end := strings.LastIndex(content, "```")
    
    if start == -1 || end == -1 || start >= end {
        // 尝试直接解析整个内容
        if err := json.Unmarshal([]byte(content), target); err != nil {
            return fmt.Errorf("no valid JSON found in message: %w", err)
        }
        return nil
    }
    
    jsonStr := strings.TrimSpace(content[start+7 : end])
    
    if err := json.Unmarshal([]byte(jsonStr), target); err != nil {
        return fmt.Errorf("failed to parse JSON: %w", err)
    }
    
    return nil
}

// ExtractCodeBlocks 提取代码块
func ExtractCodeBlocks(content string) map[string]string {
    blocks := make(map[string]string)
    
    // 正则表达式匹配代码块
    re := regexp.MustCompile("```(\\w+)?\\n([\\s\\S]*?)```")
    matches := re.FindAllStringSubmatch(content, -1)
    
    for i, match := range matches {
        language := "text"
        if len(match[1]) > 0 {
            language = match[1]
        }
        
        key := fmt.Sprintf("%s_%d", language, i)
        blocks[key] = strings.TrimSpace(match[2])
    }
    
    return blocks
}

// ParseComplexityFromString 从字符串解析复杂度
func ParseComplexityFromString(complexityStr string) TaskComplexity {
    switch strings.ToLower(strings.TrimSpace(complexityStr)) {
    case "low", "简单", "低":
        return TaskComplexityLow
    case "medium", "中等", "中":
        return TaskComplexityMedium
    case "high", "高", "复杂":
        return TaskComplexityHigh
    case "very_high", "very high", "极高", "非常复杂":
        return TaskComplexityVeryHigh
    default:
        return TaskComplexityUnknown
    }
}
```

### 业务逻辑辅助函数

```go
// GenerateSessionID 生成会话ID
func GenerateSessionID() string {
    return fmt.Sprintf("session_%d_%s", time.Now().Unix(), generateRandomString(8))
}

// GeneratePlanID 生成规划ID
func GeneratePlanID() string {
    return fmt.Sprintf("plan_%d_%s", time.Now().Unix(), generateRandomString(6))
}

// GenerateStepID 生成步骤ID
func GenerateStepID(planID string, stepIndex int) string {
    return fmt.Sprintf("%s_step_%d", planID, stepIndex)
}

// generateRandomString 生成随机字符串
func generateRandomString(length int) string {
    const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
    b := make([]byte, length)
    for i := range b {
        b[i] = charset[rand.Intn(len(charset))]
    }
    return string(b)
}

// CalculateProgress 计算进度
func CalculateProgress(plan *TaskPlan) float64 {
    if plan.TotalSteps == 0 {
        return 0.0
    }
    return float64(plan.CompletedSteps) / float64(plan.TotalSteps)
}

// EstimateRemainingTime 估算剩余时间
func EstimateRemainingTime(plan *TaskPlan, executionHistory []*ExecutionRecord) time.Duration {
    if len(executionHistory) == 0 {
        return 0
    }
    
    // 计算平均执行时间
    var totalDuration time.Duration
    completedSteps := 0
    
    for _, record := range executionHistory {
        if record.Status == ExecutionStatusSuccess {
            totalDuration += record.Duration
            completedSteps++
        }
    }
    
    if completedSteps == 0 {
        return 0
    }
    
    avgDuration := totalDuration / time.Duration(completedSteps)
    remainingSteps := plan.TotalSteps - plan.CompletedSteps
    
    return avgDuration * time.Duration(remainingSteps)
}

// FilterRelevantHistory 过滤相关历史
func FilterRelevantHistory(messages []*schema.Message, currentQuery string, maxLength int) []*schema.Message {
    if len(messages) <= maxLength {
        return messages
    }
    
    // 简单策略：保留最近的消息
    // 更复杂的策略可以基于语义相似度
    start := len(messages) - maxLength
    return messages[start:]
}

// CompressConversationHistory 压缩对话历史
func CompressConversationHistory(messages []*schema.Message, targetLength int) []*schema.Message {
    if len(messages) <= targetLength {
        return messages
    }
    
    // 保留第一条和最后几条消息
    keepLast := targetLength - 1
    if keepLast <= 0 {
        return messages[len(messages)-1:]
    }
    
    result := []*schema.Message{messages[0]}
    result = append(result, messages[len(messages)-keepLast:]...)
    
    return result
}

// ValidateConfig 验证配置
func validateConfig(config *EnhancedMultiAgentConfig) error {
    if config.Host == nil {
        return errors.New("host configuration is required")
    }
    
    if len(config.Specialists) == 0 {
        return errors.New("at least one specialist is required")
    }
    
    if config.MaxRounds <= 0 {
        return errors.New("max_rounds must be positive")
    }
    
    // 验证专家配置
    for name, specialist := range config.Specialists {
        if specialist.Name == "" {
            return fmt.Errorf("specialist %s must have a name", name)
        }
        
        if specialist.ChatModel == nil {
            return fmt.Errorf("specialist %s must have a chat model", name)
        }
    }
    
    return nil
}

// CalculateConfidenceScore 计算置信度分数
func CalculateConfidenceScore(results map[string]*SpecialistResult) float64 {
    if len(results) == 0 {
        return 0.0
    }
    
    var totalConfidence float64
    var totalWeight float64
    
    for _, result := range results {
        weight := 1.0
        if result.Success {
            weight = 1.5 // 成功的结果权重更高
        }
        
        totalConfidence += result.Confidence * weight
        totalWeight += weight
    }
    
    return totalConfidence / totalWeight
}

// CalculateQualityScore 计算质量分数
func CalculateQualityScore(results map[string]*SpecialistResult) float64 {
    if len(results) == 0 {
        return 0.0
    }
    
    var totalQuality float64
    count := 0
    
    for _, result := range results {
        if result.Success {
            totalQuality += result.QualityScore
            count++
        }
    }
    
    if count == 0 {
        return 0.0
    }
    
    return totalQuality / float64(count)
}
```

### 性能监控辅助函数

```go
// PerformanceMetric 性能指标
type PerformanceMetric struct {
    Name      string    `json:"name"`
    Value     float64   `json:"value"`
    Unit      string    `json:"unit"`
    Timestamp time.Time `json:"timestamp"`
    Labels    map[string]string `json:"labels"`
}

// ResourceUsage 资源使用情况
type ResourceUsage struct {
    MemoryUsage    int64     `json:"memory_usage"`    // 内存使用量（字节）
    CPUUsage       float64   `json:"cpu_usage"`       // CPU使用率（百分比）
    GoroutineCount int       `json:"goroutine_count"` // Goroutine数量
    Timestamp      time.Time `json:"timestamp"`
}

// CollectResourceUsage 收集资源使用情况
func CollectResourceUsage() *ResourceUsage {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    
    return &ResourceUsage{
        MemoryUsage:    int64(m.Alloc),
        GoroutineCount: runtime.NumGoroutine(),
        Timestamp:      time.Now(),
    }
}

// CreatePerformanceMetric 创建性能指标
func CreatePerformanceMetric(name string, value float64, unit string, labels map[string]string) *PerformanceMetric {
    return &PerformanceMetric{
        Name:      name,
        Value:     value,
        Unit:      unit,
        Timestamp: time.Now(),
        Labels:    labels,
    }
}
```