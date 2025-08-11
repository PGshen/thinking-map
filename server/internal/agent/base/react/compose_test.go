package react

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/PGshen/thinking-map/server/internal/agent/base"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	ub "github.com/cloudwego/eino/utils/callbacks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test tools for integration testing

// CalculatorTool implements basic arithmetic operations
type CalculatorTool struct{}

func (t *CalculatorTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "calculator",
		Desc: "Performs basic arithmetic calculations like addition, subtraction, multiplication, and division.",
		ParamsOneOf: schema.NewParamsOneOfByParams(
			map[string]*schema.ParameterInfo{
				"expression": {
					Type: schema.String,
					Desc: "The mathematical expression to evaluate (e.g., '10+5', '20*3', '100/4').",
				},
			},
		),
	}, nil
}

func (t *CalculatorTool) InvokableRun(ctx context.Context, args string, opts ...tool.Option) (string, error) {
	var params struct {
		Expression string `json:"expression"`
	}

	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Simple calculator implementation for testing
	switch params.Expression {
	case "10+5":
		return "15", nil
	case "10+100":
		return "110", nil
	case "20*3":
		return "60", nil
	case "100/4":
		return "25", nil
	default:
		return fmt.Sprintf("计算结果: %s = [模拟计算结果]", params.Expression), nil
	}
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
	var params struct {
		Query string `json:"query"`
	}

	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Mock search results for testing
	return fmt.Sprintf("搜索结果: 关于'%s'的相关信息已找到。[模拟搜索结果]", params.Query), nil
}

// WeatherTool implements weather information functionality
type WeatherTool struct{}

func (t *WeatherTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "weather",
		Desc: "Gets current weather information for a specified location.",
		ParamsOneOf: schema.NewParamsOneOfByParams(
			map[string]*schema.ParameterInfo{
				"location": {
					Type: schema.String,
					Desc: "The location to get weather information for.",
				},
			},
		),
	}, nil
}

func (t *WeatherTool) InvokableRun(ctx context.Context, args string, opts ...tool.Option) (string, error) {
	var params struct {
		Location string `json:"location"`
	}

	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Mock weather data for testing
	return fmt.Sprintf("%s的天气: 晴朗，温度22-28°C，湿度65%%，东南风3级。", params.Location), nil
}

// Test callback handler for monitoring
var testCallbackHandler = ub.NewHandlerHelper().ChatModel(&ub.ModelCallbackHandler{
	OnStart: func(ctx context.Context, runInfo *callbacks.RunInfo, input *model.CallbackInput) context.Context {
		inputContent := ""
		for _, msg := range input.Messages {
			inputContent += msg.Content + "\n"
		}
		fmt.Printf("\n[TEST] Model Start: %s\n%s\n", runInfo.Name, inputContent)
		return ctx
	},
	OnEnd: func(ctx context.Context, runInfo *callbacks.RunInfo, output *model.CallbackOutput) context.Context {
		fmt.Printf("\n[TEST] Model End: %s\n%s\n", runInfo.Name, output.Message.Content)
		return ctx
	},
	OnEndWithStreamOutput: func(ctx context.Context, runInfo *callbacks.RunInfo, output *schema.StreamReader[*model.CallbackOutput]) context.Context {
		fmt.Printf("\n[TEST] Model Stream End: %s\n", runInfo.Name)
		for {
			next, err := output.Recv()
			if err != nil {
				break
			}
			fmt.Printf("%s", next.Message.Content)
		}
		fmt.Printf("\n")
		return ctx
	},
}).ToolsNode(&ub.ToolsNodeCallbackHandlers{
	OnStart: func(ctx context.Context, runInfo *callbacks.RunInfo, input *schema.Message) context.Context {
		fmt.Printf("\n[TEST] Tools Start: %s %s\n", runInfo.Name, input.ToolName)
		return ctx
	},
	OnEnd: func(ctx context.Context, runInfo *callbacks.RunInfo, output []*schema.Message) context.Context {
		outputContent := ""
		for _, msg := range output {
			outputContent += msg.Content + "\n"
		}
		fmt.Printf("\n[TEST] Tools End: %s %s\n", runInfo.Name, outputContent)
		return ctx
	},
	OnEndWithStreamOutput: func(ctx context.Context, info *callbacks.RunInfo, output *schema.StreamReader[[]*schema.Message]) context.Context {
		fmt.Printf("\n[TEST] Tools Stream End: %s\n", info.Name)
		for {
			next, err := output.Recv()
			if err != nil {
				break
			}
			outputContent := ""
			for _, msg := range next {
				if msg != nil {
					outputContent += msg.Content
				}
			}
			fmt.Printf("%s", outputContent)
		}
		fmt.Printf("\n")
		return ctx
	},
}).Handler()

