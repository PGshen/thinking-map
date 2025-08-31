package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/PGshen/thinking-map/server/internal/agent/base"
	"github.com/PGshen/thinking-map/server/internal/agent/base/multiagent"
	"github.com/PGshen/thinking-map/server/internal/agent/base/react"
	"github.com/PGshen/thinking-map/server/internal/global"
	"github.com/PGshen/thinking-map/server/internal/model"
	"github.com/PGshen/thinking-map/server/internal/pkg/logger"
	"github.com/PGshen/thinking-map/server/internal/pkg/sse"
	"github.com/PGshen/thinking-map/server/internal/pkg/utils"
	"github.com/PGshen/thinking-map/server/internal/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/PGshen/thinking-map/server/internal/agent/callback"
	"github.com/PGshen/thinking-map/server/internal/agent/decomposition"
	"github.com/PGshen/thinking-map/server/internal/model/dto"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
)

// DecompositionService handles intent recognition business logic
type DecompositionService struct {
	contextManager *ContextManager
	msgManager     *global.MessageManager
	nodeRepo       repository.ThinkingNode
}

// NewDecompositionService creates a new intent service
func NewDecompositionService(contextManager *ContextManager, nodeRepo repository.ThinkingNode) *DecompositionService {
	return &DecompositionService{
		contextManager: contextManager,
		msgManager:     global.GetMessageManager(),
		nodeRepo:       nodeRepo,
	}
}

func (s *DecompositionService) Decomposition(ctx *gin.Context, req dto.DecompositionRequest) (err error) {
	node, err := s.nodeRepo.FindByID(ctx, req.NodeID)
	if err != nil {
		return
	}
	isDecompose := req.IsDecomposed || node.Decomposition.IsDecomposed // 是否执行拆解
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
	ctxMsg := schema.UserMessage(s.contextManager.FormatContextForAgent(contextInfo))
	messages := []*schema.Message{ctxMsg}
	if req.Clarification != "" {
		messages = append(messages, schema.UserMessage(req.Clarification))
		// 3. 保存用户消息
		_, err = s.msgManager.SaveDecompositionMessage(ctx, req.NodeID, dto.CreateMessageRequest{
			ID:          uuid.NewString(),
			ParentID:    lastMsgID,
			UserID:      userID,
			MessageType: model.MsgTypeText,
			Role:        schema.User,
			Content:     model.MessageContent{Text: req.Clarification},
		})
	}
	// 拆解
	if isDecompose {
		return s.Decompose(ctx, contextInfo, messages)
	}
	// 分析
	return s.Analyze(ctx, contextInfo, messages)
}

// Analyze performs intent analysis for a given node
func (s *DecompositionService) Analyze(ctx *gin.Context, contextInfo *ContextInfo, messages []*schema.Message) (err error) {
	userID := ctx.GetString("user_id")
	defer func() {
		if err != nil {
			logger.Error("Analyze failed", zap.Error(err))
		}
	}()

	messageSender := &analyzeMessageSender{
		mapID:      contextInfo.MapInfo.ID,
		nodeID:     contextInfo.NodeInfo.ID,
		userID:     userID,
		msgManager: s.msgManager,
	}
	// 4. 调用分析Agent
	agent, err := decomposition.BuildAnalysisAgent(ctx, react.WithMessageHandler(messageSender))
	if err != nil {
		return
	}

	// 5. 执行意图识别
	sr, err := agent.Stream(ctx, messages, compose.WithCallbacks(callback.LogCbHandler))
	if err != nil {
		return
	}
	for {
		_, err := sr.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}
	}
	return
}

// 消息发送
type analyzeMessageSender struct {
	mapID      string
	nodeID     string
	userID     string
	msgManager *global.MessageManager
}

func (m *analyzeMessageSender) OnMessage(ctx context.Context, message *schema.Message) (context.Context, error) {
	// 用不上，空实现
	return ctx, nil
}

