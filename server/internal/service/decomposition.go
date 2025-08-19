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
				Type: dto.MessageTextEventType,
				Data: dto.MessageTextEvent{
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
		if len(thought.String()) == 0 && len(finalAnswer.String()) == 0 {
			return
		}
		// fmt.Println("fullMsg.Content", fullMsg.Content)
		msgReq := dto.CreateMessageRequest{
			ID:          messageID,
			UserID:      m.userID,
			MessageType: model.MsgTypeText,
			Role:        schema.Assistant,
			Content: model.MessageContent{
				Text: thought.String() + finalAnswer.String(),
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
	planCreationMessageHandler := &planCreationMessageHandler{
		mapID:      contextInfo.MapInfo.ID,
		nodeID:     contextInfo.NodeInfo.ID,
		userID:     userID,
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
		multiagent.WithPlanCreationHandler(planCreationMessageHandler),
		multiagent.WithPlanUpdateHandler(generalMessageHandler),
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
	matcher.On("user_intent", func(value interface{}, path []interface{}) {
		// fmt.Print("thought:", value)
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
	messageID := uuid.NewString()
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
				Type: dto.MessageTextEventType,
				Data: dto.MessageTextEvent{
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
		// fmt.Print("final_answer:", value)
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
		if len(thought.String()) == 0 && len(finalAnswer.String()) == 0 {
			return
		}
		// fmt.Println("fullMsg.Content", fullMsg.Content)
		msgReq := dto.CreateMessageRequest{
			ID:          messageID,
			UserID:      m.userID,
			MessageType: model.MsgTypeText,
			Role:        schema.Assistant,
			Content: model.MessageContent{
				Text: thought.String() + "\n" + finalAnswer.String(),
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

type planCreationMessageHandler struct {
	mapID      string
	nodeID     string
	userID     string
	msgManager *global.MessageManager
}

func (m *planCreationMessageHandler) OnMessage(ctx context.Context, message *schema.Message) (context.Context, error) {
	// 用不上，空实现
	return ctx, nil
}

func (m *planCreationMessageHandler) OnStreamMessage(ctx context.Context, sr *schema.StreamReader[*schema.Message]) (context.Context, error) {
	messageID := uuid.NewString()
	sseBroker := global.GetBroker()
	// 使用流式JSON解析器解析ReasoningOutput
	matcher := utils.NewSimplePathMatcher()
	// 使用增量模式避免重复内容
	parser := utils.NewStreamingJsonParser(matcher, true, true)

	var stepName strings.Builder
	var currentStepIndex int = -1     // 跟踪当前步骤索引
	var isFirstCharOfStep bool = true // 标记是否为步骤的第一个字符

	// 注册路径匹配器来提取userIntent字段
	matcher.On("steps[*].name", func(value interface{}, path []interface{}) {
		// fmt.Print("thought:", value)
		if str, ok := value.(string); ok {
			// 从path中提取数组索引
			var stepIndex int = -1
			for _, segment := range path {
				if idx, isInt := segment.(int); isInt {
					stepIndex = idx
					break
				}
			}

			// 检查是否切换到新的步骤
			if stepIndex != currentStepIndex {
				currentStepIndex = stepIndex
				isFirstCharOfStep = true
			}

			// 构建要发送的消息
			var message string
			if isFirstCharOfStep {
				// 如果是步骤的第一个字符，添加markdown有序列表格式
				if currentStepIndex > 0 {
					// 不是第一个步骤，先换行
					message = fmt.Sprintf("\n%d. %s", currentStepIndex+1, str)
				} else {
					// 第一个步骤
					message = fmt.Sprintf("%d. %s", currentStepIndex+1, str)
				}
				isFirstCharOfStep = false
			} else {
				// 同一步骤的后续字符，直接追加
				message = str
			}

			sseBroker.PublishToSession(m.mapID, sse.Event{
				ID:   m.nodeID,
				Type: dto.MessageTextEventType,
				Data: dto.MessageTextEvent{
					NodeID:    m.nodeID,
					MessageID: messageID,
					Message:   message,
					Mode:      "append",
				},
			})
			stepName.WriteString(message)
		}
	})

	defer func() {
		sr.Close()
		if len(stepName.String()) == 0 {
			return
		}
		// fmt.Println("fullMsg.Content", fullMsg.Content)
		msgReq := dto.CreateMessageRequest{
			ID:          messageID,
			UserID:      m.userID,
			MessageType: model.MsgTypeText,
			Role:        schema.Assistant,
			Content: model.MessageContent{
				Text: stepName.String(),
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
