package multiagent

import (
	"testing"

	"github.com/cloudwego/eino/schema"
	"github.com/stretchr/testify/assert"
)

func TestIsIndependentTopicContextControl(t *testing.T) {
	// 创建测试状态
	state := &MultiAgentState{
		OriginalMessages: []*schema.Message{
			{Role: schema.User, Content: "历史问题1"},
			{Role: schema.Assistant, Content: "历史回答1"},
			{Role: schema.User, Content: "当前问题"},
		},
		ExecutionHistory: []*ExecutionRecord{
			{StepID: "step1", Action: ActionTypeExecute, Status: ExecutionStatusCompleted},
		},
		ConversationContext: &ConversationContext{
			IsIndependentTopic: false,
			UserIntent:         "测试意图",
			ContextSummary:     "测试摘要",
		},
	}

	t.Run("DirectAnswer_WithContext", func(t *testing.T) {
		// 测试非独立话题时包含上下文
		state.ConversationContext.IsIndependentTopic = false
		message := buildDirectAnswerPrompt(state)
		assert.NotNil(t, message)

		// 应该包含完整的上下文信息
		assert.True(t, contains(message[0].Content, "测试意图") && contains(message[0].Content, "测试摘要"), "应该包含用户意图和上下文摘要")
	})

	t.Run("DirectAnswer_WithoutContext", func(t *testing.T) {
		// 测试独立话题时不包含上下文
		state.ConversationContext.IsIndependentTopic = true
		message := buildDirectAnswerPrompt(state)
		assert.NotNil(t, message)

		// 应该只包含当前用户问题
		assert.True(t, contains(message[0].Content, "当前问题") && !contains(message[0].Content, "历史问题1"), "应该只包含当前用户问题，不包含历史对话")
	})

	t.Run("PlanCreation_WithContext", func(t *testing.T) {
		// 测试非独立话题时包含上下文
		state.ConversationContext.IsIndependentTopic = false
		config := &MultiAgentConfig{}
		message := buildPlanCreationPrompt(state, config)
		assert.NotNil(t, message)

		// 应该包含原始消息和执行历史
		assert.True(t, contains(message.Content, "历史问题1"), "应该包含原始消息历史")
		assert.True(t, contains(message.Content, "step1"), "应该包含执行历史")
	})

	t.Run("PlanCreation_WithoutContext", func(t *testing.T) {
		// 测试独立话题时不包含上下文
		state.ConversationContext.IsIndependentTopic = true
		config := &MultiAgentConfig{}
		message := buildPlanCreationPrompt(state, config)
		assert.NotNil(t, message)

		// 应该只包含当前用户问题，不包含历史
		assert.True(t, contains(message.Content, "当前问题"), "应该包含当前用户问题")
		assert.False(t, contains(message.Content, "step1"), "不应该包含执行历史")
	})
}

// 辅助函数：检查字符串是否包含子字符串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				containsInMiddle(s, substr))))
}

func containsInMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
