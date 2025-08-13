package multiagent

import (
	"encoding/json"
	"time"

	"github.com/cloudwego/eino/schema"
)

// MultiAgentState represents the complete state of the enhanced multi-agent system
type MultiAgentState struct {
	RoundNumber int       `json:"round_number"`
	StartTime   time.Time `json:"start_time"`

	// Conversation Context
	ConversationContext *ConversationContext `json:"conversation_context,omitempty"`
	OriginalMessages    []*schema.Message    `json:"original_messages"`

	// Task Planning
	CurrentPlan *TaskPlan   `json:"current_plan,omitempty"`
	PlanHistory []*TaskPlan `json:"plan_history,omitempty"`

	// Execution Status
	ExecutionStatus  ExecutionStatus    `json:"execution_status"`
	CurrentStep      string             `json:"current_step,omitempty"`
	ExecutionHistory []*ExecutionRecord `json:"execution_history,omitempty"`

	// Specialist Results
	SpecialistResults map[string]*StepResult `json:"specialist_results,omitempty"`

	// Collected Results
	CollectedResults []*schema.Message `json:"collected_results,omitempty"`
	FinalResult      *schema.Message   `json:"final_result,omitempty"`

	// Feedback and Reflection
	FeedbackHistory []map[string]any `json:"feedback_history,omitempty"`
	ReflectionCount int              `json:"reflection_count"`

	// Execution Control
	MaxRounds      int  `json:"max_rounds"`
	ShouldContinue bool `json:"should_continue"`
	IsCompleted    bool `json:"is_completed"`

	// Final Answer
	FinalAnswer *schema.Message `json:"final_answer,omitempty"`

	// Metadata
	Metadata map[string]any `json:"metadata,omitempty"`
}

// ToJSON serializes the state to JSON
func (es *MultiAgentState) ToJSON() ([]byte, error) {
	return json.Marshal(es)
}

// FromJSON deserializes the state from JSON
func (es *MultiAgentState) FromJSON(data []byte) error {
	return json.Unmarshal(data, es)
}

// Clone creates a deep copy of the state
func (es *MultiAgentState) Clone() (*MultiAgentState, error) {
	data, err := es.ToJSON()
	if err != nil {
		return nil, err
	}

	cloned := &MultiAgentState{}
	err = cloned.FromJSON(data)
	if err != nil {
		return nil, err
	}

	return cloned, nil
}

// 基础字段管理方法
func (es *MultiAgentState) SetRoundNumber(round int) {
	es.RoundNumber = round
}

func (es *MultiAgentState) IncrementRound() {
	es.RoundNumber++
}

func (es *MultiAgentState) SetStartTime(t time.Time) {
	es.StartTime = t
}

// 对话上下文管理方法
func (es *MultiAgentState) UpdateConversationContext(ctx *ConversationContext) {
	es.ConversationContext = ctx
}

func (es *MultiAgentState) SetOriginalMessages(messages []*schema.Message) {
	es.OriginalMessages = messages
}

// 计划管理方法
func (es *MultiAgentState) SetCurrentPlan(plan *TaskPlan) {
	es.CurrentPlan = plan
}

func (es *MultiAgentState) AddPlanToHistory(plan *TaskPlan) {
	if es.PlanHistory == nil {
		es.PlanHistory = make([]*TaskPlan, 0)
	}
	es.PlanHistory = append(es.PlanHistory, plan)
}

// 执行记录管理方法
func (es *MultiAgentState) SetExecutionStatus(status ExecutionStatus) {
	es.ExecutionStatus = status
}

func (es *MultiAgentState) SetCurrentStep(stepID string) {
	es.CurrentStep = stepID
}

func (es *MultiAgentState) AddExecutionRecord(record *ExecutionRecord) {
	if es.ExecutionHistory == nil {
		es.ExecutionHistory = make([]*ExecutionRecord, 0)
	}
	es.ExecutionHistory = append(es.ExecutionHistory, record)
}

// 专家结果管理方法
func (es *MultiAgentState) UpdateSpecialistResult(specialist string, result *StepResult) {
	if es.SpecialistResults == nil {
		es.SpecialistResults = make(map[string]*StepResult)
	}
	es.SpecialistResults[specialist] = result
}

func (es *MultiAgentState) ClearSpecialistResults() {
	es.SpecialistResults = make(map[string]*StepResult)
}

// 结果收集管理方法
func (es *MultiAgentState) AddCollectedResult(result *schema.Message) {
	if es.CollectedResults == nil {
		es.CollectedResults = make([]*schema.Message, 0)
	}
	es.CollectedResults = append(es.CollectedResults, result)
}

func (es *MultiAgentState) SetFinalResult(result *schema.Message) {
	es.FinalResult = result
}

// 反馈管理方法
func (es *MultiAgentState) AddFeedback(feedback map[string]any) {
	if es.FeedbackHistory == nil {
		es.FeedbackHistory = make([]map[string]any, 0)
	}
	es.FeedbackHistory = append(es.FeedbackHistory, feedback)
}

func (es *MultiAgentState) IncrementReflection() {
	es.ReflectionCount++
}

// 元数据管理方法
func (es *MultiAgentState) SetMaxRounds(max int) {
	es.MaxRounds = max
}

func (es *MultiAgentState) SetShouldContinue(should bool) {
	es.ShouldContinue = should
}

func (es *MultiAgentState) SetCompleted(completed bool) {
	es.IsCompleted = completed
}

func (es *MultiAgentState) SetFinalAnswer(answer *schema.Message) {
	es.FinalAnswer = answer
}

func (es *MultiAgentState) SetMetadata(key string, value any) {
	if es.Metadata == nil {
		es.Metadata = make(map[string]any)
	}
	es.Metadata[key] = value
}

func (es *MultiAgentState) GetMetadata(key string) (any, bool) {
	if es.Metadata == nil {
		return nil, false
	}
	value, exists := es.Metadata[key]
	return value, exists
}
