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
	"errors"
	"fmt"
	"time"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent"
	"github.com/cloudwego/eino/schema"
)

// ModelConfig represents model configuration
type ModelConfig struct {
	Provider   string         `yaml:"provider" json:"provider"`
	Model      string         `yaml:"model" json:"model"`
	Parameters map[string]any `yaml:"parameters,omitempty" json:"parameters,omitempty"`
}

// ThinkingConfig represents thinking configuration
type ThinkingConfig struct {
	MaxSteps         int           `yaml:"max_steps" json:"max_steps"`
	Timeout          time.Duration `yaml:"timeout" json:"timeout"`
	EnableDeepThink  bool          `yaml:"enable_deep_think" json:"enable_deep_think"`
	ComplexityAnalysis bool        `yaml:"complexity_analysis" json:"complexity_analysis"`
}

// PlanningConfig represents planning configuration
type PlanningConfig struct {
	MaxSteps           int           `yaml:"max_steps" json:"max_steps"`
	Timeout            time.Duration `yaml:"timeout" json:"timeout"`
	EnableDynamicPlan  bool          `yaml:"enable_dynamic_plan" json:"enable_dynamic_plan"`
	DependencyAnalysis bool          `yaml:"dependency_analysis" json:"dependency_analysis"`
}

// EnhancedHost represents the host agent configuration
type EnhancedHost struct {
	Model              ModelConfig                                                                    `yaml:"model" json:"model"`
	SystemPrompt       string                                                                         `yaml:"system_prompt" json:"system_prompt"`
	Prompts            map[string]string                                                              `yaml:"prompts,omitempty" json:"prompts,omitempty"`
	CallableComponents map[string]compose.Invoke[[]*schema.Message, *schema.Message, agent.AgentOption] `yaml:"-" json:"-"`
	Thinking           ThinkingConfig                                                                 `yaml:"thinking" json:"thinking"`
	Planning           PlanningConfig                                                                 `yaml:"planning" json:"planning"`
}

// SpecialistConfig represents specialist-specific configuration
type SpecialistConfig struct {
	Execution      map[string]any `yaml:"execution,omitempty" json:"execution,omitempty"`
	QualityControl map[string]any `yaml:"quality_control,omitempty" json:"quality_control,omitempty"`
	ContextHandling map[string]any `yaml:"context_handling,omitempty" json:"context_handling,omitempty"`
}

// EnhancedSpecialist represents a specialist agent configuration
type EnhancedSpecialist struct {
	Name               string                                                                         `yaml:"name" json:"name"`
	IntendedUse        string                                                                         `yaml:"intended_use" json:"intended_use"`
	Model              ModelConfig                                                                    `yaml:"model" json:"model"`
	SystemPrompt       string                                                                         `yaml:"system_prompt" json:"system_prompt"`
	Prompts            map[string]string                                                              `yaml:"prompts,omitempty" json:"prompts,omitempty"`
	CallableComponents map[string]compose.Invoke[[]*schema.Message, *schema.Message, agent.AgentOption] `yaml:"-" json:"-"`
	Config             SpecialistConfig                                                               `yaml:"config,omitempty" json:"config,omitempty"`
	Concurrency        int                                                                            `yaml:"concurrency" json:"concurrency"`
	Timeout            time.Duration                                                                  `yaml:"timeout" json:"timeout"`
}

// SessionConfig represents session management configuration
type SessionConfig struct {
	HistoryLength      int           `yaml:"history_length" json:"history_length"`
	ContextWindow      int           `yaml:"context_window" json:"context_window"`
	ContextProcessing  map[string]any `yaml:"context_processing,omitempty" json:"context_processing,omitempty"`
	IntentAnalysis     map[string]any `yaml:"intent_analysis,omitempty" json:"intent_analysis,omitempty"`
	Persistence        map[string]any `yaml:"persistence,omitempty" json:"persistence,omitempty"`
}

// PerformanceConfig represents performance configuration
type PerformanceConfig struct {
	Concurrency     map[string]int `yaml:"concurrency,omitempty" json:"concurrency,omitempty"`
	MemoryManagement map[string]any `yaml:"memory_management,omitempty" json:"memory_management,omitempty"`
	Caching         map[string]any `yaml:"caching,omitempty" json:"caching,omitempty"`
	Monitoring      map[string]any `yaml:"monitoring,omitempty" json:"monitoring,omitempty"`
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Level      string         `yaml:"level" json:"level"`
	Format     string         `yaml:"format" json:"format"`
	Output     []string       `yaml:"output" json:"output"`
	File       map[string]any `yaml:"file,omitempty" json:"file,omitempty"`
	SpecialLogs map[string]any `yaml:"special_logs,omitempty" json:"special_logs,omitempty"`
}

// ExecutionControlConfig represents execution control configuration
type ExecutionControlConfig struct {
	MaxRounds        int           `yaml:"max_rounds" json:"max_rounds"`
	Timeout          time.Duration `yaml:"timeout" json:"timeout"`
	RetryPolicy      map[string]any `yaml:"retry_policy,omitempty" json:"retry_policy,omitempty"`
	ErrorHandling    map[string]any `yaml:"error_handling,omitempty" json:"error_handling,omitempty"`
	GracefulShutdown bool          `yaml:"graceful_shutdown" json:"graceful_shutdown"`
}

