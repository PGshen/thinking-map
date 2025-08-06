package enhanced_multiagent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino/schema"
)

// 节点处理器类型定义
type StatePreHandler func(ctx context.Context, input *schema.Message, state *EnhancedState) (*schema.Message, error)
type StatePostHandler func(ctx context.Context, output *schema.Message, state *EnhancedState) (*schema.Message, error)

// NodeHandlers 节点处理器集合
type NodeHandlers struct {
	PreHandler  StatePreHandler
	PostHandler StatePostHandler
}

// HostThinkNodeHandlers Host思考节点处理器
func HostThinkNodeHandlers() *NodeHandlers {
	return &NodeHandlers{
		PreHandler:  hostThinkPreHandler,
		PostHandler: hostThinkPostHandler,
	}
}

// hostThinkPreHandler Host思考节点前置处理器
func hostThinkPreHandler(ctx context.Context, input *schema.Message, state *EnhancedState) (*schema.Message, error) {
	// 保存原始消息
	if len(state.OriginalMessages) == 0 {
		state.OriginalMessages = []*schema.Message{input}
	}

	// 构建思考提示
	thinkingPrompt := buildThinkingPrompt(input, state)
	
	return &schema.Message{
		Role:    schema.User,
		Content: thinkingPrompt,
	}, nil
}

// hostThinkPostHandler Host思考节点后置处理器
func hostThinkPostHandler(ctx context.Context, output *schema.Message, state *EnhancedState) (*schema.Message, error) {
	// 解析思考结果
	thinkingResult, err := parseThinkingResult(output)
	if err != nil {
		return nil, fmt.Errorf("failed to parse thinking result: %w", err)
	}

	// 更新状态
	state.AddThinkingResult(thinkingResult)

	return output, nil
}

// DirectAnswerNodeHandlers 直接回答节点处理器
func DirectAnswerNodeHandlers() *NodeHandlers {
	return &NodeHandlers{
		PreHandler:  directAnswerPreHandler,
		PostHandler: directAnswerPostHandler,
	}
}

// directAnswerPreHandler 直接回答节点前置处理器
func directAnswerPreHandler(ctx context.Context, input *schema.Message, state *EnhancedState) (*schema.Message, error) {
	// 构建直接回答提示
	directPrompt := buildDirectAnswerPrompt(state)
	
	return &schema.Message{
		Role:    schema.User,
		Content: directPrompt,
	}, nil
}

// directAnswerPostHandler 直接回答节点后置处理器
func directAnswerPostHandler(ctx context.Context, output *schema.Message, state *EnhancedState) (*schema.Message, error) {
	// 标记任务完成
	state.IsSimpleTask = true
	state.IsCompleted = true
	state.FinalAnswer = output

	return output, nil
}

// PlanCreationNodeHandlers 规划创建节点处理器
func PlanCreationNodeHandlers() *NodeHandlers {
	return &NodeHandlers{
		PreHandler:  planCreationPreHandler,
		PostHandler: planCreationPostHandler,
	}
}

// planCreationPreHandler 规划创建节点前置处理器
func planCreationPreHandler(ctx context.Context, input *schema.Message, state *EnhancedState) (*schema.Message, error) {
	// 构建规划提示
	planningPrompt := buildPlanningPrompt(state)
	
	return &schema.Message{
		Role:    schema.User,
		Content: planningPrompt,
	}, nil
}

// planCreationPostHandler 规划创建节点后置处理器
func planCreationPostHandler(ctx context.Context, output *schema.Message, state *EnhancedState) (*schema.Message, error) {
	// 解析规划结果
	plan, err := parsePlanningResult(output)
	if err != nil {
		return nil, fmt.Errorf("failed to parse planning result: %w", err)
	}

	// 更新状态
	state.CurrentPlan = plan
	state.IsSimpleTask = false

	return output, nil
}

// PlanExecutionNodeHandlers 规划执行节点处理器
func PlanExecutionNodeHandlers() *NodeHandlers {
	return &NodeHandlers{
		PreHandler:  planExecutionPreHandler,
		PostHandler: planExecutionPostHandler,
	}
}

