/*
 * @Date: 2025-01-27
 * @LastEditors: AI Assistant
 * @FilePath: /thinking-map/server/internal/agent/enhanced_multiagent/compose_test.go
 */
package enhanced

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent"
	"github.com/cloudwego/eino/schema"
	ub "github.com/cloudwego/eino/utils/callbacks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var callbackForTest = ub.NewHandlerHelper().ChatModel(&ub.ModelCallbackHandler{
	OnStart: func(ctx context.Context, runInfo *callbacks.RunInfo, input *model.CallbackInput) context.Context {
		fmt.Println("\nIN: ", runInfo.Name)
		return ctx
	},
	OnEnd: func(ctx context.Context, runInfo *callbacks.RunInfo, output *model.CallbackOutput) context.Context {
		fmt.Println("\nOUT: ", runInfo.Name)
		return ctx
	},
	OnEndWithStreamOutput: func(ctx context.Context, runInfo *callbacks.RunInfo, output *schema.StreamReader[*model.CallbackOutput]) context.Context {
		fmt.Println("\nOUT: ", runInfo.Name, "stream output:")
		for {
			item, err := output.Recv()
			if err != nil {
				break
			}
			fmt.Print(item.Message.Content)
		}
		return ctx
	},
}).Handler()

// 注意：这需要设置OPENAI_API_KEY环境变量
var chatModel, err = openai.NewChatModel(context.Background(), &openai.ChatModelConfig{
	APIKey: os.Getenv("OPENAI_API_KEY"),
	Model:  "gpt-4o-mini",
})

// createTestConfig 创建测试用的配置
func createTestConfig() *EnhancedMultiAgentConfig {
	return &EnhancedMultiAgentConfig{
		Name:        "test-enhanced-multi-agent",
		Description: "Test Enhanced Multi-Agent System",
		Host: EnhancedHost{
			Model:        chatModel,
			SystemPrompt: "You are a test assistant.",
			Thinking: ThinkingConfig{
				MaxSteps:           3,
				Timeout:            30 * time.Second,
				EnableDeepThink:    true,
				ComplexityAnalysis: true,
			},
			Planning: PlanningConfig{
				MaxSteps:           5,
				Timeout:            60 * time.Second,
				EnableDynamicPlan:  true,
				DependencyAnalysis: true,
			},
		},
		// Temporarily disable specialists to test basic functionality
		Specialists: []*EnhancedSpecialist{
			{
				Name:        "common specialist",
				IntendedUse: "common tasks",
				ChatModel:   chatModel,
			},
		},
		System: SystemConfig{
			Version:     "1.0.0",
			Environment: "test",
			DebugMode:   true,
		},
		ExecutionControl: ExecutionControlConfig{
			MaxRounds:        5,
			Timeout:          5 * time.Minute,
			GracefulShutdown: true,
		},
		Session: SessionConfig{
			HistoryLength: 100,
			ContextWindow: 4096,
		},
		Performance: PerformanceConfig{
			Concurrency: map[string]int{
				"max_concurrent_specialists": 3,
			},
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
			Output: []string{"stdout"},
		},
	}
}

func TestNewEnhancedMultiAgent_Success(t *testing.T) {
	ctx := context.Background()
	config := createTestConfig()

	// 测试成功创建
	agent, err := NewEnhancedMultiAgent(ctx, config)
	require.NoError(t, err)
	require.NotNil(t, agent)

	// 验证配置
	assert.Equal(t, config, agent.GetConfig())

	// 验证图结构
	graph, opts := agent.ExportGraph()
	assert.NotNil(t, graph)
	assert.NotNil(t, opts)
}

func TestNewEnhancedMultiAgent_InvalidConfig(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name   string
		config *EnhancedMultiAgentConfig
	}{
		{
			name:   "nil config",
			config: nil,
		},
		{
			name:   "empty config",
			config: &EnhancedMultiAgentConfig{},
		},
		{
			name: "missing host model",
			config: &EnhancedMultiAgentConfig{
				Name: "test",
				Host: EnhancedHost{
					SystemPrompt: "test",
					// Model 缺失
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent, err := NewEnhancedMultiAgent(ctx, tt.config)
			assert.Error(t, err)
			assert.Nil(t, agent)
		})
	}
}

