package multiagent

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/PGshen/thinking-map/server/internal/agent/base"
	"github.com/PGshen/thinking-map/server/internal/pkg/logger"
	"github.com/PGshen/thinking-map/server/internal/pkg/utils"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	ub "github.com/cloudwego/eino/utils/callbacks"
	"go.uber.org/zap"
)

// MessageHandler 通用消息处理器接口
type MessageHandler interface {
	OnMessage(ctx context.Context, message *schema.Message) (context.Context, error)
	OnStreamMessage(ctx context.Context, message *schema.StreamReader[*schema.Message]) (context.Context, error)
}

// createMessageHandlerOption 创建通用的消息处理器option
func createMessageHandlerOption(handler MessageHandler, nodeKey ...string) base.AgentOption {
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
	option := base.WithComposeOptions(compose.WithCallbacks(cb).DesignateNodeWithPath(compose.NewNodePath(nodeKey...)))
	return option
}

// WithConversationAnalyzer 为对话分析节点添加消息处理器
func WithConversationAnalyzer(handler MessageHandler) base.AgentOption {
	return createMessageHandlerOption(handler, conversationAnalyzerNodeKey)
}

// WithDirectAnswerHandler 为直接回答节点添加消息处理器
func WithDirectAnswerHandler(handler MessageHandler) base.AgentOption {
	return createMessageHandlerOption(handler, directAnswerNodeKey)
}

// WithPlanCreationHandler 为计划创建节点添加消息处理器
func WithPlanCreationHandler(handler MessageHandler) base.AgentOption {
	return createMessageHandlerOption(handler, planCreationNodeKey)
}

// WithFeedbackProcessorHandler 为反馈处理节点添加消息处理器
func WithFeedbackProcessorHandler(handler MessageHandler) base.AgentOption {
	return createMessageHandlerOption(handler, feedbackProcessorNodeKey)
}

// WithPlanUpdateHandler 为计划更新节点添加消息处理器
func WithPlanUpdateHandler(handler MessageHandler) base.AgentOption {
	return createMessageHandlerOption(handler, planUpdateNodeKey)
}

// WithFinalAnswerHandler 为最终回答节点添加消息处理器
func WithFinalAnswerHandler(handler MessageHandler) base.AgentOption {
	return createMessageHandlerOption(handler, finalAnswerNodeKey)
}

// WithSpecialistHandler 为指定专家节点添加消息处理器
func WithSpecialistHandler(specialistName string, handler MessageHandler) base.AgentOption {
	return createMessageHandlerOption(handler, specialistName, "reasoning")
}

type PlanHandler interface {
	OnPlanStepCreate(ctx context.Context, plan *TaskPlan, step *PlanStep) (context.Context, error)
	OnPlanStepUpdate(ctx context.Context, plan *TaskPlan, step *PlanStep) (context.Context, error)
	OnPlanStepStatusUpdate(ctx context.Context, plan *TaskPlan, step *PlanStep) (context.Context, error)
	OnPlanStepDelete(ctx context.Context, plan *TaskPlan, step *PlanStep) (context.Context, error)
	OnPlanStepEnd(ctx context.Context, plan *TaskPlan) (context.Context, error)
}

func WithPlanHandler(handler PlanHandler) base.AgentOption {
	// 创建一个基本的 TaskPlan 实例用于流式解析
	plan := &TaskPlan{
		ID:          "temp_plan",
		Version:     1,
		Name:        "Streaming Plan",
		Description: "Plan being parsed from stream",
		Status:      ExecutionStatusPlanning,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Steps:       make([]*PlanStep, 0),
	}
	// planCreation节点
	cmHandler := &ub.ModelCallbackHandler{
		OnEndWithStreamOutput: func(ctx context.Context, runInfo *callbacks.RunInfo, output *schema.StreamReader[*model.CallbackOutput]) context.Context {
			// 实时解析流，转换为PlanStep
			c := func(output *model.CallbackOutput) (*schema.Message, error) {
				return output.Message, nil
			}
			s := schema.StreamReaderWithConvert(output, c)

			// 实现流式JSON解析，提取计划步骤
			ctx = processPlanStepsStream(ctx, s, handler, plan)

			return ctx
		},
	}
	// planExecution节点
	peHandler := callbacks.NewHandlerBuilder().OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
		// 计划执行结束
		compose.ProcessState(ctx, func(ctx context.Context, state *MultiAgentState) error {
			var currentStep *PlanStep
			for _, step := range state.CurrentPlan.Steps {
				if step.ID == state.CurrentStep {
					currentStep = step
					break
				}
			}
			if currentStep == nil {
				return errors.New("current step not found")
			}
			ctx, _ = handler.OnPlanStepStatusUpdate(ctx, state.CurrentPlan, currentStep)
			return nil
		})
		return ctx
	}).Build()
	// planUpdate节点
	puHandler := callbacks.NewHandlerBuilder().OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
		// 计划更新结束
		compose.ProcessState(ctx, func(ctx context.Context, state *MultiAgentState) error {
			var currentStep *PlanStep
			for _, step := range state.CurrentPlan.Steps {
				if step.ID == state.CurrentStep {
					currentStep = step
					break
				}
			}
			if currentStep == nil {
				return errors.New("current step not found")
			}
			ctx, _ = handler.OnPlanStepUpdate(ctx, state.CurrentPlan, currentStep)
			return nil
		})
		return ctx
	}).Build()

	cb := ub.NewHandlerHelper().ChatModel(cmHandler).Handler()
	option := base.WithComposeOptions(compose.WithCallbacks(cb).DesignateNodeWithPath(compose.NewNodePath(planCreationNodeKey)),
		compose.WithCallbacks(peHandler).DesignateNodeWithPath(compose.NewNodePath(planExecutionNodeKey)),
		compose.WithCallbacks(puHandler).DesignateNodeWithPath(compose.NewNodePath(planUpdateNodeKey)),
	)
	return option
}

