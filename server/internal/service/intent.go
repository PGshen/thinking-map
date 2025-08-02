package service

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/PGshen/thinking-map/server/internal/agent/react"
	"github.com/PGshen/thinking-map/server/internal/global"
	"github.com/PGshen/thinking-map/server/internal/model"
	"github.com/PGshen/thinking-map/server/internal/pkg/logger"
	"github.com/PGshen/thinking-map/server/internal/pkg/sse"
	"github.com/PGshen/thinking-map/server/internal/pkg/utils"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/PGshen/thinking-map/server/internal/agent/callback"
	"github.com/PGshen/thinking-map/server/internal/agent/intent"
	"github.com/PGshen/thinking-map/server/internal/model/dto"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
)

// IntentService handles intent recognition business logic
type IntentService struct {
	contextManager *ContextManager
	msgManager     *global.MessageManager
}

// NewIntentService creates a new intent service
func NewIntentService(contextManager *ContextManager) *IntentService {
	return &IntentService{
		contextManager: contextManager,
		msgManager:     global.GetMessageManager(),
	}
}

// RecognizeDecomposition performs intent recognition for a given node
func (s *IntentService) RecognizeDecomposition(ctx *gin.Context, req dto.DecompositionRecognitionRequest) (event string, sr *schema.StreamReader[*schema.Message], err error) {
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

	option, future := react.WithMessageFuture()
	// 4. 调用意图识别Agent
	agent, err := intent.BuildIntentRecognitionAgent(ctx, option)
	if err != nil {
		return
	}

	go func() {
		sIter := future.GetMessageStreams()
		lastMessageID := contextInfo.NodeInfo.Decomposition.LastMessageID
		for {
			// First message should be the assistant message for tool calling
			stream, hasNext, err2 := sIter.Next()
			if err2 != nil {
				break
			}
			if !hasNext {
				break
			}
			// 不开启新协程处理，按顺序接收消息
			func(ctx *gin.Context, mapID string, nodeID string, sr *schema.StreamReader[*schema.Message]) {
				// 使用流式JSON解析器解析ReasoningOutput
				matcher := utils.NewSimplePathMatcher()
				// 使用增量模式避免重复内容
				parser := utils.NewStreamingJsonParser(matcher, true, true)

				var thought, finalAnswer strings.Builder

				// 注册路径匹配器来提取thought和final_answer字段
				matcher.On("thought", func(value interface{}, path []interface{}) {
					// fmt.Print("thought:", value)
					if str, ok := value.(string); ok {
						global.GetBroker().Publish(mapID, sse.Event{
							ID:   nodeID,
							Type: dto.MsgTextEventType,
							Data: str,
						})
						thought.WriteString(str)
					}
				})

				matcher.On("final_answer", func(value interface{}, path []interface{}) {
					if str, ok := value.(string); ok {
						global.GetBroker().Publish(mapID, sse.Event{
							ID:   nodeID,
							Type: dto.MsgTextEventType,
							Data: str,
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
					messageID := uuid.NewString()
					msgReq := dto.CreateMessageRequest{
						ID:          messageID,
						ParentID:    lastMessageID,
						MessageType: model.MsgTypeText,
						Role:        schema.Assistant,
						Content: model.MessageContent{
							Text: thought.String() + finalAnswer.String(),
						},
					}
					_, err = s.msgManager.CreateMessage(ctx, userID, msgReq)
					if err != nil {
						logger.Error("create message failed", zap.Error(err))
						return
					}
					// 将最新messageID挂载到节点上
					s.msgManager.LinkMessageToNode(ctx, req.NodeID, messageID, dto.ConversationTypeDecomposition)
					lastMessageID = messageID // 最新消息ID
				}()
			outer:
				for {
					select {
					case <-ctx.Done():
						fmt.Println("context done", ctx.Err())
						return
					default:
						chunk, err2 := sr.Recv()
						if err2 != nil {
							if errors.Is(err2, io.EOF) {
								fmt.Println()
								break outer
							}
						}
						// 不保存工具调用的
						if chunk.Role == schema.Tool {
							return
						}
						fmt.Print(chunk.Content)
						if err = parser.Write(chunk.Content); err != nil {
							logger.Error("parse reasoning response failed", zap.Error(err))
						}
					}
				}
			}(ctx, contextInfo.MapInfo.ID, req.NodeID, stream)
		}
	}()

	// 5. 执行意图识别
	sr, err = agent.Stream(ctx, messages, compose.WithCallbacks(callback.LogCbHandler))
	if err != nil {
		return
	}
	// srs := sr.Copy(2)
	// sr = srs[0]
	// 6. 保存消息记录
	// go s.msgManager.SaveStreamMessage(ctx, srs[1], req.MsgID, "")
	return
}
