package multiagent

import (
	"fmt"

	"github.com/cloudwego/eino/schema"
)

func buildDirectAnswerPrompt(state *MultiAgentState) *schema.Message {
	prompt := fmt.Sprintf(`Provide a direct answer to the user's request.

User Intent: %s
Context: %s

Please provide a clear, helpful response.`,
		state.ConversationContext.UserIntent,
		state.ConversationContext.ContextSummary,
	)

	return &schema.Message{
		Role:    schema.User,
		Content: prompt,
	}
}

func buildPlanCreationPrompt(state *MultiAgentState, specialists []*Specialist) *schema.Message {
	specialistList := ""
	for _, specialist := range specialists {
		specialistList += fmt.Sprintf("- %s: %s\n", specialist.Name, specialist.IntendedUse)
	}

	prompt := fmt.Sprintf(`Create a detailed execution plan for the following task.

Task Context:
- User Intent: %s
- Complexity: %s
- Key Topics: %v

Available Specialists:
%s

IMPORTANT: You MUST respond with ONLY a valid JSON object. Do not include any explanations, comments, or additional text before or after the JSON. Your response should start with { and end with }.

Create a plan with the following JSON structure:
{
  "id": "unique_plan_id",
  "name": "Plan Name",
  "description": "Plan Description",
  "steps": [
    {
      "id": "step_1",
      "name": "Step Name",
      "description": "Step Description",
      "assigned_specialist": "specialist_name",
      "priority": 1,
      "dependencies": [],
      "parameters": {}
    }
  ]
}

Notice:
- The plan must be executable.
- Each step must be assigned to a specialist, and the specialist must be available.

Remember: Output ONLY the JSON object, no other text.`,
		state.ConversationContext.UserIntent,
		state.ConversationContext.Complexity.String(),
		state.ConversationContext.KeyTopics,
		specialistList,
	)

	return &schema.Message{
		Role:    schema.User,
		Content: prompt,
	}
}

func buildSpecialistPrompt(specialist *Specialist, step *PlanStep, state *MultiAgentState) *schema.Message {
	prompt := fmt.Sprintf(`You are a %s specialist. Execute the following step:

Step: %s
Description: %s

Context:
- User Intent: %s
- Overall Plan: %s

Parameters: %v

Please complete this step and provide your result.`,
		specialist.Name,
		step.Name,
		step.Description,
		state.ConversationContext.UserIntent,
		state.CurrentPlan.Description,
		step.Parameters,
	)

	return &schema.Message{
		Role:    schema.User,
		Content: prompt,
	}
}

func buildFeedbackPrompt(state *MultiAgentState) []*schema.Message {
	prompt := `Analyze the execution results and provide comprehensive feedback.

Original User Intent: ` + state.ConversationContext.UserIntent + `

Current Plan:
`
	if state.CurrentPlan != nil {
		prompt += fmt.Sprintf("Plan: %s\nDescription: %s\n", state.CurrentPlan.Name, state.CurrentPlan.Description)
		for _, step := range state.CurrentPlan.Steps {
			prompt += fmt.Sprintf("- Step %s: %s (Status: %s)\n", step.ID, step.Name, step.Status.String())
		}
	}

	prompt += `\nExecution Results:\n`
	for _, result := range state.CollectedResults {
		prompt += result.Content + "\n\n"
	}

	prompt += fmt.Sprintf(`\nExecution History: %d records
Round: %d/%d

IMPORTANT: You MUST respond with ONLY a valid JSON object. Do not include any explanations, comments, or additional text before or after the JSON. Your response should start with { and end with }.

Provide feedback in JSON format:
{
  "execution_completed": false,
  "overall_quality": 0.8,
  "plan_needs_update": false,
  "issues": ["issue1", "issue2"],
  "suggestions": ["suggestion1", "suggestion2"],
  "confidence": 0.9,
  "next_action_reason": "Explanation for the recommended next action"
}

Decision criteria:
- execution_completed: true if the task is fully completed and satisfactory
- plan_needs_update: true if the current plan needs modification to better achieve the goal
- If execution_completed=false and plan_needs_update=false, continue with current plan

Remember: Output ONLY the JSON object, no other text.`, len(state.ExecutionHistory), state.RoundNumber, state.MaxRounds)

	return []*schema.Message{{
		Role:    schema.User,
		Content: prompt,
	}}
}

func buildPlanUpdatePrompt(state *MultiAgentState) []*schema.Message {
	prompt := `Update the current plan based on feedback and execution results.

Original User Intent: ` + state.ConversationContext.UserIntent + `

Current Plan:
`
	if state.CurrentPlan != nil {
		prompt += fmt.Sprintf("Plan: %s\nDescription: %s\n", state.CurrentPlan.Name, state.CurrentPlan.Description)
		for _, step := range state.CurrentPlan.Steps {
			prompt += fmt.Sprintf("- Step %s: %s (Status: %s, Priority: %d)\n", step.ID, step.Name, step.Status.String(), step.Priority)
		}
	}

	prompt += `\nExecution Results:\n`
	for _, result := range state.CollectedResults {
		prompt += result.Content + "\n\n"
	}

	// Add feedback information
	if len(state.FeedbackHistory) > 0 {
		latestFeedback := state.FeedbackHistory[len(state.FeedbackHistory)-1]
		if content, ok := latestFeedback["content"].(string); ok {
			prompt += "\nLatest Feedback:\n" + content + "\n\n"
		}
	}

	// Add feedback decision context
	if reason, exists := state.GetMetadata("feedback_next_action_reason"); exists {
		if reasonStr, ok := reason.(string); ok {
			prompt += "Reason for Plan Update: " + reasonStr + "\n\n"
		}
	}

	prompt += fmt.Sprintf(`Round: %d/%d

Provide updated plan in JSON format:
{
  "name": "Updated Plan Name",
  "description": "Detailed plan description",
  "update_reason": "Why this update is needed",
  "steps": [
    {
      "id": "step1",
      "name": "Step name",
      "description": "Detailed step description",
      "assigned_specialist": "specialist_name",
      "priority": 1,
      "dependencies": ["prerequisite_step_ids"],
      "parameters": {"key": "value"}
    }
  ]
}

Guidelines:
- Address the issues identified in the feedback
- Maintain continuity with completed steps
- Optimize step order and dependencies
- Assign appropriate specialists to each step
- Set realistic priorities based on importance and dependencies`, state.RoundNumber, state.MaxRounds)

	return []*schema.Message{{
		Role:    schema.User,
		Content: prompt,
	}}
}

func buildFinalAnswerPrompt(state *MultiAgentState) *schema.Message {
	// Build prompt for final answer generation
	content := "Please provide a comprehensive final answer based on the following analysis and execution results:\n\n"

	// Add original question
	if len(state.OriginalMessages) > 0 {
		content += "Original Question: " + state.OriginalMessages[0].Content + "\n\n"
	}

	// Add execution plan if available
	if state.CurrentPlan != nil {
		content += "Execution Plan:\n"
		for i, step := range state.CurrentPlan.Steps {
			content += fmt.Sprintf("%d. %s\n", i+1, step.Description)
		}
		content += "\n"
	}

	// Add collected results
	if len(state.CollectedResults) > 0 {
		content += "Analysis Results:\n"
		for i, result := range state.CollectedResults {
			content += fmt.Sprintf("Result %d: %s\n", i+1, result.Content)
		}
		content += "\n"
	}

	content += "Please synthesize all the above information into a clear, comprehensive, and well-structured final answer."

	return &schema.Message{
		Role:    schema.User,
		Content: content,
	}
}
