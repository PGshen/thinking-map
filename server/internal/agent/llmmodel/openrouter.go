package llmmodel

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	openai2 "github.com/cloudwego/eino-ext/libs/acl/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/spf13/viper"
)

const (
	defaultOpenRouterBaseURL = "https://openrouter.ai/api/v1"
	defaultOpenRouterTimeout = 300 * time.Second
)

// DefaultOpenRouterModelConfig builds an OpenAI-compatible configuration for OpenRouter.
func DefaultOpenRouterModelConfig(ctx context.Context) (*openai.ChatModelConfig, error) {
	apiKey := viper.GetString("llm.openrouter.api_key")
	if apiKey == "" {
		return nil, fmt.Errorf("llm.openrouter.api_key is required")
	}

	modelName := viper.GetString("llm.openrouter.model")
	if modelName == "" {
		return nil, fmt.Errorf("llm.openrouter.model is required")
	}

	baseURL := viper.GetString("llm.openrouter.base_url")
	if baseURL == "" {
		baseURL = defaultOpenRouterBaseURL
	}

	timeout := viper.GetDuration("llm.openrouter.timeout")
	if timeout <= 0 {
		timeout = defaultOpenRouterTimeout
	}

	return &openai.ChatModelConfig{
		APIKey:  apiKey,
		Model:   modelName,
		BaseURL: baseURL,
		Timeout: timeout,
	}, nil
}

// NewOpenRouterModel creates a tool-calling model backed by OpenRouter's OpenAI-compatible API.
func NewOpenRouterModel(ctx context.Context, responseFormat *openai2.ChatCompletionResponseFormat) (model.ToolCallingChatModel, error) {
	cfg, err := DefaultOpenRouterModelConfig(ctx)
	if err != nil {
		return nil, err
	}
	if responseFormat != nil {
		cfg.ResponseFormat = responseFormat
	}

	return openai.NewChatModel(ctx, cfg)
}
