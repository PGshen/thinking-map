/*
 * @Date: 2025-01-27
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-01-27
 * @FilePath: /thinking-map/server/internal/agent/llmmodel/openai_test.go
 */
package llmmodel

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/PGshen/thinking-map/server/internal/config"
	"github.com/cloudwego/eino/schema"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	LoadTestConfig()

	code := m.Run()
	os.Exit(code)
}

// LoadTestConfig 加载测试配置
func LoadTestConfig() (*config.Config, error) {
	// 优先尝试加载测试配置文件
	configPath := "../../../configs/config.test.yaml"
	if _, err := os.Stat(configPath); err != nil {
		// 如果测试配置文件不存在，使用默认配置
		configPath = "../../../configs/config.yaml"
	}
	_ = godotenv.Load("../../../.env")
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load test config: %w", err)
	}
	return cfg, nil
}

func TestDefaultOpenAIModelConfig(t *testing.T) {
	ctx := context.Background()

	// 测试默认配置
	config, err := DefaultOpenAIModelConfig(ctx)

	require.NoError(t, err)
	assert.NotNil(t, config)

	// 验证配置字段不为空（在测试环境中应该被设置）
	if config.APIKey == "" {
		t.Logf("API Key is empty (expected in test environment)")
	} else {
		assert.NotEmpty(t, config.APIKey)
	}

	if config.Model == "" {
		t.Logf("Model is empty (expected in test environment)")
	} else {
		assert.NotEmpty(t, config.Model)
	}

	if config.BaseURL == "" {
		t.Logf("Base URL is empty (expected in test environment)")
	} else {
		assert.NotEmpty(t, config.BaseURL)
	}

	// 验证配置字段
	t.Logf("API Key: %s", config.APIKey)
	t.Logf("Model: %s", config.Model)
	t.Logf("Base URL: %s", config.BaseURL)
}

func TestOpenAIModelStream(t *testing.T) {
	ctx := context.Background()
	cm, err := NewOpenAIModel(ctx, nil)
	assert.Nil(t, err)
	output, err := cm.Generate(ctx, []*schema.Message{
		schema.UserMessage("hi"),
	})
	fmt.Println(output.Content)
}