// createTestChatModel creates a chat model for testing
func createTestChatModel(t *testing.T) model.ToolCallingChatModel {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set, skipping integration test")
	}

	chatModel, err := openai.NewChatModel(context.Background(), &openai.ChatModelConfig{
		APIKey:  apiKey,
		Model:   "gpt-4o-mini",
		BaseURL: "https://api.openai.com/v1",
	})
	require.NoError(t, err)
	return chatModel
}

// createTestConfig creates a test configuration for the ReAct agent
func createTestConfig(chatModel model.ToolCallingChatModel) *ReactAgentConfig {
	return &ReactAgentConfig{
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
}

// TestNewAgent_Success tests successful agent creation
func TestNewAgent_Success(t *testing.T) {
	ctx := context.Background()
	chatModel := createTestChatModel(t)
	config := createTestConfig(chatModel)

	// Test successful agent creation
	agent, err := NewAgent(ctx, *config)
	require.NoError(t, err)
	require.NotNil(t, agent)

	// Verify agent properties
	assert.NotNil(t, agent.runnable)
	assert.NotNil(t, agent.graph)
	assert.Equal(t, config.MaxStep, agent.config.MaxStep)
	assert.Equal(t, config.DebugMode, agent.config.DebugMode)
}

// TestNewAgent_InvalidConfig tests agent creation with invalid config
func TestNewAgent_InvalidConfig(t *testing.T) {
	ctx := context.Background()

	// Test with nil config
	_, err := NewAgent(ctx, ReactAgentConfig{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ToolCallingModel cannot be nil")

	// Test with nil model
	config := &ReactAgentConfig{
		ToolCallingModel: nil,
		MaxStep:          10,
	}
	_, err = NewAgent(ctx, *config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ToolCallingModel cannot be nil")
}

// TestAgent_Generate_SimpleCalculation tests agent generation with simple calculation
func TestAgent_Generate_SimpleCalculation(t *testing.T) {
	ctx := context.Background()
	chatModel := createTestChatModel(t)
	config := createTestConfig(chatModel)

	agent, err := NewAgent(ctx, *config)
	require.NoError(t, err)

	// Test simple calculation
	messages := []*schema.Message{
		schema.UserMessage("请计算 10+100 等于多少"),
	}

	result, err := agent.Generate(ctx, messages, base.WithComposeOptions(compose.WithCallbacks(testCallbackHandler)))
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify result contains expected content
	assert.NotEmpty(t, result.Content)
	// fmt.Printf("\n[TEST] Generate Result: %s\n", result.Content)
}

// TestAgent_Stream_SimpleCalculation tests agent streaming with simple calculation
func TestAgent_Stream_SimpleCalculation(t *testing.T) {
	ctx := context.Background()
	chatModel := createTestChatModel(t)
	config := createTestConfig(chatModel)

	agent, err := NewAgent(ctx, *config)
	require.NoError(t, err)

	// Test streaming with simple calculation
	messages := []*schema.Message{
		schema.UserMessage("请计算 10+5 等于多少"),
	}

	stream, err := agent.Stream(ctx, messages, base.WithComposeOptions(compose.WithCallbacks(testCallbackHandler)))
	require.NoError(t, err)
	require.NotNil(t, stream)

	// Collect all chunks from stream
	finalResponse, err := schema.ConcatMessageStream(stream)
	require.NoError(t, err)
	require.NotNil(t, finalResponse)

	// Verify result contains expected content
	assert.NotEmpty(t, finalResponse.Content)
	fmt.Printf("\n[TEST] Stream Result: %s\n", finalResponse.Content)
}

// TestAgent_Generate_WeatherQuery tests agent generation with weather query
func TestAgent_Generate_WeatherQuery(t *testing.T) {
	ctx := context.Background()
	chatModel := createTestChatModel(t)
	config := createTestConfig(chatModel)

	agent, err := NewAgent(ctx, *config)
	require.NoError(t, err)

	// Test weather query
	messages := []*schema.Message{
		schema.UserMessage("请查询北京今天的天气情况"),
	}

	result, err := agent.Stream(ctx, messages, base.WithComposeOptions(compose.WithCallbacks(testCallbackHandler)))
	require.NoError(t, err)
	require.NotNil(t, result)
	result.Close()
}

// TestAgent_Generate_SearchQuery tests agent generation with search query
func TestAgent_Generate_SearchQuery(t *testing.T) {
	// Skip this test for now due to complexity
	// t.Skip("Skipping search query test - requires optimization")

	ctx := context.Background()
	chatModel := createTestChatModel(t)
	config := createTestConfig(chatModel)
	// Increase max iterations for search queries
	config.MaxStep = 25

	agent, err := NewAgent(ctx, *config)
	require.NoError(t, err)

	// Test search query with simpler request
	messages := []*schema.Message{
		schema.UserMessage("搜索人工智能"),
	}

	result, err := agent.Stream(ctx, messages, base.WithComposeOptions(compose.WithCallbacks(testCallbackHandler)))
	require.NoError(t, err)
	require.NotNil(t, result)
	result.Close()
}

// TestAgent_Generate_ComplexQuery tests agent generation with complex multi-step query
func TestAgent_Generate_ComplexQuery(t *testing.T) {
	ctx := context.Background()
	chatModel := createTestChatModel(t)
	config := createTestConfig(chatModel)
	// Increase max iterations for complex queries
	config.MaxStep = 25

	agent, err := NewAgent(ctx, *config)
	require.NoError(t, err)

	// Test complex multi-step query with simpler request
	messages := []*schema.Message{
		schema.UserMessage("计算20*3，然后查询上海天气"),
	}

	stream, err := agent.Stream(ctx, messages, base.WithComposeOptions(compose.WithCallbacks(testCallbackHandler)))
	require.NoError(t, err)
	require.NotNil(t, stream)
	stream.Close()
}

// TestAgent_Generate_WithOptions tests agent generation with options
func TestAgent_Generate_WithOptions(t *testing.T) {
	ctx := context.Background()
	chatModel := createTestChatModel(t)
	config := createTestConfig(chatModel)

	agent, err := NewAgent(ctx, *config)
	require.NoError(t, err)

	// Test with callback options
	messages := []*schema.Message{
		schema.UserMessage("请计算 100/4 等于多少"),
	}

	opts := []base.AgentOption{
		base.WithComposeOptions(compose.WithCallbacks(testCallbackHandler)),
	}

	result, err := agent.Generate(ctx, messages, opts...)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify result contains expected content
	assert.NotEmpty(t, result.Content)
	fmt.Printf("\n[TEST] Generate with Options Result: %s\n", result.Content)
}

// TestAgent_Stream_WithTimeout tests agent streaming with timeout
func TestAgent_Stream_WithTimeout(t *testing.T) {
	ctx := context.Background()
	chatModel := createTestChatModel(t)
	config := createTestConfig(chatModel)

	agent, err := NewAgent(ctx, *config)
	require.NoError(t, err)

	// Create context with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Test streaming with timeout
	messages := []*schema.Message{
		schema.UserMessage("请简单介绍一下你自己"),
	}

	stream, err := agent.Stream(ctxWithTimeout, messages)
	require.NoError(t, err)
	require.NotNil(t, stream)

	// Collect all chunks from stream
	finalResponse, err := schema.ConcatMessageStream(stream)
	require.NoError(t, err)
	require.NotNil(t, finalResponse)

	// Verify result contains expected content
	assert.NotEmpty(t, finalResponse.Content)
	fmt.Printf("\n[TEST] Stream with Timeout Result: %s\n", finalResponse.Content)
}

// TestAgent_Generate_EmptyInput tests agent generation with empty input
func TestAgent_Generate_EmptyInput(t *testing.T) {
	ctx := context.Background()
	chatModel := createTestChatModel(t)
	config := createTestConfig(chatModel)

	agent, err := NewAgent(ctx, *config)
	require.NoError(t, err)

	// Test with empty messages
	_, err = agent.Generate(ctx, []*schema.Message{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "input messages cannot be empty")

	// Test with nil messages
	_, err = agent.Generate(ctx, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "input messages cannot be empty")
}

// TestAgent_Stream_EmptyInput tests agent streaming with empty input
func TestAgent_Stream_EmptyInput(t *testing.T) {
	ctx := context.Background()
	chatModel := createTestChatModel(t)
	config := createTestConfig(chatModel)

	agent, err := NewAgent(ctx, *config)
	require.NoError(t, err)

	// Test with empty messages
	_, err = agent.Stream(ctx, []*schema.Message{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "input messages cannot be empty")

	// Test with nil messages
	_, err = agent.Stream(ctx, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "input messages cannot be empty")
}

// TestParseReasoningResponse tests the reasoning response parsing function
func TestParseReasoningResponse(t *testing.T) {
	// Test with tool calls in message
	messageWithToolCalls := &schema.Message{
		Role:    schema.Assistant,
		Content: "我需要使用计算器",
		ToolCalls: []schema.ToolCall{
			{
				ID:   "call_1",
				Type: "function",
				Function: schema.FunctionCall{
					Name:      "calculator",
					Arguments: `{"expression": "10+5"}`,
				},
			},
		},
	}

	reasoning, err := parseReasoningResponse(messageWithToolCalls)
	require.NoError(t, err)
	assert.Equal(t, "tool_call", reasoning.Action)
	assert.Equal(t, "我需要使用计算器", reasoning.Thought)
	assert.Len(t, reasoning.ToolCalls, 1)
	assert.Equal(t, "calculator", reasoning.ToolCalls[0].Function.Name)

	// Test JSON format parsing
	jsonContent := `{
		"thought": "用户询问计算问题，我需要使用计算器工具",
		"action": "tool_call",
		"confidence": 0.9
	}`

	jsonMessage := &schema.Message{
		Role:    schema.Assistant,
		Content: jsonContent,
	}

	reasoning2, err := parseReasoningResponse(jsonMessage)
	require.NoError(t, err)
	assert.Equal(t, "tool_call", reasoning2.Action)
	assert.Equal(t, "用户询问计算问题，我需要使用计算器工具", reasoning2.Thought)
	assert.Equal(t, 0.9, reasoning2.Confidence)

	// Test fallback parsing
	plainMessage := &schema.Message{
		Role:    schema.Assistant,
		Content: "这是一个普通的回复",
	}

	reasoning3, err := parseReasoningResponse(plainMessage)
	require.NoError(t, err)
	assert.Equal(t, "continue", reasoning3.Action)
	assert.Equal(t, "这是一个普通的回复", reasoning3.Thought)
	assert.Equal(t, 0.8, reasoning3.Confidence) // Default confidence
}

// TestValidateConfig tests the configuration validation function
func TestValidateConfig(t *testing.T) {
	// Test valid config
	chatModel := createTestChatModel(t)
	validConfig := createTestConfig(chatModel)
	err := validateConfig(validConfig)
	assert.NoError(t, err)

	// Test nil config
	err = validateConfig(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config cannot be nil")

	// Test config with nil model
	invalidConfig := &ReactAgentConfig{
		ToolCallingModel: nil,
		MaxStep:          10,
	}
	err = validateConfig(invalidConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ToolCallingModel cannot be nil")

	// Test config with invalid MaxStep
	invalidStepConfig := &ReactAgentConfig{
		ToolCallingModel: chatModel,
		MaxStep:          0,
	}
	err = validateConfig(invalidStepConfig)
	assert.NoError(t, err) // Should auto-correct to default value
	assert.Equal(t, 10, invalidStepConfig.MaxStep)
}

// BenchmarkAgent_Generate benchmarks the agent generation performance
func BenchmarkAgent_Generate(b *testing.B) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		b.Skip("OPENAI_API_KEY not set, skipping benchmark")
	}

	ctx := context.Background()
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:  apiKey,
		Model:   "gpt-4o-mini",
		BaseURL: "https://api.openai.com/v1",
	})
	if err != nil {
		b.Fatalf("Failed to create chat model: %v", err)
	}

	config := createTestConfig(chatModel)
	agent, err := NewAgent(ctx, *config)
	if err != nil {
		b.Fatalf("Failed to create agent: %v", err)
	}

	messages := []*schema.Message{
		schema.UserMessage("请计算 2+2 等于多少"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := agent.Generate(ctx, messages)
		if err != nil {
			b.Fatalf("Generate failed: %v", err)
		}
	}
}
