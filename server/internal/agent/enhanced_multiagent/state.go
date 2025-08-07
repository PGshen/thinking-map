package enhanced_multiagent

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
)

// NewEnhancedState 创建新的增强状态
func NewEnhancedState(sessionID string) *EnhancedState {
	if sessionID == "" {
		sessionID = uuid.New().String()
	}

	return &EnhancedState{
		OriginalMessages:         []*schema.Message{},
		ConversationContext:      nil,
		CurrentPlan:              nil,
		CurrentExecution:         nil,
		CurrentSpecialistResults: make(map[string]*SpecialistResult),
		CurrentCollectedResults:  nil,
		CurrentFeedbackResult:    nil,
		CurrentThinkingResult:    nil,
		ExecutionHistory:         []*ExecutionRecord{},
		ThinkingHistory:          []*ThinkingResult{},
		CurrentRound:             0,
		MaxRounds:                5, // 默认最大轮次
		IsSimpleTask:             false,
		IsCompleted:              false,
		FinalAnswer:              nil,
		SessionID:                sessionID,
		CallTimestamp:            time.Now(),
	}
}

// InitializeFromMessages 从消息历史初始化状态
func (s *EnhancedState) InitializeFromMessages(messages []*schema.Message) {
	s.OriginalMessages = messages
	s.CallTimestamp = time.Now()

	// 分析对话上下文
	s.ConversationContext = analyzeConversationContext(messages)
}

// AddThinkingResult 添加思考结果
func (s *EnhancedState) AddThinkingResult(result *ThinkingResult) {
	s.CurrentThinkingResult = result
	s.ThinkingHistory = append(s.ThinkingHistory, result)
}

// SetPlan 设置当前计划
func (s *EnhancedState) SetPlan(plan *TaskPlan) {
	s.CurrentPlan = plan
}

// UpdatePlan 更新计划
func (s *EnhancedState) UpdatePlan(newPlan *TaskPlan, updateType PlanUpdateType, reason string) {
	if s.CurrentPlan != nil {
		// 记录更新历史
		changesList := extractPlanChanges(s.CurrentPlan, newPlan)
		changesMap := make(map[string]interface{})
		for i, change := range changesList {
			changesMap[fmt.Sprintf("change_%d", i)] = change
		}
		update := &PlanUpdate{
			Version:     newPlan.Version,
			UpdateType:  updateType,
			Description: reason,
			Timestamp:   time.Now(),
			Changes:     changesMap,
		}
		newPlan.UpdateHistory = append(newPlan.UpdateHistory, update)
	}
	s.CurrentPlan = newPlan
}

// AddSpecialistResult 添加专家执行结果
func (s *EnhancedState) AddSpecialistResult(specialistName string, result *SpecialistResult) {
	s.CurrentSpecialistResults[specialistName] = result
}

// SetCollectedResults 设置收集的结果
func (s *EnhancedState) SetCollectedResults(results *CollectedResults) {
	s.CurrentCollectedResults = results
}

// SetFeedbackResult 设置反馈结果
func (s *EnhancedState) SetFeedbackResult(feedback *FeedbackResult) {
	s.CurrentFeedbackResult = feedback
}

// NextRound 进入下一轮执行
func (s *EnhancedState) NextRound() {
	// 记录当前轮次的执行历史
	if s.CurrentCollectedResults != nil && s.CurrentFeedbackResult != nil {
		record := &ExecutionRecord{
			Round:     s.CurrentRound,
			Results:   s.CurrentCollectedResults,
			Feedback:  s.CurrentFeedbackResult,
			Duration:  time.Since(s.CallTimestamp),
			Timestamp: time.Now(),
			Status:    ExecutionStatusCompleted,
			Metadata:  make(map[string]interface{}),
		}
		s.ExecutionHistory = append(s.ExecutionHistory, record)
	}

	// 进入下一轮
	s.CurrentRound++

	// 清理当前轮次的临时状态
	s.CurrentSpecialistResults = make(map[string]*SpecialistResult)
	s.CurrentCollectedResults = nil
	s.CurrentFeedbackResult = nil
}

// Complete 标记任务完成
func (s *EnhancedState) Complete(finalAnswer *schema.Message) {
	s.IsCompleted = true
	s.FinalAnswer = finalAnswer

	// 记录最后一轮的执行历史
	if s.CurrentCollectedResults != nil {
		record := &ExecutionRecord{
			Round:     s.CurrentRound,
			Results:   s.CurrentCollectedResults,
			Feedback:  s.CurrentFeedbackResult,
			Duration:  time.Since(s.CallTimestamp),
			Timestamp: time.Now(),
			Status:    ExecutionStatusCompleted,
			Metadata:  make(map[string]interface{}),
		}
		s.ExecutionHistory = append(s.ExecutionHistory, record)
	}
}

// ShouldContinue 判断是否应该继续执行
func (s *EnhancedState) ShouldContinue() bool {
	if s.IsCompleted {
		return false
	}
	if s.CurrentRound >= s.MaxRounds {
		return false
	}
	if s.CurrentFeedbackResult != nil {
		return s.CurrentFeedbackResult.ShouldContinue
	}
	return true
}

// GetCurrentStep 获取当前执行步骤
func (s *EnhancedState) GetCurrentStep() *PlanStep {
	if s.CurrentPlan == nil {
		return nil
	}
	return getCurrentExecutionStep(s.CurrentPlan, s)
}

// GetNextExecutableStep 获取下一个可执行步骤
func (s *EnhancedState) GetNextExecutableStep() *PlanStep {
	if s.CurrentPlan == nil {
		return nil
	}
	return findNextExecutableStep(s.CurrentPlan)
}

