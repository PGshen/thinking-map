package multiagent

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFeedbackStructUsage(t *testing.T) {
	// Create a new state
	state := &MultiAgentState{
		RoundNumber:     1,
		StartTime:       time.Now(),
		ExecutionStatus: ExecutionStatusStarted,
		MaxRounds:       5,
		ShouldContinue:  true,
		IsCompleted:     false,
	}

	// Create feedback using the Feedback struct
	feedback := &Feedback{
		ExecutionCompleted: false,
		OverallQuality:     0.8,
		PlanNeedsUpdate:    true,
		Issues:             []string{"issue1", "issue2"},
		Suggestions:        []string{"suggestion1", "suggestion2"},
		Confidence:         0.9,
		NextActionReason:   "Need to update plan based on feedback",
	}

	// Add feedback to state
	state.AddFeedback(feedback)

	// Verify feedback was added correctly
	assert.Len(t, state.FeedbackHistory, 1)
	assert.Equal(t, feedback, state.FeedbackHistory[0])

	// Verify feedback fields
	latestFeedback := state.FeedbackHistory[0]
	assert.False(t, latestFeedback.ExecutionCompleted)
	assert.Equal(t, 0.8, latestFeedback.OverallQuality)
	assert.True(t, latestFeedback.PlanNeedsUpdate)
	assert.Equal(t, []string{"issue1", "issue2"}, latestFeedback.Issues)
	assert.Equal(t, []string{"suggestion1", "suggestion2"}, latestFeedback.Suggestions)
	assert.Equal(t, 0.9, latestFeedback.Confidence)
	assert.Equal(t, "Need to update plan based on feedback", latestFeedback.NextActionReason)

	// Add another feedback
	feedback2 := &Feedback{
		ExecutionCompleted: true,
		OverallQuality:     0.95,
		PlanNeedsUpdate:    false,
		Issues:             []string{},
		Suggestions:        []string{},
		Confidence:         0.98,
		NextActionReason:   "Task completed successfully",
	}

	state.AddFeedback(feedback2)

	// Verify both feedbacks are stored
	assert.Len(t, state.FeedbackHistory, 2)
	assert.Equal(t, feedback, state.FeedbackHistory[0])
	assert.Equal(t, feedback2, state.FeedbackHistory[1])
}

func TestPlanUpdatePromptGeneration(t *testing.T) {
	// Create a state with conversation context
	state := &MultiAgentState{
		RoundNumber:     2,
		StartTime:       time.Now(),
		ExecutionStatus: ExecutionStatusExecuting,
		MaxRounds:       5,
		ShouldContinue:  true,
		IsCompleted:     false,
		ConversationContext: &ConversationContext{
			UserIntent: "Test user intent",
		},
	}

	// Add some feedback
	feedback := &Feedback{
		ExecutionCompleted: false,
		OverallQuality:     0.7,
		PlanNeedsUpdate:    true,
		Issues:             []string{"Performance issue"},
		Suggestions:        []string{"Optimize algorithm"},
		Confidence:         0.8,
		NextActionReason:   "Algorithm needs optimization",
	}

	state.AddFeedback(feedback)

	// Test plan update prompt generation (this function includes feedback history)
	updateMessages := buildPlanUpdatePrompt(state)
	assert.Len(t, updateMessages, 1)
	assert.Contains(t, updateMessages[0].Content, "Test user intent")
	assert.Contains(t, updateMessages[0].Content, "Execution Completed: false")
	assert.Contains(t, updateMessages[0].Content, "Overall Quality: 0.70")
	assert.Contains(t, updateMessages[0].Content, "Plan Needs Update: true")
	assert.Contains(t, updateMessages[0].Content, "Confidence: 0.80")
	assert.Contains(t, updateMessages[0].Content, "Issues: [Performance issue]")
	assert.Contains(t, updateMessages[0].Content, "Suggestions: [Optimize algorithm]")
	assert.Contains(t, updateMessages[0].Content, "Reason for Plan Update: Algorithm needs optimization")

	// Test feedback prompt generation (this function is for generating feedback analysis prompts)
	messages := buildFeedbackPrompt(state)
	assert.Len(t, messages, 1)
	assert.Contains(t, messages[0].Content, "Test user intent")
	assert.Contains(t, messages[0].Content, "Analyze the execution results")
}