package llmmodel

import (
	"context"
	"fmt"
	"strings"

	openai2 "github.com/cloudwego/eino-ext/libs/acl/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/spf13/viper"
)

const (
	ProviderOpenAI     = "openai"
	ProviderOpenRouter = "openrouter"
)

// NewModel creates the configured LLM provider. OpenAI remains the default for
// backward compatibility when llm.provider is not configured.
func NewModel(ctx context.Context, responseFormat *openai2.ChatCompletionResponseFormat) (model.ToolCallingChatModel, error) {
	provider := strings.ToLower(strings.TrimSpace(viper.GetString("llm.provider")))
	if provider == "" {
		provider = ProviderOpenAI
	}

	switch provider {
	case ProviderOpenAI:
		return NewOpenAIModel(ctx, responseFormat)
	case ProviderOpenRouter:
		return NewOpenRouterModel(ctx, responseFormat)
	default:
		return nil, fmt.Errorf(
			"unsupported llm provider %q (supported providers: %s, %s)",
			provider,
			ProviderOpenAI,
			ProviderOpenRouter,
		)
	}
}
