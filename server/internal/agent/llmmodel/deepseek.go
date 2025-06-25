package model

import (
	"context"

	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino/components/model"
	"github.com/spf13/viper"
)

func DefaultDeepSeekModelConfig(ctx context.Context) (*deepseek.ChatModelConfig, error) {

	config := &deepseek.ChatModelConfig{
		APIKey:             viper.GetString("llm.deepseek.api_key"),
		Model:              viper.GetString("llm.deepseek.model"),
		MaxTokens:          8000,
		ResponseFormatType: deepseek.ResponseFormatTypeJSONObject,
	}
	return config, nil
}

func NewDeepSeekModel(ctx context.Context, responseFormat deepseek.ResponseFormatType) (cm model.ChatModel, err error) {
	cfg, _ := DefaultDeepSeekModelConfig(ctx)
	cfg.ResponseFormatType = responseFormat
	cm, err = deepseek.NewChatModel(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return cm, nil
}