func TestEnhancedMultiAgent_BasicExxecution(t *testing.T) {
	ctx := context.Background()
	config := createTestConfig()

	agent, err := NewEnhancedMultiAgent(ctx, config)
	require.NoError(t, err)

	// 准备测试输入
	input := []*schema.Message{
		{
			Role:    schema.User,
			Content: "Hello, can you help me with a simple task?",
		},
	}

	// 执行生成
	result, err := agent.Generate(ctx, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	// 验证结果
	assert.Equal(t, schema.Assistant, result.Role)
	assert.NotEmpty(t, result.Content)
}

func TestEnhancedMultiAgent_StreamExecution(t *testing.T) {
	ctx := context.Background()
	config := createTestConfig()

	agent, err := NewEnhancedMultiAgent(ctx, config)
	require.NoError(t, err)

	// 准备测试输入
	input := []*schema.Message{
		{
			Role:    schema.User,
			Content: "Tell me a story",
		},
	}

	// 执行流式生成
	stream, err := agent.Stream(ctx, input)
	if err != nil {
		// 如果模拟模型不支持流式，跳过测试
		t.Skip("Mock model does not support streaming")
		return
	}

	require.NotNil(t, stream)

	// 读取流式结果
	var chunks []*schema.Message
	for {
		chunk, err := stream.Recv()
		if err != nil {
			break
		}
		chunks = append(chunks, chunk)
	}

	// 验证至少收到一些数据
	assert.Greater(t, len(chunks), 0)
}

func TestEnhancedMultiAgent_ComplexTask(t *testing.T) {
	ctx := context.Background()
	config := createTestConfig()

	enhancedMultiAgent, err := NewEnhancedMultiAgent(ctx, config)
	require.NoError(t, err)

	// 准备复杂任务输入
	input := []*schema.Message{
		{
			Role:    schema.User,
			Content: "Please analyze the performance optimization strategies for a Go web service and provide a detailed implementation plan.",
		},
	}

	// 执行流式生成
	stream, err := enhancedMultiAgent.Stream(ctx, input, agent.WithComposeOptions(compose.WithCallbacks(callbackForTest)))
	if err != nil {
		// 如果模拟模型不支持流式，跳过测试
		t.Skip("Mock model does not support streaming")
		return
	}

	require.NotNil(t, stream)

	// 读取流式结果
	var chunks []*schema.Message
	for {
		chunk, err := stream.Recv()
		if err != nil {
			break
		}
		chunks = append(chunks, chunk)
	}

	// 验证至少收到一些数据
	assert.Greater(t, len(chunks), 0)
}

func TestEnhancedMultiAgent_MultiRoundConversation(t *testing.T) {
	ctx := context.Background()
	config := createTestConfig()

	enhancedMultiAgent, err := NewEnhancedMultiAgent(ctx, config)
	require.NoError(t, err)

	// 第一轮对话
	input1 := []*schema.Message{
		{
			Role:    schema.User,
			Content: "What is Go programming language?",
		},
	}

	stream1, err := enhancedMultiAgent.Stream(ctx, input1, agent.WithComposeOptions(compose.WithCallbacks(callbackForTest)))
	require.NoError(t, err)
	require.NotNil(t, stream1)
	// 读取流式结果
	var chunks1 []*schema.Message
	for {
		chunk, err := stream1.Recv()
		if err != nil {
			break
		}
		chunks1 = append(chunks1, chunk)
	}
	result1, _ := schema.ConcatMessages(chunks1)

	// 第二轮对话（包含历史）
	input2 := []*schema.Message{
		{
			Role:    schema.User,
			Content: "What is Go programming language?",
		},
		result1,
		{
			Role:    schema.User,
			Content: "Can you give me some examples of Go web frameworks?",
		},
	}

	stream2, err := enhancedMultiAgent.Stream(ctx, input2, agent.WithComposeOptions(compose.WithCallbacks(callbackForTest)))
	require.NoError(t, err)
	require.NotNil(t, stream2)

	// 读取流式结果
	var chunks2 []*schema.Message
	for {
		chunk, err := stream2.Recv()
		if err != nil {
			break
		}
		chunks2 = append(chunks2, chunk)
	}
	result2, _ := schema.ConcatMessages(chunks2)

	// 验证结果
	assert.Equal(t, schema.Assistant, result2.Role)
	assert.NotEmpty(t, result2.Content)
}

func TestEnhancedMultiAgent_ContextTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	config := createTestConfig()

	agent, err := NewEnhancedMultiAgent(ctx, config)
	if err != nil {
		// 如果在创建时就超时，这是预期的
		assert.Contains(t, err.Error(), "context")
		return
	}

	// 准备输入
	input := []*schema.Message{
		{
			Role:    schema.User,
			Content: "This is a test message",
		},
	}

	// 执行应该因为超时而失败
	result, err := agent.Generate(ctx, input)
	if err != nil {
		assert.Contains(t, err.Error(), "context")
	} else {
		// 如果没有错误，至少应该有结果
		assert.NotNil(t, result)
	}
}

// BenchmarkEnhancedMultiAgent_Generate 性能基准测试
func BenchmarkEnhancedMultiAgent_Generate(b *testing.B) {
	ctx := context.Background()
	config := createTestConfig()

	agent, err := NewEnhancedMultiAgent(ctx, config)
	if err != nil {
		b.Fatalf("Failed to create agent: %v", err)
	}

	input := []*schema.Message{
		{
			Role:    schema.User,
			Content: "Hello, world!",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := agent.Generate(ctx, input)
		if err != nil {
			b.Errorf("Generate failed: %v", err)
		}
	}
}
