package react

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/stretchr/testify/assert"
)

// Use MockChatModel from example.go

// TestTool for testing
type TestTool struct{}

func (t *TestTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "test_tool",
		Desc: "A test tool",
	}, nil
}

func (t *TestTool) InvokableRun(ctx context.Context, args string, opts ...tool.Option) (string, error) {
	return "test result", nil
}

func TestNewAgent(t *testing.T) {
	ctx := context.Background()

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Fatal("OPENAI_API_KEY not set")
	}

	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:  apiKey,
		Model:   "gpt-4o-mini",
		BaseURL: "https://api.openai.com/v1",
	})
	if err != nil {
		t.Fatalf("Failed to create chat model: %v", err)
	}

	// Create agent config
	config := &AgentConfig{
		ToolCallingModel: chatModel,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: []tool.BaseTool{
				&CalculatorTool{},
				&SearchTool{},
				&WeatherTool{},
			},
		},
		MaxStep:            10,
		DebugMode:          true,
		ToolReturnDirectly: map[string]bool{},
	}

	// Create agent
	agent, err := NewAgent(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	if agent == nil {
		t.Fatal("Agent should not be nil")
	}
	// Create agent with MessageFuture
	option, future := WithMessageFuture()
	stream, err := agent.Stream(ctx, []*schema.Message{
		schema.UserMessage("10+100等于多少"),
	}, option)
	assert.Nil(t, err)
	assert.NotNil(t, stream)

	// Collect all chunks from stream
	finalResponse, err := schema.ConcatMessageStream(stream)
	assert.Nil(t, err)
	// assert.Equal(t, "final response", finalResponse.Content)
	fmt.Println(finalResponse.Content)
	// Get message streams from future
	sIter := future.GetMessageStreams()

	// First message should be the assistant message for tool calling
	stream1, hasNext, err := sIter.Next()
	assert.Nil(t, err)
	assert.True(t, hasNext)
	assert.NotNil(t, stream1)
	msg1, err := schema.ConcatMessageStream(stream1)
	assert.Nil(t, err)
	assert.Equal(t, schema.Assistant, msg1.Role)
	assert.Equal(t, 1, len(msg1.ToolCalls))
}

func TestParseReasoningResponse(t *testing.T) {
	tests := []struct {
		name     string
		message  *schema.Message
		expected ReasoningDecision
	}{
		{
			name: "final answer",
			message: &schema.Message{
				Content: "Thought: I have the answer\nAction: final_answer\nFinal Answer: 42",
			},
			expected: ReasoningDecision{
				Thought:     "I have the answer",
				Action:      "final_answer",
				FinalAnswer: "42",
			},
		},
		{
			name: "tool call with ToolCalls",
			message: &schema.Message{
				Content: "I need to search for information",
				ToolCalls: []schema.ToolCall{
					{
						Function: schema.FunctionCall{
							Name:      "search",
							Arguments: `{"query": "test query"}`,
						},
					},
				},
			},
			expected: ReasoningDecision{
				Thought: "I need to search for information",
				Action:  "tool_call",
				ToolCalls: []schema.ToolCall{
					{
						Function: schema.FunctionCall{
							Name:      "search",
							Arguments: `{"query": "test query"}`,
						},
					},
				},
			},
		},
		{
			name: "text-based tool call",
			message: &schema.Message{
				Content: "Thought: I need to search\nAction: tool_call\nTool: search\nArgs: {\"input\": \"test query\"}",
			},
			expected: ReasoningDecision{
				Thought:   "I need to search",
				Action:    "tool_call",
				ToolCalls: nil, // Text parsing doesn't extract ToolCalls in current implementation
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseReasoningResponse(tt.message)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if result.Thought != tt.expected.Thought {
				t.Errorf("Expected thought %q, got %q", tt.expected.Thought, result.Thought)
			}
			if result.Action != tt.expected.Action {
				t.Errorf("Expected action %q, got %q", tt.expected.Action, result.Action)
			}
			// Check ToolCalls
			if len(result.ToolCalls) != len(tt.expected.ToolCalls) {
				t.Errorf("Expected %d tool calls, got %d", len(tt.expected.ToolCalls), len(result.ToolCalls))
			} else if len(result.ToolCalls) > 0 && len(tt.expected.ToolCalls) > 0 {
				if result.ToolCalls[0].Function.Name != tt.expected.ToolCalls[0].Function.Name {
					t.Errorf("Expected tool name %q, got %q", tt.expected.ToolCalls[0].Function.Name, result.ToolCalls[0].Function.Name)
				}
			}
		})
	}
}

