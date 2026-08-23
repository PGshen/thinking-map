package llmmodel

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewModelDefaultsToOpenAI(t *testing.T) {
	setOpenRouterTestConfig(t, map[string]any{
		"llm.provider":        "",
		"llm.openai.api_key":  "test-openai-key",
		"llm.openai.base_url": "https://openai.example/v1",
		"llm.openai.model":    "gpt-test",
	})

	chatModel, err := NewModel(context.Background(), nil)

	require.NoError(t, err)
	require.NotNil(t, chatModel)
}

func TestNewModelSelectsOpenRouter(t *testing.T) {
	setOpenRouterTestConfig(t, map[string]any{
		"llm.provider":            " OpenRouter ",
		"llm.openrouter.api_key":  "test-openrouter-key",
		"llm.openrouter.base_url": "https://openrouter.example/api/v1",
		"llm.openrouter.model":    "openai/gpt-test",
	})

	chatModel, err := NewModel(context.Background(), nil)

	require.NoError(t, err)
	require.NotNil(t, chatModel)
}

func TestNewModelRejectsUnsupportedProvider(t *testing.T) {
	setOpenRouterTestConfig(t, map[string]any{
		"llm.provider": "unknown",
	})

	chatModel, err := NewModel(context.Background(), nil)

	require.Nil(t, chatModel)
	require.EqualError(t, err, `unsupported llm provider "unknown" (supported providers: openai, openrouter)`)
}
