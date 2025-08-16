package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

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
	isDecompose := req.IsDecompose || node.Decomposition.IsDecomposed // 是否执行拆解
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
	// 使用流式JSON解析器解析ReasoningOutput
	matcher := utils.NewSimplePathMatcher()
	// 使用增量模式避免重复内容
	parser := utils.NewStreamingJsonParser(matcher, true, true)

	var thought, finalAnswer strings.Builder

	// 注册路径匹配器来提取thought和final_answer字段
	matcher.On("thought", func(value interface{}, path []interface{}) {
		// fmt.Print("thought:", value)
		if str, ok := value.(string); ok {
			global.GetBroker().Publish(m.mapID, sse.Event{
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
			global.GetBroker().Publish(m.mapID, sse.Event{
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

	conversationAnalyzerMessageSender := &conversationAnalyzerMessageSender{
		mapID:      contextInfo.MapInfo.ID,
		nodeID:     contextInfo.NodeInfo.ID,
		userID:     userID,
		msgManager: s.msgManager,
	}
	messageSender := &messageSender{
		mapID:      contextInfo.MapInfo.ID,
		nodeID:     contextInfo.NodeInfo.ID,
		userID:     userID,
		msgManager: s.msgManager,
	}
	// 4. 调用分析Agent
	agent, err := decomposition.BuildDecompositionAgent(ctx,
		multiagent.WithConversationAnalyzer(conversationAnalyzerMessageSender),
		multiagent.WithDirectAnswerHandler(messageSender),
		multiagent.WithFinalAnswerHandler(messageSender),
		multiagent.WithSpecialistHandler("DecompositionDecisionAgent", messageSender),
		multiagent.WithSpecialistHandler("ProblemDecompositionAgent", messageSender),
	)
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

type conversationAnalyzerMessageSender struct {
	mapID      string
	nodeID     string
	userID     string
	msgManager *global.MessageManager
}

func (m *conversationAnalyzerMessageSender) OnMessage(ctx context.Context, message *schema.Message) (context.Context, error) {
	// 用不上，空实现
	return ctx, nil
}

func (m *conversationAnalyzerMessageSender) OnStreamMessage(ctx context.Context, sr *schema.StreamReader[*schema.Message]) (context.Context, error) {
	messageID := uuid.NewString()
	// 使用流式JSON解析器解析ReasoningOutput
	matcher := utils.NewSimplePathMatcher()
	// 使用增量模式避免重复内容
	parser := utils.NewStreamingJsonParser(matcher, true, true)

	var userIntent strings.Builder

	// 注册路径匹配器来提取userIntent字段
	matcher.On("user_intent", func(value interface{}, path []interface{}) {
		// fmt.Print("thought:", value)
		if str, ok := value.(string); ok {
			global.GetBroker().Publish(m.mapID, sse.Event{
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
type messageSender struct {
	mapID      string
	nodeID     string
	userID     string
	msgManager *global.MessageManager
}

func (m *messageSender) OnMessage(ctx context.Context, message *schema.Message) (context.Context, error) {
	return ctx, nil
}

func (m *messageSender) OnStreamMessage(ctx context.Context, sr *schema.StreamReader[*schema.Message]) (context.Context, error) {
	messageID := uuid.NewString()
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
			global.GetBroker().Publish(m.mapID, sse.Event{
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
