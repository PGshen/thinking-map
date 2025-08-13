package react

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/schema"
)

type InitHandler struct {
	config ReactAgentConfig
}

func NewInitHandler(config ReactAgentConfig) *InitHandler {
	return &InitHandler{
		config: config,
	}
}

func (h *InitHandler) PreHandler(ctx context.Context, input []*schema.Message, state *AgentState) ([]*schema.Message, error) {
	// Initialize state
	state.Messages = append(state.Messages, input...)
	state.Iteration = 0
	state.Completed = false
	state.FinalAnswer = ""
	state.ReturnDirectlyToolCallID = ""
	return input, nil
}

type ReasoningHandler struct {
	config ReactAgentConfig
}

func NewReasoningHandler(config ReactAgentConfig) *ReasoningHandler {
	return &ReasoningHandler{
		config: config,
	}
}

func (h *ReasoningHandler) PreHandler(ctx context.Context, input []*schema.Message, state *AgentState) ([]*schema.Message, error) {
	state.Iteration++
	// Build system prompt
	systemPrompt := buildReasoningSystemPrompt()

	// Prepare messages with system prompt
	messages := []*schema.Message{
		{Role: schema.System, Content: systemPrompt},
	}
	messages = append(messages, state.Messages...)
	// for i, message := range messages {
	// 	fmt.Printf("ReasoningHandler: %d %s %s\n", i, message.Role, message.Content)
	// }

	return messages, nil
}

func (h *ReasoningHandler) PostHandler(ctx context.Context, output *schema.Message, state *AgentState) (*schema.Message, error) {
	// Parse reasoning response
	reasoning, err := parseReasoningResponse(output)
	if err != nil {
		return nil, fmt.Errorf("failed to parse reasoning response: %w", err)
	}

	// Add to reasoning history
	state.ReasoningHistory = append(state.ReasoningHistory, *reasoning)

	// Add output to messages
	state.Messages = append(state.Messages, output)

	return output, nil
}

type ToolHandler struct {
	config ReactAgentConfig
}

func NewToolHandler(config ReactAgentConfig) *ToolHandler {
	return &ToolHandler{
		config: config,
	}
}

func (h *ToolHandler) PostHandler(ctx context.Context, output []*schema.Message, state *AgentState) ([]*schema.Message, error) {
	// Update state with tool result
	state.Messages = append(state.Messages, output...)

	// Check if any tool should return directly
	if h.config.ToolReturnDirectly != nil {
		for _, msg := range output {
			if msg.ToolCallID != "" {
				// Find the corresponding tool call to get tool name
				for _, reasoning := range state.ReasoningHistory {
					for _, toolCall := range reasoning.ToolCalls {
						if toolCall.ID == msg.ToolCallID {
							if shouldReturn, exists := h.config.ToolReturnDirectly[toolCall.Function.Name]; exists && shouldReturn {
								state.ReturnDirectlyToolCallID = msg.ToolCallID
								break
							}
						}
					}
				}
			}
		}
	}

	return output, nil
}

type CompleteHandler struct {
	config ReactAgentConfig
}

func NewCompleteHandler(config ReactAgentConfig) *CompleteHandler {
	return &CompleteHandler{
		config: config,
	}
}

func (h *CompleteHandler) PostHandler(ctx context.Context, output *schema.Message, state *AgentState) (*schema.Message, error) {
	// Create final message based on state
	var finalMessage *schema.Message

	if state.FinalAnswer != "" {
		finalMessage = &schema.Message{
			Role:    schema.Assistant,
			Content: state.FinalAnswer,
		}
	} else if len(state.Messages) > 0 {
		finalMessage = state.Messages[len(state.Messages)-1]
	} else {
		finalMessage = &schema.Message{
			Role:    schema.Assistant,
			Content: "I apologize, but I was unable to provide a response.",
		}
	}

	return finalMessage, nil
}