// planExecutionPreHandler 规划执行节点前置处理器
func planExecutionPreHandler(ctx context.Context, input *schema.Message, state *EnhancedState) (*schema.Message, error) {
	// 获取下一个待执行的步骤
	nextStep := state.GetNextPendingStep()
	if nextStep == nil {
		return nil, fmt.Errorf("no pending step found")
	}

	// 更新步骤状态为执行中
	err := state.UpdateStepStatus(nextStep.ID, StepStatusExecuting)
	if err != nil {
		return nil, fmt.Errorf("failed to update step status: %w", err)
	}

	// 创建执行上下文
	executionContext := &ExecutionContext{
		TaskDescription: nextStep.Description,
		PlanStep:        nextStep,
		Messages:        state.OriginalMessages,
		Parameters:      make(map[string]interface{}),
		ExpectedFormat:  "text",
	}

	state.CurrentExecution = executionContext
	state.CurrentPlan.CurrentStep = nextStep.ID

	return input, nil
}

// planExecutionPostHandler 规划执行节点后置处理器
func planExecutionPostHandler(ctx context.Context, output *schema.Message, state *EnhancedState) (*schema.Message, error) {
	// 执行准备完成，返回原始输出
	return output, nil
}

// SpecialistNodeHandlers 专家节点处理器
func SpecialistNodeHandlers(specialistName string) *NodeHandlers {
	return &NodeHandlers{
		PreHandler:  func(ctx context.Context, input *schema.Message, state *EnhancedState) (*schema.Message, error) {
			return specialistPreHandler(ctx, input, state, specialistName)
		},
		PostHandler: func(ctx context.Context, output *schema.Message, state *EnhancedState) (*schema.Message, error) {
			return specialistPostHandler(ctx, output, state, specialistName)
		},
	}
}

// specialistPreHandler 专家节点前置处理器
func specialistPreHandler(ctx context.Context, input *schema.Message, state *EnhancedState, specialistName string) (*schema.Message, error) {
	// 构建专家提示
	specialistPrompt := buildSpecialistPrompt(state, specialistName)
	
	return &schema.Message{
		Role:    schema.User,
		Content: specialistPrompt,
	}, nil
}

// specialistPostHandler 专家节点后置处理器
func specialistPostHandler(ctx context.Context, output *schema.Message, state *EnhancedState, specialistName string) (*schema.Message, error) {
	// 创建专家结果
	result := &SpecialistResult{
		SpecialistName: specialistName,
		Result:         output,
		Status:         ExecutionStatusSuccess, // 默认成功，实际应根据输出判断
		Duration:       time.Second,            // 实际应记录真实执行时间
		Confidence:     0.8,                    // 实际应根据输出评估置信度
	}

	// 保存到状态
	if state.CurrentSpecialistResults == nil {
		state.CurrentSpecialistResults = make(map[string]*SpecialistResult)
	}
	state.CurrentSpecialistResults[specialistName] = result

	return output, nil
}

// ResultCollectorNodeHandlers 结果收集节点处理器
func ResultCollectorNodeHandlers() *NodeHandlers {
	return &NodeHandlers{
		PreHandler:  resultCollectorPreHandler,
		PostHandler: resultCollectorPostHandler,
	}
}

// resultCollectorPreHandler 结果收集节点前置处理器
func resultCollectorPreHandler(ctx context.Context, input *schema.Message, state *EnhancedState) (*schema.Message, error) {
	// 收集所有专家结果
	collectedResults := NewCollectedResults()
	
	for _, result := range state.CurrentSpecialistResults {
		collectedResults.AddResult(result)
	}

	// 生成汇总
	collectedResults.Summary = generateResultSummary(collectedResults)

	// 更新状态
	state.CurrentCollectedResults = collectedResults
	state.AddExecutionRecord(collectedResults)

	return input, nil
}

// resultCollectorPostHandler 结果收集节点后置处理器
func resultCollectorPostHandler(ctx context.Context, output *schema.Message, state *EnhancedState) (*schema.Message, error) {
	return output, nil
}

// FeedbackProcessorNodeHandlers 反馈处理节点处理器
func FeedbackProcessorNodeHandlers() *NodeHandlers {
	return &NodeHandlers{
		PreHandler:  feedbackProcessorPreHandler,
		PostHandler: feedbackProcessorPostHandler,
	}
}

