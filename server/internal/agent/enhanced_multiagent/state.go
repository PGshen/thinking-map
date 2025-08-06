package enhanced_multiagent

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cloudwego/eino/schema"
)

// NewEnhancedState 创建新的增强状态
func NewEnhancedState(maxRounds int) *EnhancedState {
	return &EnhancedState{
		OriginalMessages:         make([]*schema.Message, 0),
		CurrentSpecialistResults: make(map[string]*SpecialistResult),
		ExecutionHistory:         make([]*ExecutionRecord, 0),
		ThinkingHistory:          make([]*ThinkingResult, 0),
		CurrentRound:             0,
		MaxRounds:                maxRounds,
		IsSimpleTask:             false,
		IsCompleted:              false,
	}
}

// Clone 克隆状态（深拷贝）
func (s *EnhancedState) Clone() *EnhancedState {
	data, err := json.Marshal(s)
	if err != nil {
		return nil
	}
	
	var cloned EnhancedState
	err = json.Unmarshal(data, &cloned)
	if err != nil {
		return nil
	}
	
	return &cloned
}

// Checkpoint 创建状态检查点
func (s *EnhancedState) Checkpoint() *StateCheckpoint {
	return &StateCheckpoint{
		Timestamp: time.Now(),
		Round:     s.CurrentRound,
		State:     s.Clone(),
	}
}

// IsMaxRoundsReached 检查是否达到最大轮次
func (s *EnhancedState) IsMaxRoundsReached() bool {
	return s.CurrentRound >= s.MaxRounds
}

// AddThinkingResult 添加思考结果到历史
func (s *EnhancedState) AddThinkingResult(thinking *ThinkingResult) {
	s.CurrentThinkingResult = thinking
	s.ThinkingHistory = append(s.ThinkingHistory, thinking)
}

// AddExecutionRecord 添加执行记录到历史
func (s *EnhancedState) AddExecutionRecord(results *CollectedResults) {
	record := &ExecutionRecord{
		Round:     s.CurrentRound,
		Results:   results,
		Timestamp: time.Now(),
	}
	s.ExecutionHistory = append(s.ExecutionHistory, record)
}

// GetCurrentStep 获取当前执行步骤
func (s *EnhancedState) GetCurrentStep() *PlanStep {
	if s.CurrentPlan == nil || len(s.CurrentPlan.Steps) == 0 {
		return nil
	}
	
	for _, step := range s.CurrentPlan.Steps {
		if step.ID == s.CurrentPlan.CurrentStep {
			return step
		}
	}
	
	return nil
}

// GetNextPendingStep 获取下一个待执行的步骤
func (s *EnhancedState) GetNextPendingStep() *PlanStep {
	if s.CurrentPlan == nil {
		return nil
	}
	
	for _, step := range s.CurrentPlan.Steps {
		if step.Status == StepStatusPending {
			// 检查依赖是否满足
			if s.areDependenciesSatisfied(step) {
				return step
			}
		}
	}
	
	return nil
}

// areDependenciesSatisfied 检查步骤依赖是否满足
func (s *EnhancedState) areDependenciesSatisfied(step *PlanStep) bool {
	if len(step.Dependencies) == 0 {
		return true
	}
	
	for _, depID := range step.Dependencies {
		depStep := s.getStepByID(depID)
		if depStep == nil || depStep.Status != StepStatusCompleted {
			return false
		}
	}
	
	return true
}

// getStepByID 根据ID获取步骤
func (s *EnhancedState) getStepByID(id int) *PlanStep {
	if s.CurrentPlan == nil {
		return nil
	}
	
	for _, step := range s.CurrentPlan.Steps {
		if step.ID == id {
			return step
		}
	}
	
	return nil
}

// UpdateStepStatus 更新步骤状态
func (s *EnhancedState) UpdateStepStatus(stepID int, status StepStatus) error {
	step := s.getStepByID(stepID)
	if step == nil {
		return fmt.Errorf("step with ID %d not found", stepID)
	}
	
	step.Status = status
	step.UpdatedAt = time.Now()
	
	// 如果步骤完成，添加到已完成列表
	if status == StepStatusCompleted {
		s.CurrentPlan.CompletedSteps = append(s.CurrentPlan.CompletedSteps, stepID)
	}
	
	return nil
}

// IsAllStepsCompleted 检查所有步骤是否完成
func (s *EnhancedState) IsAllStepsCompleted() bool {
	if s.CurrentPlan == nil || len(s.CurrentPlan.Steps) == 0 {
		return false
	}
	
	for _, step := range s.CurrentPlan.Steps {
		if step.Status != StepStatusCompleted && step.Status != StepStatusSkipped {
			return false
		}
	}
	
	return true
}

// TaskPlan 相关方法

// NewTaskPlan 创建新的任务规划
func NewTaskPlan(content string) *TaskPlan {
	return &TaskPlan{
		Content:            content,
		CurrentStep:        0,
		TotalSteps:         0,
		CompletedSteps:     make([]int, 0),
		Steps:              make([]*PlanStep, 0),
		Version:            1,
		AllowDynamicUpdate: true,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
		UpdateHistory:      make([]*PlanUpdate, 0),
	}
}

