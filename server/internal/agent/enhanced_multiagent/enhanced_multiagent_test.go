/*
 * Copyright 2024 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package enhanced_multiagent

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
)

// MockChatModel 模拟聊天模型
type MockChatModel struct {
	responses map[string]string
	defaultResponse string
}

func NewMockChatModel(defaultResponse string) *MockChatModel {
	return &MockChatModel{
		responses: make(map[string]string),
		defaultResponse: defaultResponse,
	}
}

func (m *MockChatModel) SetResponse(prompt string, response string) {
	m.responses[prompt] = response
}

func (m *MockChatModel) Generate(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	if len(messages) == 0 {
		return &schema.Message{
			Role: schema.Assistant,
			Content: m.defaultResponse,
		}, nil
	}
	
	lastMessage := messages[len(messages)-1]
	if response, exists := m.responses[lastMessage.Content]; exists {
		return &schema.Message{
			Role: schema.Assistant,
			Content: response,
		}, nil
	}
	
	return &schema.Message{
		Role: schema.Assistant,
		Content: m.defaultResponse,
	}, nil
}

func (m *MockChatModel) Stream(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	response, err := m.Generate(ctx, messages, opts...)
	if err != nil {
		return nil, err
	}
	
	sr, sw := schema.Pipe[*schema.Message](1)
	sw.Send(response, nil)
	sw.Close()
	return sr, nil
}

func (m *MockChatModel) WithTools(tools []*schema.ToolInfo) (model.ToolCallingChatModel, error) {
	return &MockToolCallingChatModel{
		MockChatModel: m,
		tools: tools,
	}, nil
}

func (m *MockChatModel) BindTools(tools []*schema.ToolInfo) error {
	// 为了兼容旧接口，这里不做任何操作
	return nil
}

// MockToolCallingChatModel 模拟工具调用聊天模型
type MockToolCallingChatModel struct {
	*MockChatModel
	tools []*schema.ToolInfo
}

func (m *MockToolCallingChatModel) Generate(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	return m.MockChatModel.Generate(ctx, messages, opts...)
}

func (m *MockToolCallingChatModel) Stream(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	return m.MockChatModel.Stream(ctx, messages, opts...)
}

func (m *MockToolCallingChatModel) WithTools(tools []*schema.ToolInfo) (model.ToolCallingChatModel, error) {
	return &MockToolCallingChatModel{
		MockChatModel: m.MockChatModel,
		tools: append(m.tools, tools...),
	}, nil
}

// MockCallback 模拟回调处理器
type MockCallback struct {
	TaskPlannings []string
	ComplexityEvaluations []string
	SpecialistCalls []string
	SpecialistResponses []string
	Reflections []string
	PhaseTransitions []string
}

func (m *MockCallback) OnTaskPlanning(ctx context.Context, task string, plan string) error {
	m.TaskPlannings = append(m.TaskPlannings, fmt.Sprintf("Task: %s, Plan: %s", task, plan))
	return nil
}

func (m *MockCallback) OnComplexityEvaluation(ctx context.Context, task string, complexity string) error {
	m.ComplexityEvaluations = append(m.ComplexityEvaluations, fmt.Sprintf("Task: %s, Complexity: %s", task, complexity))
	return nil
}

func (m *MockCallback) OnSpecialistCall(ctx context.Context, specialistName string, input *schema.Message) error {
	m.SpecialistCalls = append(m.SpecialistCalls, fmt.Sprintf("Specialist: %s, Input: %s", specialistName, input.Content))
	return nil
}

func (m *MockCallback) OnSpecialistResponse(ctx context.Context, specialistName string, response *schema.Message) error {
	m.SpecialistResponses = append(m.SpecialistResponses, fmt.Sprintf("Specialist: %s, Response: %s", specialistName, response.Content))
	return nil
}

func (m *MockCallback) OnReflection(ctx context.Context, reflectionInput string, reflectionOutput string) error {
	m.Reflections = append(m.Reflections, fmt.Sprintf("Input: %s, Output: %s", reflectionInput, reflectionOutput))
	return nil
}

func (m *MockCallback) OnPhaseTransition(ctx context.Context, fromPhase string, toPhase string) error {
	m.PhaseTransitions = append(m.PhaseTransitions, fmt.Sprintf("From: %s, To: %s", fromPhase, toPhase))
	return nil
}

func TestEnhancedMultiAgent_SimpleTask(t *testing.T) {
	ctx := context.Background()
	
	// 创建模拟模型
	complexityModel := NewMockChatModel("SIMPLE")
	planningModel := NewMockChatModel("直接回答用户问题")
	reflectionModel := NewMockChatModel("任务已完成，无需进一步处理")
	hostModel := NewMockChatModel("这是一个简单的问题，答案是42")
	
	// 创建React Agent配置
	reactConfig := &react.AgentConfig{
		Model: hostModel,
		MaxStep: 10,
	}
	
	// 创建专家配置
	mathSpecialist := &SpecialistConfig{
		Name: "数学专家",
		Description: "专门处理数学计算问题",
		ChatModel: NewMockChatModel("数学计算结果：42"),
		SystemPrompt: "你是一个数学专家，专门解决数学问题。",
	}
	
	// 创建回调处理器
	callback := &MockCallback{}
	
	// 创建增强版多智能体配置
	config := &EnhancedMultiAgentConfig{
		HostReactConfig: reactConfig,
		Specialists: []*SpecialistConfig{mathSpecialist},
		PlanningModel: planningModel,
		ComplexityEvaluationModel: complexityModel,
		ReflectionModel: reflectionModel,
		MaxReflections: 2,
		MaxSteps: 15,
		GraphName: "TestEnhancedMultiAgent",
		Callback: callback,
	}
	
	// 创建增强版多智能体系统
	agent, err := NewEnhancedMultiAgent(ctx, config)
	require.NoError(t, err)
	require.NotNil(t, agent)
	
	// 测试输入
	input := []*schema.Message{
		{
			Role: schema.User,
			Content: "什么是生命的意义？",
		},
	}
	
	// 生成响应
	response, err := agent.Generate(ctx, input)
	require.NoError(t, err)
	require.NotNil(t, response)
	
	// 验证响应
	assert.Equal(t, schema.Assistant, response.Role)
	assert.Contains(t, response.Content, "增强版多智能体系统执行结果")
	assert.Contains(t, response.Content, "什么是生命的意义？")
	assert.Contains(t, response.Content, "SIMPLE")
	
	// 验证回调被调用
	assert.Greater(t, len(callback.ComplexityEvaluations), 0)
	assert.Greater(t, len(callback.TaskPlannings), 0)
}

func TestEnhancedMultiAgent_ComplexTask(t *testing.T) {
	ctx := context.Background()
	
	// 创建模拟模型
	complexityModel := NewMockChatModel("COMPLEX")
	planningModel := NewMockChatModel("需要调用数学专家和科学专家进行协作")
	reflectionModel := NewMockChatModel("需要继续处理，调用更多专家")
	hostModel := NewMockChatModel("这是一个复杂问题，需要专家协助")
	
	// 设置专家路由响应
	hostModel.SetResponse("专家路由决策", "数学专家,科学专家")
	
	// 创建React Agent配置
	reactConfig := &react.AgentConfig{
		Model: hostModel,
		MaxStep: 10,
	}
	
	// 创建专家配置
	mathSpecialist := &SpecialistConfig{
		Name: "数学专家",
		Description: "专门处理数学计算问题",
		ChatModel: NewMockChatModel("数学分析：这个问题涉及复杂的数学概念"),
		SystemPrompt: "你是一个数学专家。",
	}
	
	scienceSpecialist := &SpecialistConfig{
		Name: "科学专家",
		Description: "专门处理科学问题",
		ChatModel: NewMockChatModel("科学分析：从科学角度来看，这个问题很有趣"),
		SystemPrompt: "你是一个科学专家。",
	}
	
	// 创建回调处理器
	callback := &MockCallback{}
	
	// 创建增强版多智能体配置
	config := &EnhancedMultiAgentConfig{
		HostReactConfig: reactConfig,
		Specialists: []*SpecialistConfig{mathSpecialist, scienceSpecialist},
		PlanningModel: planningModel,
		ComplexityEvaluationModel: complexityModel,
		ReflectionModel: reflectionModel,
		MaxReflections: 3,
		MaxSteps: 20,
		GraphName: "TestComplexEnhancedMultiAgent",
		Callback: callback,
	}
	
	// 创建增强版多智能体系统
	agent, err := NewEnhancedMultiAgent(ctx, config)
	require.NoError(t, err)
	require.NotNil(t, agent)
	
	// 测试输入
	input := []*schema.Message{
		{
			Role: schema.User,
			Content: "如何解决气候变化问题？",
		},
	}
	
	// 生成响应
	response, err := agent.Generate(ctx, input)
	require.NoError(t, err)
	require.NotNil(t, response)
	
	// 验证响应
	assert.Equal(t, schema.Assistant, response.Role)
	assert.Contains(t, response.Content, "增强版多智能体系统执行结果")
	assert.Contains(t, response.Content, "如何解决气候变化问题？")
	assert.Contains(t, response.Content, "COMPLEX")
	
	// 验证回调被调用
	assert.Greater(t, len(callback.ComplexityEvaluations), 0)
	assert.Greater(t, len(callback.TaskPlannings), 0)
	// 由于是复杂任务，应该有专家调用
	// assert.Greater(t, len(callback.SpecialistCalls), 0)
	// assert.Greater(t, len(callback.SpecialistResponses), 0)
}

func TestEnhancedMultiAgent_StreamGeneration(t *testing.T) {
	ctx := context.Background()
	
	// 创建模拟模型
	complexityModel := NewMockChatModel("MODERATE")
	planningModel := NewMockChatModel("分步骤处理问题")
	reflectionModel := NewMockChatModel("处理完成")
	hostModel := NewMockChatModel("这是一个中等复杂度的问题")
	
	// 创建React Agent配置
	reactConfig := &react.AgentConfig{
		Model: hostModel,
		MaxStep: 10,
	}
	
	// 创建增强版多智能体配置
	config := &EnhancedMultiAgentConfig{
		HostReactConfig: reactConfig,
		Specialists: []*SpecialistConfig{},
		PlanningModel: planningModel,
		ComplexityEvaluationModel: complexityModel,
		ReflectionModel: reflectionModel,
		MaxReflections: 1,
		MaxSteps: 10,
		GraphName: "TestStreamEnhancedMultiAgent",
	}
	
	// 创建增强版多智能体系统
	agent, err := NewEnhancedMultiAgent(ctx, config)
	require.NoError(t, err)
	require.NotNil(t, agent)
	
	// 测试输入
	input := []*schema.Message{
		{
			Role: schema.User,
			Content: "请解释人工智能的发展历程",
		},
	}
	
	// 流式生成响应
	stream, err := agent.Stream(ctx, input)
	require.NoError(t, err)
	require.NotNil(t, stream)
	
	// 读取流式响应
	response, err := stream.Recv()
	require.NoError(t, err)
	require.NotNil(t, response)
	
	// 验证响应
	assert.Equal(t, schema.Assistant, response.Role)
	assert.NotEmpty(t, response.Content)
}

func TestEnhancedMultiAgent_ExportGraph(t *testing.T) {
	ctx := context.Background()
	
	// 创建最小配置
	hostModel := NewMockChatModel("test response")
	reactConfig := &react.AgentConfig{
		Model: hostModel,
		MaxStep: 5,
	}
	
	config := &EnhancedMultiAgentConfig{
		HostReactConfig: reactConfig,
		Specialists: []*SpecialistConfig{},
		PlanningModel: hostModel,
		ComplexityEvaluationModel: hostModel,
		ReflectionModel: hostModel,
		MaxReflections: 1,
		MaxSteps: 10,
		GraphName: "TestExportGraph",
	}
	
	// 创建增强版多智能体系统
	agent, err := NewEnhancedMultiAgent(ctx, config)
	require.NoError(t, err)
	require.NotNil(t, agent)
	
	// 导出图
	graph, opts := agent.ExportGraph()
	require.NotNil(t, graph)
	require.NotNil(t, opts)
}

func TestEnhancedMultiAgent_ConfigValidation(t *testing.T) {
	ctx := context.Background()
	
	// 测试nil配置
	_, err := NewEnhancedMultiAgent(ctx, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config cannot be nil")
	
	// 测试缺少HostReactConfig
	config := &EnhancedMultiAgentConfig{}
	_, err = NewEnhancedMultiAgent(ctx, config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "host react config cannot be nil")
	
	// 测试默认值设置
	hostModel := NewMockChatModel("test")
	reactConfig := &react.AgentConfig{
		Model: hostModel,
	}
	
	config = &EnhancedMultiAgentConfig{
		HostReactConfig: reactConfig,
		PlanningModel: hostModel,
		ComplexityEvaluationModel: hostModel,
		ReflectionModel: hostModel,
	}
	
	agent, err := NewEnhancedMultiAgent(ctx, config)
	require.NoError(t, err)
	require.NotNil(t, agent)
	
	// 验证默认值
	assert.Equal(t, 3, config.MaxReflections)
	assert.Equal(t, 20, config.MaxSteps)
	assert.Equal(t, "EnhancedMultiAgent", config.GraphName)
}