// feedbackProcessorPreHandler 反馈处理节点前置处理器
func feedbackProcessorPreHandler(ctx context.Context, input *schema.Message, state *EnhancedState) (*schema.Message, error) {
	// 构建反馈分析提示
	feedbackPrompt := buildFeedbackPrompt(state)
	
	return &schema.Message{
		Role:    schema.User,
		Content: feedbackPrompt,
	}, nil
}

// feedbackProcessorPostHandler 反馈处理节点后置处理器
func feedbackProcessorPostHandler(ctx context.Context, output *schema.Message, state *EnhancedState) (*schema.Message, error) {
	// 解析反馈结果
	feedbackResult, err := parseFeedbackResult(output, state)
	if err != nil {
		return nil, fmt.Errorf("failed to parse feedback result: %w", err)
	}

	// 更新状态
	state.CurrentFeedbackResult = feedbackResult

	return output, nil
}

// PlanUpdateNodeHandlers 规划更新节点处理器
func PlanUpdateNodeHandlers() *NodeHandlers {
	return &NodeHandlers{
		PreHandler:  planUpdatePreHandler,
		PostHandler: planUpdatePostHandler,
	}
}

// planUpdatePreHandler 规划更新节点前置处理器
func planUpdatePreHandler(ctx context.Context, input *schema.Message, state *EnhancedState) (*schema.Message, error) {
	// 构建规划更新提示
	updatePrompt := buildPlanUpdatePrompt(state)
	
	return &schema.Message{
		Role:    schema.User,
		Content: updatePrompt,
	}, nil
}

// planUpdatePostHandler 规划更新节点后置处理器
func planUpdatePostHandler(ctx context.Context, output *schema.Message, state *EnhancedState) (*schema.Message, error) {
	// 解析更新后的规划
	updatedPlan, err := parseUpdatedPlan(output, state.CurrentPlan)
	if err != nil {
		return nil, fmt.Errorf("failed to parse updated plan: %w", err)
	}

	// 更新状态
	state.CurrentPlan = updatedPlan
	state.CurrentRound++

	return output, nil
}

// FinalAnswerNodeHandlers 最终答案节点处理器
func FinalAnswerNodeHandlers() *NodeHandlers {
	return &NodeHandlers{
		PreHandler:  finalAnswerPreHandler,
		PostHandler: finalAnswerPostHandler,
	}
}

// finalAnswerPreHandler 最终答案节点前置处理器
func finalAnswerPreHandler(ctx context.Context, input *schema.Message, state *EnhancedState) (*schema.Message, error) {
	// 构建最终答案提示
	finalPrompt := buildFinalAnswerPrompt(state)
	
	return &schema.Message{
		Role:    schema.User,
		Content: finalPrompt,
	}, nil
}

// finalAnswerPostHandler 最终答案节点后置处理器
func finalAnswerPostHandler(ctx context.Context, output *schema.Message, state *EnhancedState) (*schema.Message, error) {
	// 标记任务完成
	state.IsCompleted = true
	state.FinalAnswer = output

	return output, nil
}

// 分支判断函数

// ComplexityBranchCondition 复杂度分支判断
func ComplexityBranchCondition(state *EnhancedState) string {
	if state.CurrentThinkingResult == nil {
		return "complex_task" // 默认复杂任务
	}

	switch state.CurrentThinkingResult.Complexity {
	case TaskComplexitySimple:
		return "direct_answer"
	default:
		return "complex_task"
	}
}

// ReflectionBranchCondition 反思分支判断
func ReflectionBranchCondition(state *EnhancedState) string {
	if state.CurrentFeedbackResult == nil {
		return "continue" // 默认继续
	}

	if state.CurrentFeedbackResult.ShouldContinue && !state.IsMaxRoundsReached() {
		return "continue"
	}

	return "complete"
}

// SpecialistBranchCondition 专家分支判断
func SpecialistBranchCondition(state *EnhancedState) []string {
	if state.CurrentExecution == nil || state.CurrentExecution.PlanStep == nil {
		return []string{} // 没有可执行的步骤
	}

	// 根据步骤的AssignedTo字段确定需要调用的专家
	assignedTo := state.CurrentExecution.PlanStep.AssignedTo
	if assignedTo != "" {
		return []string{assignedTo}
	}

	// 如果没有明确分配，根据任务描述智能选择
	return intelligentSpecialistSelection(state.CurrentExecution.TaskDescription)
}

