package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/PGshen/thinking-map/server/internal/agent/base"
	"github.com/PGshen/thinking-map/server/internal/agent/base/multiagent"
	"github.com/PGshen/thinking-map/server/internal/agent/callback"
	"github.com/PGshen/thinking-map/server/internal/agent/conclusion"
	conclusionv2 "github.com/PGshen/thinking-map/server/internal/agent/conclusionv2"
	"github.com/PGshen/thinking-map/server/internal/global"
	"github.com/PGshen/thinking-map/server/internal/model"
	"github.com/PGshen/thinking-map/server/internal/model/dto"
	"github.com/PGshen/thinking-map/server/internal/pkg/logger"
	"github.com/PGshen/thinking-map/server/internal/pkg/sse"
	"github.com/PGshen/thinking-map/server/internal/pkg/utils"
	"github.com/PGshen/thinking-map/server/internal/repository"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ConclusionService 处理结论生成相关的业务逻辑
type ConclusionService struct {
	contextManager *ContextManager
	msgManager     *global.MessageManager
	nodeRepo       repository.ThinkingNode
}

// NewConclusionService 创建一个新的结论服务
func NewConclusionService(contextManager *ContextManager, nodeRepo repository.ThinkingNode) *ConclusionService {
	return &ConclusionService{
		contextManager: contextManager,
		msgManager:     global.GetMessageManager(),
		nodeRepo:       nodeRepo,
	}
}

// GenerateConclusion 处理结论生成请求
func (s *ConclusionService) GenerateConclusion(ctx *gin.Context, req dto.ConclusionRequest) (err error) {
	node, err := s.nodeRepo.FindByID(ctx, req.NodeID)
	if err != nil {
		return
	}
	initConclusion := node.Conclusion.Content
	lastMsgID := node.Decomposition.LastMessageID
	userID := ctx.GetString("user_id")
	
	// 1. 构建上下文消息
	contextInfo, err := s.contextManager.GetNodeContextWithConversation(ctx, req.NodeID, lastMsgID)
	if err != nil {
		return
	}
	// 将mapID, nodeID传入到ctx
	ctx.Set("mapID", contextInfo.MapInfo.ID)
	ctx.Set("nodeID", req.NodeID)

	// 2. 构建用户消息
	userMessage := &conclusion.UserMessage{
		Reference:   req.Reference,
		Instruction: req.Instruction,
		Conclusion:  initConclusion,
	}
	
	return s.Conclude(ctx, contextInfo, userMessage, userID)
}

// Conclude 生成节点结论
func (s *ConclusionService) Conclude(ctx *gin.Context, contextInfo *ContextInfo, userMessage *conclusion.UserMessage, userID string) (err error) {
	defer func() {
		if err != nil {
			logger.Error("Generate conclusion failed", zap.Error(err))
		}
	}()

	// 创建消息处理器
	reasoningMessageHandler := &conclusionReasoningMessageHandler{
		mapID:      contextInfo.MapInfo.ID,
		nodeID:     contextInfo.NodeInfo.ID,
		userID:     userID,
		msgManager: s.msgManager,
	}
	generalMessageHandler := &conclusionGeneralMessageHandler{
		mapID:      contextInfo.MapInfo.ID,
		nodeID:     contextInfo.NodeInfo.ID,
		userID:     userID,
		msgManager: s.msgManager,
	}
	planMessageHandler := &conclusionPlanMessageHandler{
		mapID:      contextInfo.MapInfo.ID,
		nodeID:     contextInfo.NodeInfo.ID,
		userID:     userID,
		messageID:  uuid.NewString(),
		msgManager: s.msgManager,
	}

	// 将mapID, nodeID保存至ctx, 工具调用时会用到
	ctx.Set("mapID", contextInfo.MapInfo.ID)
	ctx.Set("nodeID", contextInfo.NodeInfo.ID)

	// 调用结论生成Agent
	agent, err := conclusionv2.BuildConclusionAgent(ctx)
	if err != nil {
		return
	}

	// 将UserMessage转换为[]*schema.Message
	messages := userMessage.History
	if len(messages) == 0 {
		// 如果没有历史消息，创建一个用户消息
		messages = []*schema.Message{
			{
				Role:    schema.User,
				Content: userMessage.Instruction,
			},
		}
	}

	// 执行结论生成
	options := []base.AgentOption{
		multiagent.WithDirectAnswerHandler(reasoningMessageHandler),
		multiagent.WithFinalAnswerHandler(generalMessageHandler),
		multiagent.WithPlanHandler(planMessageHandler),
		multiagent.WithSpecialistHandler("ConclusionGenerationAgent", reasoningMessageHandler),
		multiagent.WithSpecialistHandler("ConclusionOptimizationAgent", reasoningMessageHandler),
		multiagent.WithSpecialistHandler("general_specialist", reasoningMessageHandler),
	}
	opts := base.GetComposeOptions(options...)
	opts = append(opts, compose.WithCallbacks(callback.LogCbHandler))
	
	sr, err := agent.Stream(ctx, messages, opts...)
	if err != nil {
		return
	}
	
	for {
		chunk, err := sr.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}
		fmt.Printf("%s", chunk.Content)
	}

	return
}