// processPlanStepsStream 处理计划步骤的流式解析
func processPlanStepsStream(ctx context.Context, sr *schema.StreamReader[*schema.Message], handler PlanHandler, plan *TaskPlan) context.Context {
	// 使用流式JSON解析器解析计划步骤
	matcher := utils.NewSimplePathMatcher()
	// 使用非实时非增量模式
	parser := utils.NewStreamingJsonParser(matcher, false, false)

	var createdSteps map[int]bool = make(map[int]bool) // 记录已创建的步骤

	// 封装通过path提取索引的方法
	extractStepIndex := func(path []interface{}) int {
		for _, segment := range path {
			if idx, isInt := segment.(int); isInt {
				return idx
			}
		}
		return -1
	}

	// 封装获取或创建步骤的方法
	ensureStep := func(stepIndex int) *PlanStep {
		// 扩展 plan.Steps 切片以容纳新索引
		for len(plan.Steps) <= stepIndex {
			plan.Steps = append(plan.Steps, &PlanStep{
				ID:     fmt.Sprintf("step_%d", len(plan.Steps)+1),
				Status: StepStatusPending,
			})
		}
		return plan.Steps[stepIndex]
	}

	// 封装处理步骤创建或更新的方法
	handleStepOperation := func(stepIndex int, step *PlanStep, plan *TaskPlan) {
		if !createdSteps[stepIndex] {
			// 第一次遇到这个步骤，执行创建操作
			handler.OnPlanStepCreate(ctx, plan, step)
			createdSteps[stepIndex] = true
		} else {
			// 已存在的步骤，执行更新操作
			handler.OnPlanStepUpdate(ctx, plan, step)
		}
	}
	// 注册路径匹配器来提取步骤名称字段
	matcher.On("steps[*].id", func(value interface{}, path []interface{}) {
		if str, ok := value.(string); ok {
			stepIndex := extractStepIndex(path)
			if stepIndex >= 0 {
				step := ensureStep(stepIndex)
				step.ID = str
				handleStepOperation(stepIndex, step, plan)
			}
		}
	})

	// 注册路径匹配器来提取步骤名称字段
	matcher.On("steps[*].name", func(value interface{}, path []interface{}) {
		if str, ok := value.(string); ok {
			stepIndex := extractStepIndex(path)
			if stepIndex >= 0 {
				step := ensureStep(stepIndex)
				step.Name = str
				handleStepOperation(stepIndex, step, plan)
			}
		}
	})

	// 注册其他步骤字段的匹配器
	matcher.On("steps[*].description", func(value interface{}, path []interface{}) {
		if str, ok := value.(string); ok {
			stepIndex := extractStepIndex(path)
			if stepIndex >= 0 {
				step := ensureStep(stepIndex)
				step.Description = str
				handleStepOperation(stepIndex, step, plan)
			}
		}
	})

	matcher.On("steps[*].assignedSpecialist", func(value interface{}, path []interface{}) {
		if str, ok := value.(string); ok {
			stepIndex := extractStepIndex(path)
			if stepIndex >= 0 {
				step := ensureStep(stepIndex)
				step.AssignedSpecialist = str
				handleStepOperation(stepIndex, step, plan)
			}
		}
	})

	matcher.On("steps[*].status", func(value interface{}, path []interface{}) {
		if status, ok := value.(StepStatus); ok {
			stepIndex := extractStepIndex(path)
			if stepIndex >= 0 {
				step := ensureStep(stepIndex)
				step.Status = status
				handleStepOperation(stepIndex, step, plan)
			}
		}
	})

	defer func() {
		sr.Close()
		// 通知本轮解析结束
		handler.OnPlanStepEnd(ctx, plan)
	}()

	// 处理流式数据
outer:
	for {
		select {
		case <-ctx.Done():
			logger.Info("context done", zap.Error(ctx.Err()))
			return ctx
		default:
			chunk, err := sr.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					break outer
				}
				logger.Error("receive stream chunk failed", zap.Error(err))
				break outer
			}
			if chunk != nil && chunk.Content != "" {
				if err := parser.Write(chunk.Content); err != nil {
					logger.Error("parse plan response failed", zap.Error(err))
				}
			}
		}
	}

	// 结束解析
	if err := parser.End(); err != nil {
		logger.Error("end plan parsing failed", zap.Error(err))
	}

	return ctx
}