// intelligentSpecialistSelection 智能专家选择
func intelligentSpecialistSelection(taskDescription string) []string {
	// 简单的关键词匹配逻辑，实际应该更智能
	taskLower := strings.ToLower(taskDescription)
	
	var specialists []string
	
	if strings.Contains(taskLower, "code") || strings.Contains(taskLower, "program") {
		specialists = append(specialists, "code_specialist")
	}
	
	if strings.Contains(taskLower, "research") || strings.Contains(taskLower, "analyze") {
		specialists = append(specialists, "research_specialist")
	}
	
	if strings.Contains(taskLower, "write") || strings.Contains(taskLower, "document") {
		specialists = append(specialists, "writing_specialist")
	}
	
	// 如果没有匹配到特定专家，使用通用专家
	if len(specialists) == 0 {
		specialists = append(specialists, "general_specialist")
	}
	
	return specialists
}

// 辅助函数（占位符实现）

// buildThinkingPrompt 构建思考提示
func buildThinkingPrompt(input *schema.Message, state *EnhancedState) string {
	return fmt.Sprintf(`请分析以下任务的复杂度并制定处理策略：

任务：%s

请按以下格式回答：
{
  "thought": "你的思考过程",
  "complexity": "simple|moderate|complex",
  "reasoning": "复杂度判断的理由",
  "next_action": "direct_answer|create_plan"
}`, input.Content)
}

// buildDirectAnswerPrompt 构建直接回答提示
func buildDirectAnswerPrompt(state *EnhancedState) string {
	originalQuery := ""
	if len(state.OriginalMessages) > 0 {
		originalQuery = state.OriginalMessages[0].Content
	}
	return fmt.Sprintf("请直接回答以下问题：\n\n%s", originalQuery)
}

// buildPlanningPrompt 构建规划提示
func buildPlanningPrompt(state *EnhancedState) string {
	originalQuery := ""
	if len(state.OriginalMessages) > 0 {
		originalQuery = state.OriginalMessages[0].Content
	}
	return fmt.Sprintf(`请为以下复杂任务制定详细的执行计划：

任务：%s

请按以下JSON格式回答：
{
  "content": "计划的Markdown描述",
  "steps": [
    {
      "description": "步骤描述",
      "assigned_to": "负责的专家",
      "priority": 1,
      "dependencies": []
    }
  ]
}`, originalQuery)
}

// buildSpecialistPrompt 构建专家提示
func buildSpecialistPrompt(state *EnhancedState, specialistName string) string {
	if state.CurrentExecution == nil {
		return "没有可执行的任务"
	}
	return fmt.Sprintf("作为%s，请执行以下任务：\n\n%s", specialistName, state.CurrentExecution.TaskDescription)
}

// buildFeedbackPrompt 构建反馈提示
func buildFeedbackPrompt(state *EnhancedState) string {
	summary := ""
	if state.CurrentCollectedResults != nil {
		summary = state.CurrentCollectedResults.Summary
	}
	return fmt.Sprintf(`请分析以下执行结果并提供反馈：

执行结果：%s

请按以下JSON格式回答：
{
  "feedback": "你的反馈",
  "should_continue": true/false,
  "suggested_action": "continue|complete",
  "plan_update_suggestion": "规划更新建议（如果需要）"
}`, summary)
}

// buildPlanUpdatePrompt 构建规划更新提示
func buildPlanUpdatePrompt(state *EnhancedState) string {
	feedback := ""
	if state.CurrentFeedbackResult != nil {
		feedback = state.CurrentFeedbackResult.Feedback
	}
	return fmt.Sprintf("根据以下反馈更新执行计划：\n\n%s", feedback)
}

// buildFinalAnswerPrompt 构建最终答案提示
func buildFinalAnswerPrompt(state *EnhancedState) string {
	summary := ""
	if state.CurrentCollectedResults != nil {
		summary = state.CurrentCollectedResults.Summary
	}
	return fmt.Sprintf("基于以下执行结果，生成最终答案：\n\n%s", summary)
}

// generateResultSummary 生成结果汇总
func generateResultSummary(results *CollectedResults) string {
	successCount := len(results.SuccessfulResults)
	failCount := len(results.FailedResults)
	total := len(results.Results)
	
	return fmt.Sprintf("执行完成：总计%d个任务，成功%d个，失败%d个", total, successCount, failCount)
}

// 解析函数（占位符实现）

