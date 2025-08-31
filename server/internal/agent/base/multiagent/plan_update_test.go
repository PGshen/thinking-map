package multiagent

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlanUpdateHandler_IncrementalUpdate(t *testing.T) {
	// Create a test plan with some steps
	initialPlan := &TaskPlan{
		ID:          "test-plan-1",
		Version:     1,
		Name:        "Test Plan",
		Description: "A test plan for incremental updates",
		Status:      ExecutionStatusExecuting,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Steps: []*PlanStep{
			{
				ID:                 "step-1",
				Name:               "Initial Step 1",
				Description:        "First step",
				AssignedSpecialist: "analyst",
				Priority:           1,
				Status:             StepStatusCompleted,
				Dependencies:       []string{},
				Parameters:         map[string]any{"param1": "value1"},
			},
			{
				ID:                 "step-2",
				Name:               "Initial Step 2",
				Description:        "Second step",
				AssignedSpecialist: "researcher",
				Priority:           2,
				Status:             StepStatusPending,
				Dependencies:       []string{"step-1"},
				Parameters:         map[string]any{"param2": "value2"},
			},
		},
		PlanUpdate: new(PlanUpdate),
		Metadata:   map[string]any{},
	}

	// Create handler
	handler := &PlanUpdateHandler{}

	// Test clonePlan
	t.Run("ClonePlan", func(t *testing.T) {
		cloned := handler.clonePlan(initialPlan)

		// Verify the clone is a deep copy
		assert.Equal(t, initialPlan.ID, cloned.ID)
		assert.Equal(t, initialPlan.Version, cloned.Version)
		assert.Equal(t, len(initialPlan.Steps), len(cloned.Steps))

		// Modify original and ensure clone is not affected
		initialPlan.Steps[0].Name = "Modified Name"
		assert.NotEqual(t, initialPlan.Steps[0].Name, cloned.Steps[0].Name)
		assert.Equal(t, "Initial Step 1", cloned.Steps[0].Name)
	})

	// Test addStep operation
	t.Run("AddStep", func(t *testing.T) {
		plan := handler.clonePlan(initialPlan)

		addOp := &OperationData{
			Type:   "add",
			StepID: "step-3",
			StepData: &StepData{
				ID:                 "step-3",
				Name:               "New Step 3",
				Description:        "A new step added incrementally",
				AssignedSpecialist: "writer",
				Priority:           3,
				Dependencies:       []string{"step-2"},
				Parameters:         map[string]any{"param3": "value3"},
			},
			Position: "", // Append at end
			Reason:   "Adding new step based on feedback",
		}

		err := handler.addStep(plan, addOp)
		require.NoError(t, err)

		assert.Equal(t, 3, len(plan.Steps))
		assert.Equal(t, "step-3", plan.Steps[2].ID)
		assert.Equal(t, "New Step 3", plan.Steps[2].Name)
		assert.Equal(t, StepStatusPending, plan.Steps[2].Status)
	})

	// Test modifyStep operation
	t.Run("ModifyStep", func(t *testing.T) {
		plan := handler.clonePlan(initialPlan)

		modifyOp := &OperationData{
			Type:   "modify",
			StepID: "step-2",
			StepData: &StepData{
				Name:        "Modified Step 2",
				Description: "Updated description for step 2",
				Priority:    5,
			},
			Reason: "Updating step based on new requirements",
		}

		err := handler.modifyStep(plan, modifyOp)
		require.NoError(t, err)

		assert.Equal(t, "Modified Step 2", plan.Steps[1].Name)
		assert.Equal(t, "Updated description for step 2", plan.Steps[1].Description)
		assert.Equal(t, 5, plan.Steps[1].Priority)
		// Original values should be preserved if not specified
		assert.Equal(t, "researcher", plan.Steps[1].AssignedSpecialist)
	})

	// Test that completed steps cannot be modified
	t.Run("CannotModifyCompletedStep", func(t *testing.T) {
		plan := handler.clonePlan(initialPlan)

		modifyOp := &OperationData{
			Type:   "modify",
			StepID: "step-1", // This step is completed
			StepData: &StepData{
				Name: "Should Not Change",
			},
		}

		err := handler.modifyStep(plan, modifyOp)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot modify completed step")
	})

	// Test removeStep operation
	t.Run("RemoveStep", func(t *testing.T) {
		plan := handler.clonePlan(initialPlan)

		removeOp := &OperationData{
			Type:   "remove",
			StepID: "step-2",
			Reason: "Step no longer needed",
		}

		err := handler.removeStep(plan, removeOp)
		require.NoError(t, err)

		assert.Equal(t, 1, len(plan.Steps))
		assert.Equal(t, "step-1", plan.Steps[0].ID)
	})

	// Test findStepIndex
	t.Run("FindStepIndex", func(t *testing.T) {
		plan := handler.clonePlan(initialPlan)

		index := handler.findStepIndex(plan, "step-2")
		assert.Equal(t, 1, index)

		index = handler.findStepIndex(plan, "non-existent")
		assert.Equal(t, -1, index)
	})
}

func TestSelectiveClearSpecialistResults(t *testing.T) {
	// Create test state with specialist results
	state := &MultiAgentState{
		SpecialistResults: map[string]*StepResult{
			"step-1": {Success: true, Confidence: 0.9},
			"step-2": {Success: false, Confidence: 0.5},
			"step-3": {Success: true, Confidence: 0.8},
		},
	}

	handler := &PlanUpdateHandler{}

	// Test that only affected steps are cleared
	operations := []OperationData{
		{Type: "modify", StepID: "step-2"},
		{Type: "add", StepID: "step-4"}, // New step, no existing results
	}

	handler.selectiveClearSpecialistResults(state, operations)

	// step-1 and step-3 should remain, step-2 should be cleared
	assert.Contains(t, state.SpecialistResults, "step-1")
	assert.NotContains(t, state.SpecialistResults, "step-2")
	assert.Contains(t, state.SpecialistResults, "step-3")
	assert.Equal(t, 2, len(state.SpecialistResults))
}
