package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

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
	if isDecompose {
		return s.Decompose(ctx, req)
	}
	return s.Recognize(ctx, req)
}

// Recognize performs intent recognition for a given node
func (s *DecompositionService) Recognize(ctx *gin.Context, req dto.DecompositionRequest) (err error) {
	userID := ctx.GetString("user_id")
	// 1. 构建上下文消息
	contextInfo, err := s.contextManager.GetContextInfo(ctx, req.NodeID)
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
	}

	messageSender := &messageSender{
		mapID:      contextInfo.MapInfo.ID,
		nodeID:     req.NodeID,
		userID:     userID,
		msgManager: s.msgManager,
	}
	// 4. 调用意图识别Agent
	agent, err := decomposition.BuildRecognitionAgent(ctx, react.WithMessageHandler(messageSender))
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
type messageSender struct {
	mapID      string
	nodeID     string
	userID     string
	msgManager *global.MessageManager
}

func (m *messageSender) OnMessage(ctx context.Context, message *schema.Message) (context.Context, error) {
	// 用不上，空实现
	return ctx, nil
}

func (m *messageSender) OnStreamMessage(ctx context.Context, sr *schema.StreamReader[*schema.Message]) (context.Context, error) {
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
func (s *DecompositionService) Decompose(ctx *gin.Context, req dto.DecompositionRequest) (err error) {
	return
}
