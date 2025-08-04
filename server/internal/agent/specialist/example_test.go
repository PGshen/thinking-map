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

package enhanced

import (
	"context"
	"testing"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/stretchr/testify/assert"
)

// MockChatModel 模拟ChatModel用于测试
type MockChatModel struct {
	response string
}

func (m *MockChatModel) Generate(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	return schema.AssistantMessage(m.response, nil), nil
}

func (m *MockChatModel) Stream(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	result := schema.AssistantMessage(m.response, nil)
	return schema.StreamReaderFromArray([]*schema.Message{result}), nil
}

// TestPlanningMultiAgentBasic 测试基本的规划多智能体功能
func TestPlanningMultiAgentBasic(t *testing.T) {
	// 创建模拟的ChatModel
	plannerModel := &MockChatModel{response: "Planning response"}
	specialistModel := &MockChatModel{response: "Specialist response"}
	summarizerModel := &MockChatModel{response: "Summary response"}

	// 创建配置
	config := &PlanningMultiAgentConfig{
		PlannerAgent: &PlannerAgent{
			ChatModel:       plannerModel,
			PlanningPrompt:  "Please create a plan for: {{query}}",
			UpdatePrompt:    "Please update the plan based on: {{results}}",
		},
		Specialists: []*Specialist{
			{
				AgentMeta: &AgentMeta{
					Name: "TestSpecialist",
				},
				ChatModel:    specialistModel,
				SystemPrompt: "You are a test specialist",
			},
		},
		MaxIterations: 3,
		Summarizer: &Summarizer{
			ChatModel:     summarizerModel,
			SummaryPrompt: "Please summarize: {{execution_results}}",
		},
	}

	// 创建规划多智能体
	agent, err := NewPlanningMultiAgent(config)
	assert.NoError(t, err)
	assert.NotNil(t, agent)

	// 测试基本配置
	assert.Equal(t, config, agent.config)
	assert.NotNil(t, agent.graph)
}

// TestPlanningMultiAgentWithCallbacks 测试带回调的规划多智能体
func TestPlanningMultiAgentWithCallbacks(t *testing.T) {
	// 创建模拟的ChatModel
	plannerModel := &MockChatModel{response: "Planning response"}
	specialistModel := &MockChatModel{response: "Specialist response"}

	// 创建配置
	config := &PlanningMultiAgentConfig{
		PlannerAgent: &PlannerAgent{
			ChatModel:       plannerModel,
			PlanningPrompt:  "Please create a plan for: {{query}}",
			UpdatePrompt:    "Please update the plan based on: {{results}}",
		},
		Specialists: []*Specialist{
			{
				AgentMeta: &AgentMeta{
					Name: "TestSpecialist",
				},
				ChatModel:    specialistModel,
				SystemPrompt: "You are a test specialist",
			},
		},
		MaxIterations: 3,
	}

	// 创建回调
	callback := &DefaultCallback{}

	// 创建规划多智能体（带回调）
	agent, err := NewPlanningMultiAgent(config, WithCallbacks(callback))
	assert.NoError(t, err)
	assert.NotNil(t, agent)
}

// TestPlanningMultiAgentGenerate 测试生成功能
func TestPlanningMultiAgentGenerate(t *testing.T) {
	// 创建模拟的ChatModel
	plannerModel := &MockChatModel{response: "Planning response"}
	specialistModel := &MockChatModel{response: "Specialist response"}

	// 创建配置
	config := &PlanningMultiAgentConfig{
		PlannerAgent: &PlannerAgent{
			ChatModel:       plannerModel,
			PlanningPrompt:  "Please create a plan for: {{query}}",
			UpdatePrompt:    "Please update the plan based on: {{results}}",
		},
		Specialists: []*Specialist{
			{
				AgentMeta: &AgentMeta{
					Name: "TestSpecialist",
				},
				ChatModel:    specialistModel,
				SystemPrompt: "You are a test specialist",
			},
		},
		MaxIterations: 3,
	}

	// 创建规划多智能体
	agent, err := NewPlanningMultiAgent(config)
	assert.NoError(t, err)

	// 测试生成功能
	ctx := context.Background()
	input := schema.UserMessage("Test query")

	response, err := agent.Generate(ctx, input)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "Planning MultiAgent is under development", response.Content)
}

// TestPlanningMultiAgentStream 测试流式生成功能
func TestPlanningMultiAgentStream(t *testing.T) {
	// 创建模拟的ChatModel
	plannerModel := &MockChatModel{response: "Planning response"}
	specialistModel := &MockChatModel{response: "Specialist response"}

	// 创建配置
	config := &PlanningMultiAgentConfig{
		PlannerAgent: &PlannerAgent{
			ChatModel:       plannerModel,
			PlanningPrompt:  "Please create a plan for: {{query}}",
			UpdatePrompt:    "Please update the plan based on: {{results}}",
		},
		Specialists: []*Specialist{
			{
				AgentMeta: &AgentMeta{
					Name: "TestSpecialist",
				},
				ChatModel:    specialistModel,
				SystemPrompt: "You are a test specialist",
			},
		},
		MaxIterations: 3,
	}

	// 创建规划多智能体
	agent, err := NewPlanningMultiAgent(config)
	assert.NoError(t, err)

	// 测试流式生成功能
	ctx := context.Background()
	input := schema.UserMessage("Test query")

	stream, err := agent.Stream(ctx, input)
	assert.NoError(t, err)
	assert.NotNil(t, stream)

	// 读取流中的消息
	msg, err := stream.Recv()
	assert.NoError(t, err)
	assert.NotNil(t, msg)
	assert.Equal(t, "Planning MultiAgent is under development", msg.Content)

	// 关闭流
	stream.Close()
}