// parseThinkingResult 解析思考结果
func parseThinkingResult(output *schema.Message) (*ThinkingResult, error) {
	// 尝试解析JSON格式的思考结果
	var result struct {
		Thought    string `json:"thought"`
		Complexity string `json:"complexity"`
		Reasoning  string `json:"reasoning"`
		NextAction string `json:"next_action"`
	}
	
	err := json.Unmarshal([]byte(output.Content), &result)
	if err != nil {
		// 如果JSON解析失败，使用默认值
		return &ThinkingResult{
			Thought:    output.Content,
			Complexity: TaskComplexityModerate,
			Reasoning:  "无法解析复杂度",
			NextAction: ActionTypeCreatePlan,
			Timestamp:  time.Now(),
		}, nil
	}
	
	// 转换复杂度
	complexity := TaskComplexityModerate
	switch result.Complexity {
	case "simple":
		complexity = TaskComplexitySimple
	case "complex":
		complexity = TaskComplexityComplex
	}
	
	// 转换行动类型
	nextAction := ActionTypeCreatePlan
	switch result.NextAction {
	case "direct_answer":
		nextAction = ActionTypeDirectAnswer
	}
	
	return &ThinkingResult{
		Thought:    result.Thought,
		Complexity: complexity,
		Reasoning:  result.Reasoning,
		NextAction: nextAction,
		Timestamp:  time.Now(),
	}, nil
}

// parsePlanningResult 解析规划结果
func parsePlanningResult(output *schema.Message) (*TaskPlan, error) {
	// 尝试解析JSON格式的规划结果
	var result struct {
		Content string `json:"content"`
		Steps   []struct {
			Description  string `json:"description"`
			AssignedTo   string `json:"assigned_to"`
			Priority     int    `json:"priority"`
			Dependencies []int  `json:"dependencies"`
		} `json:"steps"`
	}
	
	err := json.Unmarshal([]byte(output.Content), &result)
	if err != nil {
		// 如果JSON解析失败，创建简单规划
		plan := NewTaskPlan(output.Content)
		plan.AddStep("执行任务", "general_specialist", 1, nil)
		return plan, nil
	}
	
	plan := NewTaskPlan(result.Content)
	for _, step := range result.Steps {
		plan.AddStep(step.Description, step.AssignedTo, step.Priority, step.Dependencies)
	}
	
	return plan, nil
}

// parseFeedbackResult 解析反馈结果
func parseFeedbackResult(output *schema.Message, state *EnhancedState) (*FeedbackResult, error) {
	// 尝试解析JSON格式的反馈结果
	var result struct {
		Feedback             string `json:"feedback"`
		ShouldContinue       bool   `json:"should_continue"`
		SuggestedAction      string `json:"suggested_action"`
		PlanUpdateSuggestion string `json:"plan_update_suggestion"`
	}
	
	err := json.Unmarshal([]byte(output.Content), &result)
	if err != nil {
		// 如果JSON解析失败，使用默认值
		return &FeedbackResult{
			Feedback:             output.Content,
			ShouldContinue:       false,
			SuggestedAction:      ActionTypeReflect,
			CollectedResults:     state.CurrentCollectedResults,
		}, nil
	}
	
	// 转换行动类型
	suggestedAction := ActionTypeReflect
	switch result.SuggestedAction {
	case "continue":
		suggestedAction = ActionTypeUpdatePlan
	case "complete":
		suggestedAction = ActionTypeReflect
	}
	
	return &FeedbackResult{
		Feedback:             result.Feedback,
		ShouldContinue:       result.ShouldContinue,
		SuggestedAction:      suggestedAction,
		PlanUpdateSuggestion: result.PlanUpdateSuggestion,
		CollectedResults:     state.CurrentCollectedResults,
	}, nil
}

// parseUpdatedPlan 解析更新后的规划
func parseUpdatedPlan(output *schema.Message, currentPlan *TaskPlan) (*TaskPlan, error) {
	// 简单实现：基于当前规划创建新版本
	updatedPlan := currentPlan.Clone()
	if updatedPlan == nil {
		return currentPlan, nil
	}
	
	updatedPlan.UpdateVersion(PlanUpdateTypeModifyPlan, "根据反馈更新规划", map[string]interface{}{
		"feedback": output.Content,
	})
	
	return updatedPlan, nil
}