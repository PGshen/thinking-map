package enhanced

import (
	"context"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent"
	"github.com/cloudwego/eino/schema"
	ub "github.com/cloudwego/eino/utils/callbacks"
)

// MessageHandler 通用消息处理器接口
type MessageHandler interface {
	OnMessage(ctx context.Context, message *schema.Message) (context.Context, error)
	OnStreamMessage(ctx context.Context, message *schema.StreamReader[*schema.Message]) (context.Context, error)
}

// createMessageHandlerOption 创建通用的消息处理器option
func createMessageHandlerOption(handler MessageHandler, nodeKey string) agent.AgentOption {
	cmHandler := &ub.ModelCallbackHandler{
		OnEnd: func(ctx context.Context, runInfo *callbacks.RunInfo, output *model.CallbackOutput) context.Context {
			ctx, _ = handler.OnMessage(ctx, output.Message)
			return ctx
		},
		OnEndWithStreamOutput: func(ctx context.Context, runInfo *callbacks.RunInfo, output *schema.StreamReader[*model.CallbackOutput]) context.Context {
			c := func(output *model.CallbackOutput) (*schema.Message, error) {
				return output.Message, nil
			}
			s := schema.StreamReaderWithConvert(output, c)
			ctx, _ = handler.OnStreamMessage(ctx, s)
			return ctx
		},
	}
	cb := ub.NewHandlerHelper().ChatModel(cmHandler).Handler()
	option := agent.WithComposeOptions(compose.WithCallbacks(cb).DesignateNode(nodeKey))
	return option
}

// WithConversationAnalyzer 为对话分析节点添加消息处理器
func WithConversationAnalyzer(handler MessageHandler) agent.AgentOption {
	return createMessageHandlerOption(handler, conversationAnalyzerNodeKey)
}

// WithDirectAnswerHandler 为直接回答节点添加消息处理器
func WithDirectAnswerHandler(handler MessageHandler) agent.AgentOption {
	return createMessageHandlerOption(handler, directAnswerNodeKey)
}

// WithPlanCreationHandler 为计划创建节点添加消息处理器
func WithPlanCreationHandler(handler MessageHandler) agent.AgentOption {
	return createMessageHandlerOption(handler, planCreationNodeKey)
}

// WithFeedbackProcessorHandler 为反馈处理节点添加消息处理器
func WithFeedbackProcessorHandler(handler MessageHandler) agent.AgentOption {
	return createMessageHandlerOption(handler, feedbackProcessorNodeKey)
}

// WithPlanUpdateHandler 为计划更新节点添加消息处理器
func WithPlanUpdateHandler(handler MessageHandler) agent.AgentOption {
	return createMessageHandlerOption(handler, planUpdateNodeKey)
}

// WithFinalAnswerHandler 为最终回答节点添加消息处理器
func WithFinalAnswerHandler(handler MessageHandler) agent.AgentOption {
	return createMessageHandlerOption(handler, finalAnswerNodeKey)
}

// WithPlanExecutionHandler 为计划执行节点添加消息处理器
// 注意：Lambda节点的回调处理需要根据具体实现调整
func WithPlanExecutionHandler(handler MessageHandler) agent.AgentOption {
	return createMessageHandlerOption(handler, planExecutionNodeKey)
}

// WithResultCollectorHandler 为结果收集节点添加消息处理器
// 注意：Lambda节点的回调处理需要根据具体实现调整
func WithResultCollectorHandler(handler MessageHandler) agent.AgentOption {
	return createMessageHandlerOption(handler, resultCollectorNodeKey)
}

// WithSpecialistHandler 为指定专家节点添加消息处理器
func WithSpecialistHandler(specialistName string, handler MessageHandler) agent.AgentOption {
	return createMessageHandlerOption(handler, specialistName)
}
