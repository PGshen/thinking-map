package llmmodel

import (
	"context"

	"github.com/cloudwego/eino-ext/components/model/openai"
	openai2 "github.com/cloudwego/eino-ext/libs/acl/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/spf13/viper"
)

func DefaultOpenAIModelConfig(ctx context.Context) (*openai.ChatModelConfig, error) {
	config := &openai.ChatModelConfig{
		APIKey:  viper.GetString("llm.openai.api_key"),
		Model:   viper.GetString("llm.openai.model"),
		BaseURL: viper.GetString("llm.openai.base_url"),
	}
	return config, nil
}

func NewOpenAIModel(ctx context.Context, responseFormat *openai2.ChatCompletionResponseFormat) (cm model.ToolCallingChatModel, err error) {
	cfg, _ := DefaultOpenAIModelConfig(ctx)
	if responseFormat != nil {
		cfg.ResponseFormat = responseFormat
	}
	cm, err = openai.NewChatModel(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return cm, nil
}
