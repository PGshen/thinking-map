package multiagent

import (
	"fmt"
	"strings"

	"github.com/cloudwego/eino/schema"
)

func buildConversationAnalysisPrompt(messages []*schema.Message) *schema.Message {
	prompt := `Analyze the following conversation and extract key information:

Conversation:
`
	for _, msg := range messages {
		prompt += fmt.Sprintf("%s: %s\n", msg.Role, msg.Content)
	}

	prompt += `
IMPORTANT: You MUST respond with ONLY a valid JSON object. Do not include any explanations, comments, or additional text before or after the JSON. Your response should start with { and end with }.

Please analyze and provide the following information in JSON format:
{
  "userIntent": "Brief description of what the user wants to achieve",
  "keyTopics": ["topic1", "topic2", "topic3"],
  "contextSummary": "Summary of the conversation context",
  "complexity": "simple|moderate|complex|very_complex",
  "metadata": {}
}

Remember: 
- Output ONLY the JSON object, no other text.
- Reply in the same language as the user's question (Chinese for Chinese questions, English for English questions)
`

	return &schema.Message{
		Role:    schema.User,
		Content: prompt,
	}
}

func buildDirectAnswerPrompt(state *MultiAgentState) *schema.Message {
	prompt := fmt.Sprintf(`Provide a direct answer to the user's request.

User Intent: %s
Context: %s

Please provide a clear, helpful response.

Notice:
- Reply in the same language as the user's question (Chinese for Chinese questions, English for English questions)
`,
		state.ConversationContext.UserIntent,
		state.ConversationContext.ContextSummary,
	)

	return &schema.Message{
		Role:    schema.User,
		Content: prompt,
	}
}

func buildPlanCreationPrompt(state *MultiAgentState, config *MultiAgentConfig) *schema.Message {
	specialistList := ""
	for _, specialist := range config.Specialists {
		specialistList += fmt.Sprintf("- %s: %s\n", specialist.Name, specialist.IntendedUse)
	}

	planningPrompt := config.Host.Planning.PlanningPrompt

	prompt := fmt.Sprintf(`Create a detailed execution plan for the following task.

Task Context:
- User Intent: %s
- Complexity: %s
- Key Topics: %s

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
      "assignedSpecialist": "specialist_name",
      "priority": 1,
      "dependencies": [],
      "parameters": {}
    }
  ]
}

Notice:
- The plan must be executable.
- Control the number of steps in the plan.
- Each step must be assigned to a specialist, and the specialist must be available.
- Reply in the same language as the user's question (Chinese for Chinese questions, English for English questions)
- Must strictly follow JSON format for replies, do not add any extra text`,
		state.ConversationContext.UserIntent,
		state.ConversationContext.Complexity,
		strings.Join(state.ConversationContext.KeyTopics, ", "),
		specialistList,
	)

	return &schema.Message{
		Role:    schema.User,
		Content: planningPrompt + "\n" + prompt,
	}
}

func buildSpecialistPrompt(specialist *Specialist, step *PlanStep, state *MultiAgentState) []*schema.Message {
	messages := []*schema.Message{}
	if specialist.SystemPrompt != "" {
		messages = append(messages, &schema.Message{
			Role:    schema.System,
			Content: specialist.SystemPrompt,
		})
	} else {
		messages = append(messages, &schema.Message{
			Role:    schema.System,
			Content: fmt.Sprintf(`You are a %s specialist, intended to %s.`, specialist.Name, specialist.IntendedUse),
		})
	}
	// Build specialist prompt
	prompt := fmt.Sprintf(`Execute the following step:

Step: %s
Description: %s

Context:
- User Intent: %s
- Overall Plan: %s

Please complete this step and provide your result.

Notice:
- Reply in the same language as the user's question (Chinese for Chinese questions, English for English questions)
`,
		step.Name,
		step.Description,
		state.ConversationContext.UserIntent,
		state.CurrentPlan.Description,
	)
	messages = append(messages, &schema.Message{
		Role:    schema.User,
		Content: prompt,
	})

	return messages
}

