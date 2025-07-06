package understanding

import (
	"context"

	"github.com/PGshen/thinking-map/server/internal/agent/llmmodel"
	"github.com/PGshen/thinking-map/server/internal/model/dto"
	"github.com/PGshen/thinking-map/server/internal/pkg/logger"
	"github.com/PGshen/thinking-map/server/internal/pkg/utils"
	"github.com/cloudwego/eino-ext/libs/acl/openai"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/getkin/kin-openapi/openapi3gen"
	"go.uber.org/zap"
)

// 问题理解&意图识别
func BuildUnderstandingAgent(ctx context.Context) (r compose.Runnable[[]*schema.Message, *schema.Message], err error) {
	generator := openapi3gen.NewGenerator(
		openapi3gen.UseAllExportedFields(),
	)

	// 从结构体生成 Schema
	understandingSchema, err := generator.NewSchemaRefForValue(&dto.UnderstandingResponse{}, nil)
	if err != nil {
		logger.Fatal("NewSchemaRefForValue failed", zap.Error(err))
		return nil, err
	}
	utils.MakeAllFieldsRequired(understandingSchema.Value)
	responseFormat := &openai.ChatCompletionResponseFormat{
		Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
		JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
			Name:        "understanding",
			Description: "问题理解&意图识别的json结果",
			Strict:      true,
			Schema:      understandingSchema.Value,
		},
	}
	cm, err := llmmodel.NewOpenAIModel(ctx, responseFormat)
	if err != nil {
		return nil, err
	}
	chain := compose.NewChain[[]*schema.Message, *schema.Message]()
	chain.AppendLambda(compose.InvokableLambdaWithOption(func(ctx context.Context, input []*schema.Message, opts ...any) (output []*schema.Message, err error) {
		systemMsg := schema.SystemMessage(systemPrompt)
		return append([]*schema.Message{systemMsg}, input...), nil
	})).AppendChatModel(cm)
	return chain.Compile(ctx, compose.WithGraphName("understanding"))
}
