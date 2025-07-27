package react

import (
	"errors"
	"fmt"
)

// Error types for the agent
var (
	// Configuration errors
	ErrInvalidConfig   = errors.New("invalid agent configuration")
	ErrMissingModel    = errors.New("chat model is required")
	ErrMissingTools    = errors.New("at least one tool is required")
	ErrInvalidMaxSteps = errors.New("max steps must be greater than 0")

	// Execution errors
	ErrMaxStepsExceeded = errors.New("maximum execution steps exceeded")
	ErrExecutionFailed  = errors.New("agent execution failed")
	ErrContextCancelled = errors.New("execution context was cancelled")
	ErrTimeout          = errors.New("execution timeout")

	// Reasoning errors
	ErrReasoningFailed  = errors.New("reasoning step failed")
	ErrInvalidReasoning = errors.New("invalid reasoning format")
	ErrParsingFailed    = errors.New("failed to parse reasoning response")
	ErrEmptyResponse    = errors.New("empty response from model")

	// Tool errors
	ErrToolNotFound        = errors.New("tool not found")
	ErrToolExecutionFailed = errors.New("tool execution failed")
	ErrInvalidToolArgs     = errors.New("invalid tool arguments")
	ErrToolTimeout         = errors.New("tool execution timeout")

	// Decision errors
	ErrInvalidDecision = errors.New("invalid decision")
	ErrDecisionFailed  = errors.New("decision making failed")
	ErrNoValidAction   = errors.New("no valid action determined")

	// State errors
	ErrInvalidState      = errors.New("invalid agent state")
	ErrStateUpdateFailed = errors.New("failed to update agent state")
)

// AgentError represents a structured error with context
type AgentError struct {
	Type      string                 `json:"type"`
	Message   string                 `json:"message"`
	Component string                 `json:"component"`
	Context   map[string]interface{} `json:"context,omitempty"`
	Cause     error                  `json:"cause,omitempty"`
}

// Error implements the error interface
func (e *AgentError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s:%s] %s: %v", e.Component, e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s:%s] %s", e.Component, e.Type, e.Message)
}

// Unwrap returns the underlying cause
func (e *AgentError) Unwrap() error {
	return e.Cause
}

// NewAgentError creates a new AgentError
func NewAgentError(errorType, component, message string, cause error) *AgentError {
	return &AgentError{
		Type:      errorType,
		Component: component,
		Message:   message,
		Cause:     cause,
		Context:   make(map[string]interface{}),
	}
}

// WithContext adds context to the error
func (e *AgentError) WithContext(key string, value interface{}) *AgentError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// Error type constants
const (
	ErrorTypeConfig     = "config"
	ErrorTypeExecution  = "execution"
	ErrorTypeReasoning  = "reasoning"
	ErrorTypeTool       = "tool"
	ErrorTypeDecision   = "decision"
	ErrorTypeState      = "state"
	ErrorTypeTimeout    = "timeout"
	ErrorTypeValidation = "validation"
)

// Component constants
const (
	ComponentAgent     = "agent"
	ComponentReasoning = "reasoning"
	ComponentDecision  = "decision"
	ComponentTool      = "tool"
	ComponentState     = "state"
	ComponentMonitor   = "monitor"
	ComponentConfig    = "config"
)

// Helper functions for creating specific errors

// NewConfigError creates a configuration error
func NewConfigError(message string, cause error) *AgentError {
	return NewAgentError(ErrorTypeConfig, ComponentConfig, message, cause)
}

// NewExecutionError creates an execution error
func NewExecutionError(message string, cause error) *AgentError {
	return NewAgentError(ErrorTypeExecution, ComponentAgent, message, cause)
}

// NewReasoningError creates a reasoning error
func NewReasoningError(message string, cause error) *AgentError {
	return NewAgentError(ErrorTypeReasoning, ComponentReasoning, message, cause)
}

// NewToolError creates a tool error
func NewToolError(message string, cause error) *AgentError {
	return NewAgentError(ErrorTypeTool, ComponentTool, message, cause)
}

// NewDecisionError creates a decision error
func NewDecisionError(message string, cause error) *AgentError {
	return NewAgentError(ErrorTypeDecision, ComponentDecision, message, cause)
}

// NewStateError creates a state error
func NewStateError(message string, cause error) *AgentError {
	return NewAgentError(ErrorTypeState, ComponentState, message, cause)
}

// NewTimeoutError creates a timeout error
func NewTimeoutError(component, message string, cause error) *AgentError {
	return NewAgentError(ErrorTypeTimeout, component, message, cause)
}

// NewValidationError creates a validation error
func NewValidationError(component, message string, cause error) *AgentError {
	return NewAgentError(ErrorTypeValidation, component, message, cause)
}

// IsAgentError checks if an error is an AgentError
func IsAgentError(err error) bool {
	_, ok := err.(*AgentError)
	return ok
}

// GetAgentError extracts AgentError from error chain
func GetAgentError(err error) *AgentError {
	var agentErr *AgentError
	if errors.As(err, &agentErr) {
		return agentErr
	}
	return nil
}

// IsErrorType checks if an error is of a specific type
func IsErrorType(err error, errorType string) bool {
	agentErr := GetAgentError(err)
	if agentErr == nil {
		return false
	}
	return agentErr.Type == errorType
}

// IsErrorComponent checks if an error is from a specific component
func IsErrorComponent(err error, component string) bool {
	agentErr := GetAgentError(err)
	if agentErr == nil {
		return false
	}
	return agentErr.Component == component
}

// WrapError wraps an existing error with AgentError
func WrapError(err error, errorType, component, message string) *AgentError {
	if err == nil {
		return nil
	}
	return NewAgentError(errorType, component, message, err)
}

// ErrorRecovery provides error recovery strategies
type ErrorRecovery struct {
	MaxRetries  int
	RetryDelay  int // milliseconds
	Recoverable map[string]bool
}

// NewErrorRecovery creates a new error recovery configuration
func NewErrorRecovery() *ErrorRecovery {
	return &ErrorRecovery{
		MaxRetries: 3,
		RetryDelay: 1000,
		Recoverable: map[string]bool{
			ErrorTypeTimeout:    true,
			ErrorTypeTool:       true,
			ErrorTypeReasoning:  true,
			ErrorTypeConfig:     false,
			ErrorTypeValidation: false,
		},
	}
}

// IsRecoverable checks if an error type is recoverable
func (er *ErrorRecovery) IsRecoverable(errorType string) bool {
	recoverable, exists := er.Recoverable[errorType]
	return exists && recoverable
}

// ShouldRetry determines if an operation should be retried based on the error
func (er *ErrorRecovery) ShouldRetry(err error, attemptCount int) bool {
	if attemptCount >= er.MaxRetries {
		return false
	}

	agentErr := GetAgentError(err)
	if agentErr == nil {
		return false
	}

	return er.IsRecoverable(agentErr.Type)
}

// GetRetryDelay returns the delay before retrying
func (er *ErrorRecovery) GetRetryDelay(attemptCount int) int {
	// Exponential backoff
	delay := er.RetryDelay
	for i := 0; i < attemptCount; i++ {
		delay *= 2
	}
	return delay
}