func buildFeedbackPrompt(state *MultiAgentState) []*schema.Message {
	prompt := `Analyze the execution results and provide comprehensive feedback.

Original User Intent: ` + state.ConversationContext.UserIntent + `

`
	if state.CurrentPlan != nil {
		prompt += fmt.Sprintf("Plan: %s\nDescription: %s\n", state.CurrentPlan.Name, state.CurrentPlan.Description)
		for _, step := range state.CurrentPlan.Steps {
			prompt += fmt.Sprintf("- Step %s: %s (Status: %s)\n", step.ID, step.Name, step.Status)
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

Notice:
- The plan must be executable.
- Each step must be assigned to a specialist, and the specialist must be available.
- Reply in the same language as the user's question (Chinese for Chinese questions, English for English questions)
- Must strictly follow JSON format for replies, do not add any extra text
`, len(state.ExecutionHistory), state.RoundNumber, state.MaxRounds)

	return []*schema.Message{{
		Role:    schema.User,
		Content: prompt,
	}}
}

func buildPlanUpdatePrompt(state *MultiAgentState) []*schema.Message {
	prompt := `Analyze the current plan and provide incremental updates based on feedback and execution results.

Original User Intent: ` + state.ConversationContext.UserIntent + `

Current Plan:
`
	if state.CurrentPlan != nil {
		prompt += fmt.Sprintf("Plan: %s\nDescription: %s\n", state.CurrentPlan.Name, state.CurrentPlan.Description)
		for _, step := range state.CurrentPlan.Steps {
			prompt += fmt.Sprintf("- Step %s: %s (Status: %s, Priority: %d)\n", step.ID, step.Name, step.Status, step.Priority)
			if step.AssignedSpecialist != "" {
				prompt += fmt.Sprintf("  Assigned Specialist: %s\n", step.AssignedSpecialist)
			}
			if step.Description != "" {
				prompt += fmt.Sprintf("  Description: %s\n", step.Description)
			}
			if len(step.Parameters) > 0 {
				prompt += fmt.Sprintf("  Parameters: %v\n", step.Parameters)
			}
			if len(step.Dependencies) > 0 {
				prompt += fmt.Sprintf("  Dependencies: %v\n", step.Dependencies)
			}
			prompt += "\n"
		}
	}

	prompt += `\nExecution Results:\n`
	for _, result := range state.CollectedResults {
		prompt += result.Content + "\n\n"
	}

	// Add execution history for context
	if len(state.ExecutionHistory) > 0 {
		prompt += "\nExecution History:\n"
		for _, record := range state.ExecutionHistory {
			prompt += fmt.Sprintf("- Step %s: %s (Status: %s)\n", record.StepID, record.Action, record.Status)
		}
		prompt += "\n"
	}

	// Add feedback information
	if len(state.FeedbackHistory) > 0 {
		latestFeedback := state.FeedbackHistory[len(state.FeedbackHistory)-1]
		prompt += "\nLatest Feedback:\n"
		prompt += fmt.Sprintf("Execution Completed: %v\n", latestFeedback.ExecutionCompleted)
		prompt += fmt.Sprintf("Overall Quality: %.2f\n", latestFeedback.OverallQuality)
		prompt += fmt.Sprintf("Plan Needs Update: %v\n", latestFeedback.PlanNeedsUpdate)
		prompt += fmt.Sprintf("Confidence: %.2f\n", latestFeedback.Confidence)
		if len(latestFeedback.Issues) > 0 {
			prompt += fmt.Sprintf("Issues: %v\n", latestFeedback.Issues)
		}
		if len(latestFeedback.Suggestions) > 0 {
			prompt += fmt.Sprintf("Suggestions: %v\n", latestFeedback.Suggestions)
		}
		prompt += "\n"
	}

	// Add feedback decision context
	if len(state.FeedbackHistory) > 0 {
		latestFeedback := state.FeedbackHistory[len(state.FeedbackHistory)-1]
		if latestFeedback.NextActionReason != "" {
			prompt += "Reason for Plan Update: " + latestFeedback.NextActionReason + "\n\n"
		}
	}

	prompt += fmt.Sprintf(`Round: %d/%d

IMPORTANT: You MUST respond with ONLY a valid JSON object. Do not include any explanations, comments, or additional text before or after the JSON. Your response should start with { and end with }.

Provide incremental plan updates in JSON format:
{
  "update_reason": "Why this update is needed",
  "operations": [
    {
      "type": "add|modify|remove|reorder",
      "stepID": "target_stepID",
      "step_data": {
        "id": "new_stepID",
        "name": "Step name",
        "description": "Detailed step description",
        "assignedSpecialist": "specialist_name",
        "priority": 1,
        "dependencies": ["prerequisite_stepIDs"],
        "parameters": {"key": "value"}
      },
      "position": "before|after|index_number",
      "reason": "Why this operation is needed"
    }
  ],
  "plan_metadata": {
    "name": "Updated plan name (if changed)",
    "description": "Updated plan description (if changed)"
  }
}

Operation Types:
- "add": Add a new step. Use step_data and position fields.
- "modify": Modify an existing step. Use stepID and step_data fields. Can not modify completed steps.
- "remove": Remove a step. Use stepID field only.
- "reorder": Change step order. Use stepID and position fields.

Position Values:
- "before": Insert before the specified stepID
- "after": Insert after the specified stepID  
- number: Insert at specific index (0-based)

Guidelines:
- PRESERVE completed steps (Status: completed) - do not modify or remove them
- Only suggest changes that address the feedback issues
- Maintain logical dependencies between steps
- Consider the current execution state when making changes
- Be conservative - only make necessary changes

Notice:
- Reply in the same language as the user's question (Chinese for Chinese questions, English for English questions)
- Must strictly follow JSON format for replies, do not add any extra text
`, state.RoundNumber, state.MaxRounds)

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

	content += `Please synthesize all the above information into a clear, comprehensive, and well-structured final answer.

Notice:
- Reply in the same language as the user's question (Chinese for Chinese questions, English for English questions)
`

	return &schema.Message{
		Role:    schema.User,
		Content: content,
	}
}
