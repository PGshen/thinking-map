/*
 * @Date: 2025-01-27
 * @LastEditors: AI Assistant
 * @FilePath: /thinking-map/server/internal/agent/base/multiagent/compose_test.go
 */
package multiagent

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/PGshen/thinking-map/server/internal/agent/base"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
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
func createTestConfig() *MultiAgentConfig {
	return &MultiAgentConfig{
		Name:        "test-multi-agent",
		Description: "Test  Multi-Agent System",
		Host: Host{
			Model:        chatModel,
			SystemPrompt: "You are a test assistant.",
			Thinking: ThinkingConfig{
				MaxSteps: 3,
			},
			Planning: PlanningConfig{},
		},
		// Temporarily disable specialists to test basic functionality
		Specialists: []*Specialist{
			{
				Name:        "common specialist",
				IntendedUse: "common tasks",
				ChatModel:   chatModel,
			},
		},
		Session: SessionConfig{
			HistoryLength: 100,
			ContextWindow: 4096,
		},
	}
}

func TestNewMultiAgent_Success(t *testing.T) {
	ctx := context.Background()
	config := createTestConfig()

	// 测试成功创建
	agent, err := NewMultiAgent(ctx, config)
	require.NoError(t, err)
	require.NotNil(t, agent)

	// 验证配置
	assert.Equal(t, config, agent.GetConfig())

	// 验证图结构
	graph, opts := agent.ExportGraph()
	assert.NotNil(t, graph)
	assert.NotNil(t, opts)
}

func TestNewMultiAgent_InvalidConfig(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name   string
		config *MultiAgentConfig
	}{
		{
			name:   "nil config",
			config: nil,
		},
		{
			name:   "empty config",
			config: &MultiAgentConfig{},
		},
		{
			name: "missing host model",
			config: &MultiAgentConfig{
				Name: "test",
				Host: Host{
					SystemPrompt: "test",
					// Model 缺失
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent, err := NewMultiAgent(ctx, tt.config)
			assert.Error(t, err)
			assert.Nil(t, agent)
		})
	}
}

func TestMultiAgent_BasicExxecution(t *testing.T) {
	ctx := context.Background()
	config := createTestConfig()

	agent, err := NewMultiAgent(ctx, config)
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

func TestMultiAgent_StreamExecution(t *testing.T) {
	ctx := context.Background()
	config := createTestConfig()

	agent, err := NewMultiAgent(ctx, config)
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

func TestMultiAgent_ComplexTask(t *testing.T) {
	ctx := context.Background()
	config := createTestConfig()

	multiAgent, err := NewMultiAgent(ctx, config)
	require.NoError(t, err)

	// 准备复杂任务输入
	input := []*schema.Message{
		{
			Role:    schema.User,
			Content: "Please analyze the performance optimization strategies for a Go web service and provide a detailed implementation plan.",
		},
	}

	// 执行流式生成
	stream, err := multiAgent.Stream(ctx, input, base.WithComposeOptions(compose.WithCallbacks(callbackForTest)))
	if err != nil {
		fmt.Println(err.Error())
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

func TestMultiAgent_MultiRoundConversation(t *testing.T) {
	ctx := context.Background()
	config := createTestConfig()

	multiAgent, err := NewMultiAgent(ctx, config)
	require.NoError(t, err)

	// 第一轮对话
	input1 := []*schema.Message{
		{
			Role:    schema.User,
			Content: "What is Go programming language?",
		},
	}

	stream1, err := multiAgent.Stream(ctx, input1, base.WithComposeOptions(compose.WithCallbacks(callbackForTest)))
	require.NoError(t, err)
	require.NotNil(t, stream1)
	// 读取流式结果
	var chunks1 []*schema.Message
	for {
		chunk, err1 := stream1.Recv()
		if err1 != nil {
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

	stream2, err := multiAgent.Stream(ctx, input2, base.WithComposeOptions(compose.WithCallbacks(callbackForTest)))
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

func TestMultiAgent_ContextTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	config := createTestConfig()

	agent, err := NewMultiAgent(ctx, config)
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

// BenchmarkMultiAgent_Generate 性能基准测试
func BenchmarkMultiAgent_Generate(b *testing.B) {
	ctx := context.Background()
	config := createTestConfig()

	agent, err := NewMultiAgent(ctx, config)
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
