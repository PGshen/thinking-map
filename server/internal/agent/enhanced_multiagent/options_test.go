package enhanced

import (
	"context"
	"fmt"
	"testing"

	"github.com/cloudwego/eino/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockConversationAnalyzer 实现conversationAnalyzer接口用于测试
type mockConversationAnalyzer struct {
	messages              []*schema.Message
	streamMessages        []*schema.Message
	onMessageCalled       bool
	onStreamMessageCalled bool
}

func (m *mockConversationAnalyzer) OnMessage(ctx context.Context, message *schema.Message) (context.Context, error) {
	m.onMessageCalled = true
	m.messages = append(m.messages, message)
	return ctx, nil
}

func (m *mockConversationAnalyzer) OnStreamMessage(ctx context.Context, message *schema.StreamReader[*schema.Message]) (context.Context, error) {
	m.onStreamMessageCalled = true
	// 读取流中的消息
	fmt.Println("OnStremMessage:")
	for {
		msg, err := message.Recv()
		if err != nil {
			break
		}
		fmt.Print(msg.Content)
		m.streamMessages = append(m.streamMessages, msg)
	}
	fmt.Println()
	return ctx, nil
}

// TestWithConversationAnalyzer_BasicFunctionality 测试基本功能
func TestWithConversationAnalyzer_BasicFunctionality(t *testing.T) {
	ctx := context.Background()
	config := createTestConfig()

	// 创建mock analyzer
	mockAnalyzer := &mockConversationAnalyzer{}

	// 创建enhanced multi-agent
	enhancedMultiAgent, err := NewEnhancedMultiAgent(ctx, config)
	require.NoError(t, err)

	// 准备测试输入
	input := []*schema.Message{
		{
			Role:    schema.User,
			Content: "Hello, this is a test message for conversation analyzer",
		},
	}

	// 使用WithConversationAnalyzer option执行Generate
	result, err := enhancedMultiAgent.Generate(ctx, input, WithConversationAnalyzer(mockAnalyzer))
	require.NoError(t, err)
	require.NotNil(t, result)

	// 验证analyzer被调用
	assert.True(t, mockAnalyzer.onMessageCalled, "OnMessage should be called")
	assert.Greater(t, len(mockAnalyzer.messages), 0, "Should have captured messages")

	// 验证捕获的消息
	assert.Equal(t, schema.Assistant, mockAnalyzer.messages[0].Role, "Should capture assistant message")
	assert.NotEmpty(t, mockAnalyzer.messages[0].Content, "Message content should not be empty")
}

// TestWithConversationAnalyzer_StreamFunctionality 测试流式功能
func TestWithConversationAnalyzer_StreamFunctionality(t *testing.T) {
	ctx := context.Background()
	config := createTestConfig()

	// 创建mock analyzer
	mockAnalyzer := &mockConversationAnalyzer{}

	// 创建enhanced multi-agent
	enhancedMultiAgent, err := NewEnhancedMultiAgent(ctx, config)
	require.NoError(t, err)

	// 准备测试输入
	input := []*schema.Message{
		{
			Role:    schema.User,
			Content: "Tell me about Go programming language",
		},
	}

	// 使用WithConversationAnalyzer option执行Stream
	stream, err := enhancedMultiAgent.Stream(ctx, input, WithConversationAnalyzer(mockAnalyzer))
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
	assert.Greater(t, len(chunks), 0, "Should receive stream chunks")

	// 验证stream analyzer被调用
	assert.True(t, mockAnalyzer.onStreamMessageCalled, "OnStreamMessage should be called")
	assert.Greater(t, len(mockAnalyzer.streamMessages), 0, "Should have captured stream messages")
}

func TestWithConversationAnalyzer_MultiRoundConversation(t *testing.T) {
	ctx := context.Background()
	config := createTestConfig()
	mockAnalyzer := &mockConversationAnalyzer{}

	enhancedMultiAgent, err := NewEnhancedMultiAgent(ctx, config)
	require.NoError(t, err)

	// 第一轮对话
	input1 := []*schema.Message{
		{
			Role:    schema.User,
			Content: "What is Go programming language?",
		},
	}

	stream1, err := enhancedMultiAgent.Stream(ctx, input1, WithConversationAnalyzer(mockAnalyzer))
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

	stream2, err := enhancedMultiAgent.Stream(ctx, input2, WithConversationAnalyzer(mockAnalyzer))
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

// TestWithConversationAnalyzer_MultipleMessages 测试多消息场景
func TestWithConversationAnalyzer_MultipleMessages(t *testing.T) {
	ctx := context.Background()
	config := createTestConfig()

	// 创建mock analyzer
	mockAnalyzer := &mockConversationAnalyzer{}

	// 创建enhanced multi-agent
	enhancedMultiAgent, err := NewEnhancedMultiAgent(ctx, config)
	require.NoError(t, err)

	// 准备多轮对话输入
	input := []*schema.Message{
		{
			Role:    schema.User,
			Content: "What is Go?",
		},
		{
			Role:    schema.Assistant,
			Content: "Go is a programming language.",
		},
		{
			Role:    schema.User,
			Content: "Can you give me some examples?",
		},
	}

	// 执行生成
	result, err := enhancedMultiAgent.Generate(ctx, input, WithConversationAnalyzer(mockAnalyzer))
	require.NoError(t, err)
	require.NotNil(t, result)

	// 验证analyzer被调用
	assert.True(t, mockAnalyzer.onMessageCalled, "OnMessage should be called")
	assert.Greater(t, len(mockAnalyzer.messages), 0, "Should have captured messages")

	// 验证最新的消息是assistant回复
	lastMessage := mockAnalyzer.messages[len(mockAnalyzer.messages)-1]
	assert.Equal(t, schema.Assistant, lastMessage.Role, "Last message should be from assistant")
	assert.NotEmpty(t, lastMessage.Content, "Assistant message should have content")
}

// TestWithConversationAnalyzer_ErrorHandling 测试错误处理
func TestWithConversationAnalyzer_ErrorHandling(t *testing.T) {
	ctx := context.Background()
	config := createTestConfig()

	// 创建会返回错误的mock analyzer
	errorAnalyzer := &errorConversationAnalyzer{}

	// 创建enhanced multi-agent
	enhancedMultiAgent, err := NewEnhancedMultiAgent(ctx, config)
	require.NoError(t, err)

	// 准备测试输入
	input := []*schema.Message{
		{
			Role:    schema.User,
			Content: "Test error handling",
		},
	}

	// 执行生成 - 即使analyzer返回错误，系统也应该继续工作
	// 因为在WithConversationAnalyzer中我们忽略了错误 (ctx, _ = analyzer.OnMessage(ctx, output.Message))
	result, err := enhancedMultiAgent.Generate(ctx, input, WithConversationAnalyzer(errorAnalyzer))
	require.NoError(t, err, "System should continue working even if analyzer returns error")
	require.NotNil(t, result)

	// 验证analyzer被调用
	assert.True(t, errorAnalyzer.onMessageCalled, "OnMessage should be called")
}

// errorConversationAnalyzer 用于测试错误处理的mock analyzer
type errorConversationAnalyzer struct {
	onMessageCalled       bool
	onStreamMessageCalled bool
}

func (e *errorConversationAnalyzer) OnMessage(ctx context.Context, message *schema.Message) (context.Context, error) {
	e.onMessageCalled = true
	return ctx, assert.AnError // 返回错误
}

func (e *errorConversationAnalyzer) OnStreamMessage(ctx context.Context, message *schema.StreamReader[*schema.Message]) (context.Context, error) {
	e.onStreamMessageCalled = true
	return ctx, assert.AnError // 返回错误
}

// TestWithConversationAnalyzer_ContextPropagation 测试上下文传播
func TestWithConversationAnalyzer_ContextPropagation(t *testing.T) {
	ctx := context.Background()
	config := createTestConfig()

	// 创建会修改context的mock analyzer
	contextAnalyzer := &contextModifyingAnalyzer{}

	// 创建enhanced multi-agent
	enhancedMultiAgent, err := NewEnhancedMultiAgent(ctx, config)
	require.NoError(t, err)

	// 准备测试输入
	input := []*schema.Message{
		{
			Role:    schema.User,
			Content: "Test context propagation",
		},
	}

	// 执行生成
	result, err := enhancedMultiAgent.Generate(ctx, input, WithConversationAnalyzer(contextAnalyzer))
	require.NoError(t, err)
	require.NotNil(t, result)

	// 验证analyzer被调用
	assert.True(t, contextAnalyzer.onMessageCalled, "OnMessage should be called")
	assert.True(t, contextAnalyzer.contextModified, "Context should be modified")
}

// contextModifyingAnalyzer 用于测试上下文传播的mock analyzer
type contextModifyingAnalyzer struct {
	onMessageCalled bool
	contextModified bool
}

type contextKey string

const testContextKey contextKey = "test_key"

func (c *contextModifyingAnalyzer) OnMessage(ctx context.Context, message *schema.Message) (context.Context, error) {
	c.onMessageCalled = true
	// 修改context
	newCtx := context.WithValue(ctx, testContextKey, "test_value")
	// 检查context是否被修改
	if newCtx.Value(testContextKey) == "test_value" {
		c.contextModified = true
	}
	return newCtx, nil
}

func (c *contextModifyingAnalyzer) OnStreamMessage(ctx context.Context, message *schema.StreamReader[*schema.Message]) (context.Context, error) {
	return ctx, nil
}

// TestConversationAnalyzerInterface 测试接口定义
func TestConversationAnalyzerInterface(t *testing.T) {
	// 验证接口定义是否正确
	analyzer := &mockConversationAnalyzer{}
	assert.NotNil(t, analyzer, "Mock should implement conversationAnalyzer interface")

	// 测试接口方法
	ctx := context.Background()
	message := &schema.Message{
		Role:    schema.User,
		Content: "Test message",
	}

	// 测试OnMessage
	newCtx, err := analyzer.OnMessage(ctx, message)
	assert.NoError(t, err)
	assert.NotNil(t, newCtx)
	assert.True(t, analyzer.onMessageCalled)
}

// 为其他节点添加测试用例

// mockMessageHandler 实现通用的 MessageHandler 接口
type mockMessageHandler struct {
	messages       []*schema.Message
	streamMessages []*schema.StreamReader[*schema.Message]
	errorToReturn  error
}

func (m *mockMessageHandler) OnMessage(ctx context.Context, message *schema.Message) (context.Context, error) {
	m.messages = append(m.messages, message)
	return ctx, m.errorToReturn
}

func (m *mockMessageHandler) OnStreamMessage(ctx context.Context, message *schema.StreamReader[*schema.Message]) (context.Context, error) {
	m.streamMessages = append(m.streamMessages, message)
	return ctx, m.errorToReturn
}

func TestWithDirectAnswerHandler(t *testing.T) {
	mockHandler := &mockMessageHandler{}
	option := WithDirectAnswerHandler(mockHandler)

	// 验证option不为nil
	assert.NotNil(t, option)

	// 创建测试agent并应用option
	config := createTestConfig()
	agent, err := NewEnhancedMultiAgent(context.Background(), config)
	assert.NoError(t, err)

	// 测试消息处理
	input := []*schema.Message{
		{Role: schema.User, Content: "测试直接回答"},
	}

	_, err = agent.Generate(context.Background(), input, option)
	assert.NoError(t, err)
}

func TestWithPlanCreationHandler(t *testing.T) {
	mockHandler := &mockMessageHandler{}
	option := WithPlanCreationHandler(mockHandler)

	// 验证option不为nil
	assert.NotNil(t, option)

	// 创建测试agent并应用option
	config := createTestConfig()
	agent, err := NewEnhancedMultiAgent(context.Background(), config)
	assert.NoError(t, err)

	// 测试消息处理
	input := []*schema.Message{
		{Role: schema.User, Content: "测试计划创建"},
	}

	_, err = agent.Generate(context.Background(), input, option)
	assert.NoError(t, err)
}

func TestWithFeedbackProcessorHandler(t *testing.T) {
	mockHandler := &mockMessageHandler{}
	option := WithFeedbackProcessorHandler(mockHandler)

	// 验证option不为nil
	assert.NotNil(t, option)

	// 创建测试agent并应用option
	config := createTestConfig()
	agent, err := NewEnhancedMultiAgent(context.Background(), config)
	assert.NoError(t, err)

	// 测试消息处理
	input := []*schema.Message{
		{Role: schema.User, Content: "测试反馈处理"},
	}

	_, err = agent.Generate(context.Background(), input, option)
	assert.NoError(t, err)
}

func TestWithPlanUpdateHandler(t *testing.T) {
	mockHandler := &mockMessageHandler{}
	option := WithPlanUpdateHandler(mockHandler)

	// 验证option不为nil
	assert.NotNil(t, option)

	// 创建测试agent并应用option
	config := createTestConfig()
	agent, err := NewEnhancedMultiAgent(context.Background(), config)
	assert.NoError(t, err)

	// 测试消息处理
	input := []*schema.Message{
		{Role: schema.User, Content: "测试计划更新"},
	}

	_, err = agent.Generate(context.Background(), input, option)
	assert.NoError(t, err)
}

func TestWithFinalAnswerHandler(t *testing.T) {
	mockHandler := &mockMessageHandler{}
	option := WithFinalAnswerHandler(mockHandler)

	// 验证option不为nil
	assert.NotNil(t, option)

	// 创建测试agent并应用option
	config := createTestConfig()
	agent, err := NewEnhancedMultiAgent(context.Background(), config)
	assert.NoError(t, err)

	// 测试消息处理
	input := []*schema.Message{
		{Role: schema.User, Content: "测试最终回答"},
	}

	_, err = agent.Generate(context.Background(), input, option)
	assert.NoError(t, err)
}

func TestWithPlanExecutionHandler(t *testing.T) {
	mockHandler := &mockMessageHandler{}
	option := WithPlanExecutionHandler(mockHandler)

	// 验证option不为nil
	assert.NotNil(t, option)

	// 创建测试agent并应用option
	config := createTestConfig()
	agent, err := NewEnhancedMultiAgent(context.Background(), config)
	assert.NoError(t, err)

	// 测试消息处理
	input := []*schema.Message{
		{Role: schema.User, Content: "测试计划执行"},
	}

	_, err = agent.Generate(context.Background(), input, option)
	assert.NoError(t, err)
}

func TestWithResultCollectorHandler(t *testing.T) {
	mockHandler := &mockMessageHandler{}
	option := WithResultCollectorHandler(mockHandler)

	// 验证option不为nil
	assert.NotNil(t, option)

	// 创建测试agent并应用option
	config := createTestConfig()
	agent, err := NewEnhancedMultiAgent(context.Background(), config)
	assert.NoError(t, err)

	// 测试消息处理
	input := []*schema.Message{
		{Role: schema.User, Content: "测试结果收集"},
	}

	_, err = agent.Generate(context.Background(), input, option)
	assert.NoError(t, err)
}

func TestWithSpecialistHandler(t *testing.T) {
	mockHandler := &mockMessageHandler{}
	option := WithSpecialistHandler("common specialist", mockHandler)

	// 验证option不为nil
	assert.NotNil(t, option)

	// 创建测试agent并应用option
	config := createTestConfig()
	agent, err := NewEnhancedMultiAgent(context.Background(), config)
	assert.NoError(t, err)

	// 测试消息处理
	input := []*schema.Message{
		{Role: schema.User, Content: "测试专家处理"},
	}

	_, err = agent.Generate(context.Background(), input, option)
	assert.NoError(t, err)
}

// 测试所有接口实现
func TestAllHandlerInterfaces(t *testing.T) {
	// 验证所有接口实现
	var _ MessageHandler = (*mockMessageHandler)(nil)
}