func (m *analyzeMessageSender) OnStreamMessage(ctx context.Context, sr *schema.StreamReader[*schema.Message]) (context.Context, error) {
	// 生成新的messageID
	messageID := uuid.NewString()
	sseBroker := global.GetBroker()
	// 使用流式JSON解析器解析ReasoningOutput
	matcher := utils.NewSimplePathMatcher()
	// 使用增量模式避免重复内容
	parser := utils.NewStreamingJsonParser(matcher, true, true)

	var thought, finalAnswer strings.Builder

	// 注册路径匹配器来提取thought和final_answer字段
	matcher.On("thought", func(value interface{}, path []interface{}) {
		// fmt.Print("thought:", value)
		if str, ok := value.(string); ok {
			sseBroker.PublishToSession(m.mapID, sse.Event{
				ID:   m.nodeID,
				Type: dto.MessageThoughtEventType,
				Data: dto.MessageThoughtEvent{
					NodeID:    m.nodeID,
					MessageID: messageID,
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
					MessageID: messageID,
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
				ID:          messageID,
				UserID:      m.userID,
				MessageType: model.MsgTypeThought,
				Role:        schema.Assistant,
				Content: model.MessageContent{
					Thought: thought.String(),
				},
			}
			_, err2 := m.msgManager.SaveDecompositionMessage(ctx, m.nodeID, msgReq)
			if err2 != nil {
				logger.Error("create message failed", zap.Error(err2))
				return
			}
		}
		if len(finalAnswer.String()) > 0 {
			msgReq := dto.CreateMessageRequest{
				ID:          messageID,
				UserID:      m.userID,
				MessageType: model.MsgTypeText,
				Role:        schema.Assistant,
				Content: model.MessageContent{
					Text: finalAnswer.String(),
				},
			}
			_, err2 := m.msgManager.SaveDecompositionMessage(ctx, m.nodeID, msgReq)
			if err2 != nil {
				logger.Error("create message failed", zap.Error(err2))
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
			fmt.Print(chunk.Content)
			if err := parser.Write(chunk.Content); err != nil {
				logger.Error("parse reasoning response failed", zap.Error(err))
			}
		}
	}
	return ctx, nil
}

// Decompose 拆解节点
func (s *DecompositionService) Decompose(ctx *gin.Context, contextInfo *ContextInfo, messages []*schema.Message) (err error) {
	userID := ctx.GetString("user_id")
	defer func() {
		if err != nil {
			logger.Error("Decompose failed", zap.Error(err))
		}
		if !contextInfo.NodeInfo.Decomposition.IsDecomposed {
			// 更新拆解状态
			s.nodeRepo.UpdateIsDecomposed(ctx, contextInfo.NodeInfo.ID, true)
		}
	}()

	// 查询当前节点的和子节点列表。作为上下文消息，用于后续操作节点
	childrenMessages, err := s.msgManager.GetNodeChildren(ctx, contextInfo.NodeInfo.ID)
	if err != nil {
		return err
	}
	messages = append(messages, childrenMessages...)

	analyzerMessageHandler := &analyzerMessageHandler{
		mapID:      contextInfo.MapInfo.ID,
		nodeID:     contextInfo.NodeInfo.ID,
		userID:     userID,
		msgManager: s.msgManager,
	}
	generalMessageHandler := &generalMessageHandler{
		mapID:      contextInfo.MapInfo.ID,
		nodeID:     contextInfo.NodeInfo.ID,
		userID:     userID,
		msgManager: s.msgManager,
	}
	specialistMessageHandler := &specialistMessageHandler{
		mapID:      contextInfo.MapInfo.ID,
		nodeID:     contextInfo.NodeInfo.ID,
		userID:     userID,
		msgManager: s.msgManager,
	}
	planMessageHandler := &planMessageHandler{
		mapID:      contextInfo.MapInfo.ID,
		nodeID:     contextInfo.NodeInfo.ID,
		userID:     userID,
		messageID:  uuid.NewString(),
		msgManager: s.msgManager,
	}
	// 将mapID, nodeID保存至ctx, 工具调用时会用到
	ctx.Set("mapID", contextInfo.MapInfo.ID)
	ctx.Set("nodeID", contextInfo.NodeInfo.ID)
	// 4. 调用分析Agent
	agent, err := decomposition.BuildDecompositionAgent(ctx)
	if err != nil {
		return
	}

	// 5. 执行意图识别
	options := []base.AgentOption{
		multiagent.WithConversationAnalyzer(analyzerMessageHandler),
		multiagent.WithDirectAnswerHandler(generalMessageHandler),
		multiagent.WithFinalAnswerHandler(generalMessageHandler),
		multiagent.WithPlanHandler(planMessageHandler),
		multiagent.WithSpecialistHandler("DecompositionDecisionAgent", specialistMessageHandler),
		multiagent.WithSpecialistHandler("ProblemDecompositionAgent", specialistMessageHandler),
		multiagent.WithSpecialistHandler("general_specialist", specialistMessageHandler),
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

type analyzerMessageHandler struct {
	mapID      string
	nodeID     string
	userID     string
	msgManager *global.MessageManager
}

func (m *analyzerMessageHandler) OnMessage(ctx context.Context, message *schema.Message) (context.Context, error) {
	// 用不上，空实现
	return ctx, nil
}

func (m *analyzerMessageHandler) OnStreamMessage(ctx context.Context, sr *schema.StreamReader[*schema.Message]) (context.Context, error) {
	messageID := uuid.NewString()
	sseBroker := global.GetBroker()
	// 使用流式JSON解析器解析ReasoningOutput
	matcher := utils.NewSimplePathMatcher()
	// 使用增量模式避免重复内容
	parser := utils.NewStreamingJsonParser(matcher, true, true)

	var userIntent strings.Builder

	// 注册路径匹配器来提取userIntent字段
	matcher.On("userIntent", func(value interface{}, path []interface{}) {
		// fmt.Print("thought:", value)
		if str, ok := value.(string); ok {
			sseBroker.PublishToSession(m.mapID, sse.Event{
				ID:   m.nodeID,
				Type: dto.MessageThoughtEventType,
				Data: dto.MessageThoughtEvent{
					NodeID:    m.nodeID,
					MessageID: messageID,
					Message:   str,
					Mode:      "append",
				},
			})
			userIntent.WriteString(str)
		}
	})
	defer func() {
		sr.Close()
		if len(userIntent.String()) == 0 {
			return
		}
		// fmt.Println("fullMsg.Content", fullMsg.Content)
		msgReq := dto.CreateMessageRequest{
			ID:          messageID,
			UserID:      m.userID,
			MessageType: model.MsgTypeText,
			Role:        schema.Assistant,
			Content: model.MessageContent{
				Text: userIntent.String(),
			},
		}
		_, err2 := m.msgManager.SaveDecompositionMessage(ctx, m.nodeID, msgReq)
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
			if err := parser.Write(chunk.Content); err != nil {
				logger.Error("parse reasoning response failed", zap.Error(err))
			}
		}
	}
	return ctx, nil
}

// 普通消息发送器
type generalMessageHandler struct {
	mapID      string
	nodeID     string
	userID     string
	msgManager *global.MessageManager
}

func (m *generalMessageHandler) OnMessage(ctx context.Context, message *schema.Message) (context.Context, error) {
	return ctx, nil
}

func (m *generalMessageHandler) OnStreamMessage(ctx context.Context, sr *schema.StreamReader[*schema.Message]) (context.Context, error) {
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
		_, err2 := m.msgManager.SaveDecompositionMessage(ctx, m.nodeID, msgReq)
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

type specialistMessageHandler struct {
	mapID      string
	nodeID     string
	userID     string
	msgManager *global.MessageManager
}

func (m *specialistMessageHandler) OnMessage(ctx context.Context, message *schema.Message) (context.Context, error) {
	// 用不上，空实现
	return ctx, nil
}

func (m *specialistMessageHandler) OnStreamMessage(ctx context.Context, sr *schema.StreamReader[*schema.Message]) (context.Context, error) {
	thoughtMessageID := uuid.NewString()
	finalAnswerMessageID := uuid.NewString()
	sseBroker := global.GetBroker()
	// 使用流式JSON解析器解析ReasoningOutput
	matcher := utils.NewSimplePathMatcher()
	// 使用增量模式避免重复内容
	parser := utils.NewStreamingJsonParser(matcher, true, true)

	var thought, finalAnswer strings.Builder

	// 注册路径匹配器来提取userIntent字段
	matcher.On("thought", func(value interface{}, path []interface{}) {
		// fmt.Print("thought:", value)
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
		// fmt.Print("final_answer:", value)
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
			_, err2 := m.msgManager.SaveDecompositionMessage(ctx, m.nodeID, msgReq)
			if err2 != nil {
				logger.Error("create message failed", zap.Error(err2))
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
			_, err2 := m.msgManager.SaveDecompositionMessage(ctx, m.nodeID, msgReq)
			if err2 != nil {
				logger.Error("create message failed", zap.Error(err2))
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
				fmt.Print(chunk.Content)
				if err := parser.Write(chunk.Content); err != nil {
					logger.Error("parse reasoning response failed", zap.Error(err))
				}
			}
		}
	}
	return ctx, nil
}

type planMessageHandler struct {
	mapID      string
	nodeID     string
	userID     string
	messageID  string
	msgManager *global.MessageManager
}

func (p *planMessageHandler) OnPlanStepCreate(ctx context.Context, plan *multiagent.TaskPlan, step *multiagent.PlanStep) (context.Context, error) {
	p.sendPlanEvent(ctx, *plan, false)
	return ctx, nil
}

func (p *planMessageHandler) OnPlanStepUpdate(ctx context.Context, plan *multiagent.TaskPlan, step *multiagent.PlanStep) (context.Context, error) {
	p.sendPlanEvent(ctx, *plan, false)
	return ctx, nil
}

func (p *planMessageHandler) OnPlanStepStatusUpdate(ctx context.Context, plan *multiagent.TaskPlan, step *multiagent.PlanStep) (context.Context, error) {
	p.sendPlanEvent(ctx, *plan, false)
	return ctx, nil
}

func (p *planMessageHandler) OnPlanStepDelete(ctx context.Context, plan *multiagent.TaskPlan, step *multiagent.PlanStep) (context.Context, error) {
	p.sendPlanEvent(ctx, *plan, false)
	return ctx, nil
}

func (p *planMessageHandler) OnPlanStepEnd(ctx context.Context, plan *multiagent.TaskPlan) (context.Context, error) {
	p.sendPlanEvent(ctx, *plan, true)
	err := p.savePlanMessage(ctx, *plan)
	if err != nil {
		return ctx, err
	}
	//  本次消息结束，更新messageID
	p.messageID = uuid.NewString()
	return ctx, nil
}

func (p *planMessageHandler) sendPlanEvent(_ context.Context, plan multiagent.TaskPlan, isEnd bool) {
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

func (p *planMessageHandler) savePlanMessage(ctx context.Context, plan multiagent.TaskPlan) error {
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
	_, err := p.msgManager.SaveDecompositionMessage(ctx, p.nodeID, dto.CreateMessageRequest{
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
