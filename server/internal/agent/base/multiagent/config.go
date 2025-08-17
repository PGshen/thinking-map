package multiagent

import (
	"errors"
	"fmt"
	"time"

	"github.com/PGshen/thinking-map/server/internal/agent/base"
	"github.com/PGshen/thinking-map/server/internal/agent/base/react"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// ThinkingConfig represents thinking configuration
type ThinkingConfig struct {
	MaxSteps           int           `yaml:"max_steps" json:"max_steps"`
	Timeout            time.Duration `yaml:"timeout" json:"timeout"`
	EnableDeepThink    bool          `yaml:"enable_deep_think" json:"enable_deep_think"`
	ComplexityAnalysis bool          `yaml:"complexity_analysis" json:"complexity_analysis"`
}

// PlanningConfig represents planning configuration
type PlanningConfig struct {
	MaxSteps           int           `yaml:"max_steps" json:"max_steps"`
	Timeout            time.Duration `yaml:"timeout" json:"timeout"`
	EnableDynamicPlan  bool          `yaml:"enable_dynamic_plan" json:"enable_dynamic_plan"`
	DependencyAnalysis bool          `yaml:"dependency_analysis" json:"dependency_analysis"`
}

// Host represents the host agent configuration
type Host struct {
	Model        model.ToolCallingChatModel `yaml:"model" json:"model"`
	SystemPrompt string                     `yaml:"system_prompt" json:"system_prompt"`
	Prompts      map[string]string          `yaml:"prompts,omitempty" json:"prompts,omitempty"`
	Thinking     ThinkingConfig             `yaml:"thinking" json:"thinking"`
	Planning     PlanningConfig             `yaml:"planning" json:"planning"`
}

// Specialist represents a specialist agent configuration
type Specialist struct {
	Name         string `yaml:"name" json:"name"`
	IntendedUse  string `yaml:"intended_use" json:"intended_use"`
	ChatModel    model.BaseChatModel
	SystemPrompt string `yaml:"system_prompt" json:"system_prompt"`
	Invokable    compose.Invoke[[]*schema.Message, *schema.Message, base.AgentOption]
	Streamable   compose.Stream[[]*schema.Message, *schema.Message, base.AgentOption]
	ReactAgent   *react.ReactAgent
}

// SessionConfig represents session management configuration
type SessionConfig struct {
	HistoryLength     int            `yaml:"history_length" json:"history_length"`
	ContextWindow     int            `yaml:"context_window" json:"context_window"`
	ContextProcessing map[string]any `yaml:"context_processing,omitempty" json:"context_processing,omitempty"`
	IntentAnalysis    map[string]any `yaml:"intent_analysis,omitempty" json:"intent_analysis,omitempty"`
	Persistence       map[string]any `yaml:"persistence,omitempty" json:"persistence,omitempty"`
}

// MultiAgentConfig represents the complete configuration for the multi-agent system
type MultiAgentConfig struct {
	Name            string            `yaml:"name" json:"name"`
	Description     string            `yaml:"description,omitempty" json:"description,omitempty"`
	Host            Host              `yaml:"host" json:"host"`
	Specialists     []*Specialist     `yaml:"specialists" json:"specialists"`
	PromptTemplates map[string]string `yaml:"prompt_templates,omitempty" json:"prompt_templates,omitempty"`
	Session         SessionConfig     `yaml:"session" json:"session"`
	MaxRounds       int               `yaml:"max_rounds" json:"max_rounds"`
}

// Validate validates the configuration
func (config *MultiAgentConfig) Validate() error {
	if config == nil {
		return errors.New("multi agent config is nil")
	}

	if len(config.Name) == 0 {
		return errors.New("multi agent config name is empty")
	}

	if config.Host.Model == nil {
		return errors.New("host model is not configured")
	}

	// 增加一个通用的specialist, 用于处理通用任务
	config.Specialists = append(config.Specialists, &Specialist{
		Name:         generalSpecialistNodeKey,
		IntendedUse:  "General tasks",
		ChatModel:    config.Host.Model,
		SystemPrompt: "You are a general specialist, you can handle any tasks.",
	})

	for i, specialist := range config.Specialists {
		if len(specialist.Name) == 0 {
			return fmt.Errorf("specialist %d has empty name", i)
		}

		if len(specialist.IntendedUse) == 0 {
			return fmt.Errorf("specialist %s has empty intended use", specialist.Name)
		}

		if specialist.ChatModel == nil && (specialist.Invokable == nil || specialist.Streamable == nil) && specialist.ReactAgent == nil {
			return fmt.Errorf("specialist %s has no model, invokable, streamable, or react agent configured", specialist.Name)
		}
	}

	return nil
}

// GetDefaultConfig returns a default configuration
func GetDefaultConfig(chatModel model.ToolCallingChatModel) *MultiAgentConfig {
	return &MultiAgentConfig{
		Name:        "Multi-agent",
		Description: "Multi-Agent System with ReAct thinking and task planning",
		Host: Host{
			Model:        chatModel,
			SystemPrompt: "You are an intelligent host agent responsible for analyzing tasks and coordinating specialist agents.",
			Thinking: ThinkingConfig{
				MaxSteps:           5,
				Timeout:            2 * time.Minute,
				EnableDeepThink:    true,
				ComplexityAnalysis: true,
			},
			Planning: PlanningConfig{
				MaxSteps:           10,
				Timeout:            3 * time.Minute,
				EnableDynamicPlan:  true,
				DependencyAnalysis: true,
			},
		},
		Specialists: []*Specialist{
			{
				Name:         "research_specialist",
				IntendedUse:  "Research and information gathering tasks",
				ChatModel:    chatModel,
				SystemPrompt: "You are a research specialist focused on gathering and analyzing information.",
			},
			{
				Name:         "code_specialist",
				IntendedUse:  "Code generation, analysis, and debugging tasks",
				ChatModel:    chatModel,
				SystemPrompt: "You are a code specialist focused on software development tasks.",
			},
			{
				Name:         "analysis_specialist",
				IntendedUse:  "Data analysis and reasoning tasks",
				ChatModel:    chatModel,
				SystemPrompt: "You are an analysis specialist focused on data analysis and logical reasoning.",
			},
		},
		Session: SessionConfig{
			HistoryLength: 20,
			ContextWindow: 4000,
		},
	}
}