// 结论推理消息处理器
type conclusionReasoningMessageHandler struct {
	mapID      string
	nodeID     string
	userID     string
	msgManager *global.MessageManager
}

func (m *conclusionReasoningMessageHandler) OnMessage(ctx context.Context, message *schema.Message) (context.Context, error) {
	return ctx, nil
}

func (m *conclusionReasoningMessageHandler) OnStreamMessage(ctx context.Context, sr *schema.StreamReader[*schema.Message]) (context.Context, error) {
	thoughtMessageID := uuid.NewString()
	finalAnswerMessageID := uuid.NewString()
	sseBroker := global.GetBroker()
	
	// 使用流式JSON解析器解析ReasoningOutput
	matcher := utils.NewSimplePathMatcher()
	parser := utils.NewStreamingJsonParser(matcher, true, true)

	var thought, finalAnswer strings.Builder

	// 注册路径匹配器来提取thought和final_answer字段
	matcher.On("thought", func(value interface{}, path []interface{}) {
		if str, ok := value.(string); ok {
			sseBroker.PublishToSession(m.mapID, sse.Event{
				ID:   m.nodeID,
				Type: dto.MessageThoughtEventType,
				Data: dto.MessageThoughtEvent{
					NodeID:    m.nodeID,
					MessageID: thoughtMessageID,
					Message:   str,
					Mode:      "append",
				},
			})
			thought.WriteString(str)
		}
	})
	matcher.On("final_answer", func(value interface{}, path []interface{}) {
		if str, ok := value.(string); ok {
			sseBroker.PublishToSession(m.mapID, sse.Event{
				ID:   m.nodeID,
				Type: dto.MessageTextEventType,
				Data: dto.MessageTextEvent{
					NodeID:    m.nodeID,
					MessageID: finalAnswerMessageID,
					Message:   str,
					Mode:      "append",
				},
			})
			finalAnswer.WriteString(str)
		}
	})

	defer func() {
		sr.Close()
		if len(thought.String()) > 0 {
			msgReq := dto.CreateMessageRequest{
				ID:          thoughtMessageID,
				UserID:      m.userID,
				MessageType: model.MsgTypeThought,
				Role:        schema.Assistant,
				Content: model.MessageContent{
					Thought: thought.String(),
				},
			}
			_, err2 := m.msgManager.SaveConclusionMessage(ctx, m.nodeID, msgReq)
			if err2 != nil {
				logger.Error("create thought message failed", zap.Error(err2))
				return
			}
		}
		if len(finalAnswer.String()) > 0 {
			msgReq := dto.CreateMessageRequest{
				ID:          finalAnswerMessageID,
				UserID:      m.userID,
				MessageType: model.MsgTypeText,
				Role:        schema.Assistant,
				Content: model.MessageContent{
					Text: finalAnswer.String(),
				},
			}
			_, err2 := m.msgManager.SaveConclusionMessage(ctx, m.nodeID, msgReq)
			if err2 != nil {
				logger.Error("create final answer message failed", zap.Error(err2))
				return
			}
		}
	}()

outer:
	for {
		select {
		case <-ctx.Done():
			fmt.Println("context done", ctx.Err())
			return ctx, nil
		default:
			chunk, err2 := sr.Recv()
			if err2 != nil {
				if errors.Is(err2, io.EOF) {
					fmt.Println()
					break outer
				}
			}
			if chunk != nil {
				if err := parser.Write(chunk.Content); err != nil {
					logger.Error("parse reasoning response failed", zap.Error(err))
				}
			}
		}
	}
	return ctx, nil
}

// 结论普通消息处理器
type conclusionGeneralMessageHandler struct {
	mapID      string
	nodeID     string
	userID     string
	msgManager *global.MessageManager
}

func (m *conclusionGeneralMessageHandler) OnMessage(ctx context.Context, message *schema.Message) (context.Context, error) {
	return ctx, nil
}

