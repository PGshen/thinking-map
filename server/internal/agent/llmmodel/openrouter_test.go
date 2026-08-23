package llmmodel

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	openai2 "github.com/cloudwego/eino-ext/libs/acl/openai"
	"github.com/cloudwego/eino/schema"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setOpenRouterTestConfig(t *testing.T, values map[string]any) {
	t.Helper()

	previous := make(map[string]any, len(values))
	for key, value := range values {
		previous[key] = viper.Get(key)
		viper.Set(key, value)
	}
	t.Cleanup(func() {
		for key, value := range previous {
			viper.Set(key, value)
		}
	})
}

func TestDefaultOpenRouterModelConfig(t *testing.T) {
	setOpenRouterTestConfig(t, map[string]any{
		"llm.openrouter.api_key":  "test-api-key",
		"llm.openrouter.base_url": "https://openrouter.example/api/v1",
		"llm.openrouter.model":    "openai/gpt-4o-mini",
		"llm.openrouter.timeout":  "42s",
	})

	cfg, err := DefaultOpenRouterModelConfig(context.Background())

	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, "test-api-key", cfg.APIKey)
	assert.Equal(t, "https://openrouter.example/api/v1", cfg.BaseURL)
	assert.Equal(t, "openai/gpt-4o-mini", cfg.Model)
	assert.Equal(t, 42*time.Second, cfg.Timeout)
}

func TestDefaultOpenRouterModelConfigUsesDefaults(t *testing.T) {
	setOpenRouterTestConfig(t, map[string]any{
		"llm.openrouter.api_key":  "test-api-key",
		"llm.openrouter.base_url": "",
		"llm.openrouter.model":    "openai/gpt-4o-mini",
		"llm.openrouter.timeout":  0,
	})

	cfg, err := DefaultOpenRouterModelConfig(context.Background())

	require.NoError(t, err)
	assert.Equal(t, defaultOpenRouterBaseURL, cfg.BaseURL)
	assert.Equal(t, defaultOpenRouterTimeout, cfg.Timeout)
}

func TestDefaultOpenRouterModelConfigRequiresCredentialsAndModel(t *testing.T) {
	tests := []struct {
		name    string
		apiKey  string
		model   string
		wantErr string
	}{
		{name: "missing API key", model: "openai/gpt-4o-mini", wantErr: "llm.openrouter.api_key is required"},
		{name: "missing model", apiKey: "test-api-key", wantErr: "llm.openrouter.model is required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setOpenRouterTestConfig(t, map[string]any{
				"llm.openrouter.api_key": tt.apiKey,
				"llm.openrouter.model":   tt.model,
			})

			cfg, err := DefaultOpenRouterModelConfig(context.Background())

			assert.Nil(t, cfg)
			require.EqualError(t, err, tt.wantErr)
		})
	}
}

func TestOpenRouterModelGenerate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/chat/completions", r.URL.Path)
		assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))

		var request struct {
			Model          string                                `json:"model"`
			ResponseFormat *openai2.ChatCompletionResponseFormat `json:"response_format"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&request))
		assert.Equal(t, "openai/gpt-4o-mini", request.Model)
		require.NotNil(t, request.ResponseFormat)
		assert.Equal(t, openai2.ChatCompletionResponseFormatTypeJSONObject, request.ResponseFormat.Type)

		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(`{
			"id":"test-completion",
			"object":"chat.completion",
			"created":1,
			"model":"openai/gpt-4o-mini",
			"choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],
			"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}
		}`))
		require.NoError(t, err)
	}))
	t.Cleanup(server.Close)

	setOpenRouterTestConfig(t, map[string]any{
		"llm.provider":            ProviderOpenRouter,
		"llm.openrouter.api_key":  "test-api-key",
		"llm.openrouter.base_url": server.URL + "/api/v1",
		"llm.openrouter.model":    "openai/gpt-4o-mini",
		"llm.openrouter.timeout":  "5s",
	})
	responseFormat := &openai2.ChatCompletionResponseFormat{
		Type: openai2.ChatCompletionResponseFormatTypeJSONObject,
	}

	chatModel, err := NewModel(context.Background(), responseFormat)
	require.NoError(t, err)

	response, err := chatModel.Generate(context.Background(), []*schema.Message{
		schema.UserMessage("hello"),
	})
	require.NoError(t, err)
	assert.Equal(t, "ok", response.Content)
}
