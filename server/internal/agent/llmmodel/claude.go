package llmmodel

import (
	"context"

	"github.com/cloudwego/eino-ext/components/model/claude"
	"github.com/cloudwego/eino/components/model"
	"github.com/spf13/viper"
)

func DefaultClaudeConfig() (*claude.Config, error) {
	config := &claude.Config{
		Model:  viper.GetString("llm.claude.model"),
		APIKey: viper.GetString("llm.claude.api_key"),
	}
	return config, nil
}

func NewClaudeModel(ctx context.Context, config *claude.Config) (cm model.ChatModel, err error) {
	if config == nil {
		config, err = DefaultClaudeConfig()
		if err != nil {
			return nil, err
		}
	}
	cm, err = claude.NewChatModel(ctx, config)
	if err != nil {
		return nil, err
	}
	return cm, nil
}