func (m *conclusionGeneralMessageHandler) OnStreamMessage(ctx context.Context, sr *schema.StreamReader[*schema.Message]) (context.Context, error) {
	messageID := uuid.NewString()
	sseBroker := global.GetBroker()
	fullMsgs := make([]*schema.Message, 0)
	
	defer func() {
		sr.Close()
		fullMsg, err := schema.ConcatMessages(fullMsgs)
		if err != nil {
			logger.Warn("concat message failed", zap.Error(err))
			return
		}
		fullMsg.Content = strings.ReplaceAll(fullMsg.Content, "&nbsp;", " ")
		msgReq := dto.CreateMessageRequest{
			ID:          messageID,
			UserID:      m.userID,
			MessageType: model.MsgTypeText,
			Role:        schema.Assistant,
			Content: model.MessageContent{
				Text: fullMsg.Content,
			},
		}
		_, err2 := m.msgManager.SaveConclusionMessage(ctx, m.nodeID, msgReq)
		if err2 != nil {
			logger.Error("create message failed", zap.Error(err2))
			return
		}
	}()

outer:
	for {
		select {
		case <-ctx.Done():
			fmt.Println("context done", ctx.Err())
			return ctx, nil
		default:
			chunk, err2 := sr.Recv()
			if err2 != nil {
				if errors.Is(err2, io.EOF) {
					fmt.Println()
					break outer
				}
			}
			fmt.Print(chunk.Content)
			// sse 事件
			sseBroker.PublishToSession(m.mapID, sse.Event{
				ID:   m.nodeID,
				Type: dto.MessageTextEventType,
				Data: dto.MessageTextEvent{
					NodeID:    m.nodeID,
					MessageID: messageID,
					Message:   chunk.Content,
					Mode:      "append",
				},
			})
			fullMsgs = append(fullMsgs, chunk)
		}
	}
	return ctx, nil
}

// 结论计划消息处理器
type conclusionPlanMessageHandler struct {
	mapID      string
	nodeID     string
	userID     string
	messageID  string
	msgManager *global.MessageManager
}

func (p *conclusionPlanMessageHandler) OnPlanStepCreate(ctx context.Context, plan *multiagent.TaskPlan, step *multiagent.PlanStep) (context.Context, error) {
	p.sendPlanEvent(ctx, *plan, false)
	return ctx, nil
}

func (p *conclusionPlanMessageHandler) OnPlanStepUpdate(ctx context.Context, plan *multiagent.TaskPlan, step *multiagent.PlanStep) (context.Context, error) {
	p.sendPlanEvent(ctx, *plan, false)
	return ctx, nil
}

func (p *conclusionPlanMessageHandler) OnPlanStepStatusUpdate(ctx context.Context, plan *multiagent.TaskPlan, step *multiagent.PlanStep) (context.Context, error) {
	p.sendPlanEvent(ctx, *plan, false)
	// 有状态更新，也保存一下消息
	err := p.savePlanMessage(ctx, *plan)
	if err != nil {
		return ctx, err
	}
	//  本次消息结束，更新messageID
	p.messageID = uuid.NewString()
	return ctx, nil
}

func (p *conclusionPlanMessageHandler) OnPlanStepDelete(ctx context.Context, plan *multiagent.TaskPlan, step *multiagent.PlanStep) (context.Context, error) {
	p.sendPlanEvent(ctx, *plan, false)
	return ctx, nil
}

func (p *conclusionPlanMessageHandler) OnPlanOpEnd(ctx context.Context, plan *multiagent.TaskPlan) (context.Context, error) {
	p.sendPlanEvent(ctx, *plan, true)
	err := p.savePlanMessage(ctx, *plan)
	if err != nil {
		return ctx, err
	}
	//  本次消息结束，更新messageID
	p.messageID = uuid.NewString()
	return ctx, nil
}

func (p *conclusionPlanMessageHandler) sendPlanEvent(_ context.Context, plan multiagent.TaskPlan, isEnd bool) {
	global.GetBroker().PublishToSession(p.mapID, sse.Event{
		ID:   p.nodeID,
		Type: dto.MessagePlanEventType,
		Data: dto.MessagePlanEvent{
			NodeID:    p.nodeID,
			MessageID: p.messageID,
			Plan:      plan,
			IsEnd:     isEnd,
		},
	})
}

func (p *conclusionPlanMessageHandler) savePlanMessage(ctx context.Context, plan multiagent.TaskPlan) error {
	// save plan
	planSteps := make([]model.PlanStep, 0)
	for _, step := range plan.Steps {
		planSteps = append(planSteps, model.PlanStep{
			ID:                 step.ID,
			Name:               step.Name,
			Description:        step.Description,
			AssignedSpecialist: step.AssignedSpecialist,
			Status:             string(step.Status),
		})
	}
	_, err := p.msgManager.SaveConclusionMessage(ctx, p.nodeID, dto.CreateMessageRequest{
		ID:          p.messageID,
		UserID:      p.userID,
		MessageType: model.MsgTypePlan,
		Role:        schema.Assistant,
		Content: model.MessageContent{
			Plan: model.Plan{
				Steps: planSteps,
			},
		},
	})
	if err != nil {
		logger.Error("save plan message failed", zap.Error(err))
		return err
	}
	return nil
}