// ToJSON 将状态序列化为JSON
func (s *EnhancedState) ToJSON() ([]byte, error) {
	return json.Marshal(s)
}

// FromJSON 从JSON反序列化状态
func (s *EnhancedState) FromJSON(data []byte) error {
	return json.Unmarshal(data, s)
}

// Clone 克隆状态
func (s *EnhancedState) Clone() *EnhancedState {
	data, err := s.ToJSON()
	if err != nil {
		return nil
	}

	newState := &EnhancedState{}
	if err := newState.FromJSON(data); err != nil {
		return nil
	}

	return newState
}

// 辅助函数

// analyzeConversationContext 分析对话上下文
func analyzeConversationContext(messages []*schema.Message) *ConversationContext {
	if len(messages) == 0 {
		return &ConversationContext{
			TurnCount:      0,
			IsFirstTurn:    true,
			IsContinuation: false,
			AnalyzedAt:     time.Now(),
			Metadata:       make(map[string]interface{}),
		}
	}

	// 计算对话轮次
	turnCount := 0
	for _, msg := range messages {
		if msg.Role == schema.User {
			turnCount++
		}
	}

	// 获取最新的用户和助手消息
	var latestUserMsg, latestAssistantMsg *schema.Message
	for i := len(messages) - 1; i >= 0; i-- {
		msg := messages[i]
		if msg.Role == schema.User && latestUserMsg == nil {
			latestUserMsg = msg
		}
		if msg.Role == schema.Assistant && latestAssistantMsg == nil {
			latestAssistantMsg = msg
		}
		if latestUserMsg != nil && latestAssistantMsg != nil {
			break
		}
	}

	// 简单的意图分析（实际实现中可能需要更复杂的NLP处理）
	userIntent := "unknown"
	intentConfidence := 0.5
	if latestUserMsg != nil {
		content := latestUserMsg.Content
		if len(content) > 0 {
			text := content
			// 简单的关键词匹配
			if containsAny(text, []string{"帮我", "请", "能否", "可以"}) {
				userIntent = "request"
				intentConfidence = 0.8
			} else if containsAny(text, []string{"什么", "如何", "为什么", "怎么"}) {
				userIntent = "question"
				intentConfidence = 0.8
			}
		}
	}

	return &ConversationContext{
		TurnCount:              turnCount,
		IsFirstTurn:            turnCount <= 1,
		IsContinuation:         turnCount > 1,
		LatestUserMessage:      latestUserMsg,
		LatestAssistantMessage: latestAssistantMsg,
		ConversationTopic:      extractTopic(messages),
		UserIntent:             userIntent,
		IntentConfidence:       intentConfidence,
		EmotionalTone:          "neutral",
		KeyEntities:            extractEntities(messages),
		ComplexityHint:         TaskComplexityModerate,
		RelevantHistory:        messages,
		ContextSummary:         generateContextSummary(messages),
		AnalyzedAt:             time.Now(),
		Metadata:               make(map[string]interface{}),
	}
}

// containsAny 检查文本是否包含任何关键词
func containsAny(text string, keywords []string) bool {
	for _, keyword := range keywords {
		if len(text) > 0 && len(keyword) > 0 {
			// 简单的包含检查
			for i := 0; i <= len(text)-len(keyword); i++ {
				if text[i:i+len(keyword)] == keyword {
					return true
				}
			}
		}
	}
	return false
}

// extractTopic 提取对话主题
func extractTopic(messages []*schema.Message) string {
	if len(messages) == 0 {
		return "general"
	}
	// 简单实现：返回第一个用户消息的前几个词
	for _, msg := range messages {
		if msg.Role == schema.User && len(msg.Content) > 0 {
			content := msg.Content
			if len(content) > 20 {
				return content[:20] + "..."
			}
			return content
		}
	}
	return "general"
}

// extractEntities 提取关键实体
func extractEntities(messages []*schema.Message) []string {
	// 简单实现：返回空列表
	// 实际实现中可能需要NER处理
	return []string{}
}

// generateContextSummary 生成上下文摘要
func generateContextSummary(messages []*schema.Message) string {
	if len(messages) == 0 {
		return "Empty conversation"
	}
	return fmt.Sprintf("Conversation with %d messages", len(messages))
}

// getCurrentExecutionStep 获取当前执行步骤
func getCurrentExecutionStep(plan *TaskPlan, state *EnhancedState) *PlanStep {
	if plan == nil || len(plan.Steps) == 0 {
		return nil
	}

	// 返回当前步骤
	if plan.CurrentStep >= 0 && plan.CurrentStep < len(plan.Steps) {
		return plan.Steps[plan.CurrentStep]
	}

	return nil
}

// findNextExecutableStep 查找下一个可执行步骤
func findNextExecutableStep(plan *TaskPlan) *PlanStep {
	if plan == nil || len(plan.Steps) == 0 {
		return nil
	}

	// 查找第一个未完成的步骤
	for _, step := range plan.Steps {
		if step.Status == StepStatusPending || step.Status == StepStatusExecuting {
			return step
		}
	}

	return nil
}

// extractPlanChanges 提取计划变更
func extractPlanChanges(oldPlan, newPlan *TaskPlan) []string {
	changes := []string{}

	if oldPlan == nil {
		changes = append(changes, "Initial plan created")
		return changes
	}

	if newPlan == nil {
		changes = append(changes, "Plan removed")
		return changes
	}

	// 比较步骤数量
	if len(oldPlan.Steps) != len(newPlan.Steps) {
		changes = append(changes, fmt.Sprintf("Step count changed from %d to %d", len(oldPlan.Steps), len(newPlan.Steps)))
	}

	// 比较内容
	if oldPlan.Content != newPlan.Content {
		changes = append(changes, "Plan content updated")
	}

	return changes
}