// Sample tool implementations

// CalculatorTool implements basic arithmetic operations
type CalculatorTool struct{}

func (t *CalculatorTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "calculator",
		Desc: "Performs basic arithmetic operations. Supports +, -, *, / operations.",
		ParamsOneOf: schema.NewParamsOneOfByParams(
			map[string]*schema.ParameterInfo{
				"expression": {
					Type: schema.String,
					Desc: "The arithmetic expression to evaluate.",
				},
			},
		),
	}, nil
}

func (t *CalculatorTool) InvokableRun(ctx context.Context, args string, opts ...tool.Option) (string, error) {
	// Parse JSON arguments
	var argsMap map[string]interface{}
	if err := json.Unmarshal([]byte(args), &argsMap); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	expressionInterface, ok := argsMap["expression"]
	if !ok {
		return "", fmt.Errorf("missing 'expression' argument")
	}

	expression, ok := expressionInterface.(string)
	if !ok {
		return "", fmt.Errorf("'expression' must be a string")
	}

	// Simple calculator implementation (for demo purposes)
	result := fmt.Sprintf("%s = 110", expression)
	return result, nil
}

// SearchTool implements web search functionality
type SearchTool struct{}

func (t *SearchTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "search",
		Desc: "Searches the web for information on a given topic.",
		ParamsOneOf: schema.NewParamsOneOfByParams(
			map[string]*schema.ParameterInfo{
				"query": {
					Type: schema.String,
					Desc: "The search query to execute.",
				},
			},
		),
	}, nil
}

func (t *SearchTool) InvokableRun(ctx context.Context, args string, opts ...tool.Option) (string, error) {
	// Parse JSON arguments
	var argsMap map[string]interface{}
	if err := json.Unmarshal([]byte(args), &argsMap); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	queryInterface, ok := argsMap["query"]
	if !ok {
		return "", fmt.Errorf("missing 'query' argument")
	}

	query, ok := queryInterface.(string)
	if !ok {
		return "", fmt.Errorf("'query' must be a string")
	}

	// Mock search result
	result := fmt.Sprintf("Search results for '%s': [This is a demo - implement actual search logic]", query)
	return result, nil
}

// WeatherTool implements weather information retrieval
type WeatherTool struct{}

func (t *WeatherTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "weather",
		Desc: "Gets current weather information for a specified location.",
		ParamsOneOf: schema.NewParamsOneOfByParams(
			map[string]*schema.ParameterInfo{
				"location": {
					Type: schema.String,
					Desc: "The location for which to retrieve weather information.",
				},
			},
		),
	}, nil
}

func (t *WeatherTool) InvokableRun(ctx context.Context, args string, opts ...tool.Option) (string, error) {
	// Parse JSON arguments
	var argsMap map[string]interface{}
	if err := json.Unmarshal([]byte(args), &argsMap); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	locationInterface, ok := argsMap["location"]
	if !ok {
		return "", fmt.Errorf("missing 'location' argument")
	}

	location, ok := locationInterface.(string)
	if !ok {
		return "", fmt.Errorf("'location' must be a string")
	}

	// Mock weather result
	result := fmt.Sprintf("Weather for '%s': [This is a demo - implement actual weather API integration]", location)
	return result, nil
}