// SystemConfig represents system-level configuration
type SystemConfig struct {
	Version     string         `yaml:"version" json:"version"`
	Environment string         `yaml:"environment" json:"environment"`
	DebugMode   bool           `yaml:"debug_mode" json:"debug_mode"`
	Metrics     map[string]any `yaml:"metrics,omitempty" json:"metrics,omitempty"`
	Tracing     map[string]any `yaml:"tracing,omitempty" json:"tracing,omitempty"`
}

// EnhancedMultiAgentConfig represents the complete configuration for the enhanced multi-agent system
type EnhancedMultiAgentConfig struct {
	Name             string                   `yaml:"name" json:"name"`
	Description      string                   `yaml:"description,omitempty" json:"description,omitempty"`
	Host             EnhancedHost             `yaml:"host" json:"host"`
	Specialists      []*EnhancedSpecialist    `yaml:"specialists" json:"specialists"`
	System           SystemConfig             `yaml:"system" json:"system"`
	ExecutionControl ExecutionControlConfig   `yaml:"execution_control" json:"execution_control"`
	PromptTemplates  map[string]string        `yaml:"prompt_templates,omitempty" json:"prompt_templates,omitempty"`
	Session          SessionConfig            `yaml:"session" json:"session"`
	Performance      PerformanceConfig        `yaml:"performance" json:"performance"`
	Logging          LoggingConfig            `yaml:"logging" json:"logging"`
}

// Validate validates the configuration
func (config *EnhancedMultiAgentConfig) Validate() error {
	if config == nil {
		return errors.New("enhanced multi agent config is nil")
	}

	if len(config.Name) == 0 {
		return errors.New("enhanced multi agent config name is empty")
	}

	if len(config.Host.Model.Model) == 0 {
		return errors.New("host model is not configured")
	}

	if len(config.Specialists) == 0 {
		return errors.New("no specialists configured")
	}

	for i, specialist := range config.Specialists {
		if len(specialist.Name) == 0 {
			return fmt.Errorf("specialist %d has empty name", i)
		}

		if len(specialist.IntendedUse) == 0 {
			return fmt.Errorf("specialist %s has empty intended use", specialist.Name)
		}

		if len(specialist.Model.Model) == 0 {
			return fmt.Errorf("specialist %s has no model configured", specialist.Name)
		}
	}

	if config.ExecutionControl.MaxRounds <= 0 {
		config.ExecutionControl.MaxRounds = 10 // default value
	}

	if config.ExecutionControl.Timeout <= 0 {
		config.ExecutionControl.Timeout = 5 * time.Minute // default value
	}

	return nil
}

// GetDefaultConfig returns a default configuration
func GetDefaultConfig() *EnhancedMultiAgentConfig {
	return &EnhancedMultiAgentConfig{
		Name:        "enhanced-multi-agent",
		Description: "Enhanced Multi-Agent System with ReAct thinking and task planning",
		Host: EnhancedHost{
			Model: ModelConfig{
				Provider: "openai",
				Model:    "gpt-4",
			},
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
		Specialists: []*EnhancedSpecialist{
			{
				Name:        "research_specialist",
				IntendedUse: "Research and information gathering tasks",
				Model: ModelConfig{
					Provider: "openai",
					Model:    "gpt-4",
				},
				SystemPrompt: "You are a research specialist focused on gathering and analyzing information.",
				Concurrency:  1,
				Timeout:      5 * time.Minute,
			},
			{
				Name:        "code_specialist",
				IntendedUse: "Code generation, analysis, and debugging tasks",
				Model: ModelConfig{
					Provider: "openai",
					Model:    "gpt-4",
				},
				SystemPrompt: "You are a code specialist focused on software development tasks.",
				Concurrency:  1,
				Timeout:      5 * time.Minute,
			},
			{
				Name:        "analysis_specialist",
				IntendedUse: "Data analysis and reasoning tasks",
				Model: ModelConfig{
					Provider: "openai",
					Model:    "gpt-4",
				},
				SystemPrompt: "You are an analysis specialist focused on data analysis and logical reasoning.",
				Concurrency:  1,
				Timeout:      5 * time.Minute,
			},
		},
		System: SystemConfig{
			Version:     "1.0.0",
			Environment: "development",
			DebugMode:   true,
		},
		ExecutionControl: ExecutionControlConfig{
			MaxRounds:        10,
			Timeout:          10 * time.Minute,
			GracefulShutdown: true,
		},
		Session: SessionConfig{
			HistoryLength: 20,
			ContextWindow: 4000,
		},
		Performance: PerformanceConfig{
			Concurrency: map[string]int{
				"max_concurrent_specialists": 3,
				"max_concurrent_steps":       5,
			},
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
			Output: []string{"stdout"},
		},
		PromptTemplates: map[string]string{
			"complexity_analysis": "Analyze the complexity of the following task: {{.task}}",
			"step_execution":      "Execute the following step: {{.step}}",
			"result_collection":   "Collect and summarize the following results: {{.results}}",
			"feedback_analysis":   "Analyze the feedback for the following execution: {{.execution}}",
		},
	}
}