// AddStep 添加步骤到规划
func (p *TaskPlan) AddStep(description, assignedTo string, priority int, dependencies []int) *PlanStep {
	step := &PlanStep{
		ID:           len(p.Steps) + 1,
		Description:  description,
		Status:       StepStatusPending,
		AssignedTo:   assignedTo,
		Priority:     priority,
		Dependencies: dependencies,
		RetryCount:   0,
		MaxRetries:   3,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	
	p.Steps = append(p.Steps, step)
	p.TotalSteps = len(p.Steps)
	p.UpdatedAt = time.Now()
	
	return step
}

// UpdateVersion 更新规划版本
func (p *TaskPlan) UpdateVersion(updateType PlanUpdateType, description string, changes map[string]interface{}) {
	p.Version++
	p.UpdatedAt = time.Now()
	
	update := &PlanUpdate{
		Version:     p.Version,
		UpdateType:  updateType,
		Description: description,
		Timestamp:   time.Now(),
		Changes:     changes,
	}
	
	p.UpdateHistory = append(p.UpdateHistory, update)
}

// Clone 克隆任务规划（深拷贝）
func (p *TaskPlan) Clone() *TaskPlan {
	data, err := json.Marshal(p)
	if err != nil {
		return nil
	}
	
	var cloned TaskPlan
	err = json.Unmarshal(data, &cloned)
	if err != nil {
		return nil
	}
	
	return &cloned
}

// CollectedResults 相关方法

// NewCollectedResults 创建新的收集结果
func NewCollectedResults() *CollectedResults {
	return &CollectedResults{
		Results:           make(map[string]*SpecialistResult),
		SuccessfulResults: make([]*SpecialistResult, 0),
		FailedResults:     make([]*SpecialistResult, 0),
	}
}

// AddResult 添加专家结果
func (c *CollectedResults) AddResult(result *SpecialistResult) {
	c.Results[result.SpecialistName] = result
	
	switch result.Status {
	case ExecutionStatusSuccess:
		c.SuccessfulResults = append(c.SuccessfulResults, result)
	case ExecutionStatusFailed, ExecutionStatusPartial:
		c.FailedResults = append(c.FailedResults, result)
	}
}

// HasSuccessfulResults 检查是否有成功的结果
func (c *CollectedResults) HasSuccessfulResults() bool {
	return len(c.SuccessfulResults) > 0
}

// HasFailedResults 检查是否有失败的结果
func (c *CollectedResults) HasFailedResults() bool {
	return len(c.FailedResults) > 0
}

// GetSuccessRate 获取成功率
func (c *CollectedResults) GetSuccessRate() float64 {
	total := len(c.Results)
	if total == 0 {
		return 0.0
	}
	
	success := len(c.SuccessfulResults)
	return float64(success) / float64(total)
}

// 状态序列化实现

// JSONStateSerializer JSON状态序列化器
type JSONStateSerializer struct{}

// Serialize 序列化状态
func (j *JSONStateSerializer) Serialize(state *EnhancedState) ([]byte, error) {
	return json.Marshal(state)
}

// Deserialize 反序列化状态
func (j *JSONStateSerializer) Deserialize(data []byte) (*EnhancedState, error) {
	var state EnhancedState
	err := json.Unmarshal(data, &state)
	if err != nil {
		return nil, err
	}
	return &state, nil
}

// 枚举类型的字符串表示

// String 返回任务复杂度的字符串表示
func (t TaskComplexity) String() string {
	switch t {
	case TaskComplexitySimple:
		return "simple"
	case TaskComplexityModerate:
		return "moderate"
	case TaskComplexityComplex:
		return "complex"
	default:
		return "unknown"
	}
}

// String 返回行动类型的字符串表示
func (a ActionType) String() string {
	switch a {
	case ActionTypeDirectAnswer:
		return "direct_answer"
	case ActionTypeCreatePlan:
		return "create_plan"
	case ActionTypeExecuteStep:
		return "execute_step"
	case ActionTypeReflect:
		return "reflect"
	case ActionTypeUpdatePlan:
		return "update_plan"
	default:
		return "unknown"
	}
}

// String 返回步骤状态的字符串表示
func (s StepStatus) String() string {
	switch s {
	case StepStatusPending:
		return "pending"
	case StepStatusExecuting:
		return "executing"
	case StepStatusCompleted:
		return "completed"
	case StepStatusFailed:
		return "failed"
	case StepStatusSkipped:
		return "skipped"
	default:
		return "unknown"
	}
}

// String 返回执行状态的字符串表示
func (e ExecutionStatus) String() string {
	switch e {
	case ExecutionStatusSuccess:
		return "success"
	case ExecutionStatusFailed:
		return "failed"
	case ExecutionStatusPartial:
		return "partial"
	default:
		return "unknown"
	}
}