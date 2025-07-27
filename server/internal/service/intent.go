package service

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/PGshen/thinking-map/server/internal/agent/react"
	"github.com/PGshen/thinking-map/server/internal/pkg/global"
	"github.com/PGshen/thinking-map/server/internal/pkg/logger"
	"github.com/PGshen/thinking-map/server/internal/pkg/sse"
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
	msgManager     *MessageManager
}

// NewIntentService creates a new intent service
func NewIntentService(contextManager *ContextManager, msgManager *MessageManager) *IntentService {
	return &IntentService{
		contextManager: contextManager,
		msgManager:     msgManager,
	}
}

// RecognizeIntent performs intent recognition for a given node
func (s *IntentService) RecognizeIntent(ctx *gin.Context, req dto.IntentRequest) (event string, sr *schema.StreamReader[*schema.Message], err error) {
	// 1. 构建上下文消息
	contextInfo, err := s.contextManager.GetContextInfo(ctx, req.NodeID)
	if err != nil {
		return
	}

	// 2. 构建用户消息
	ctxMsg := schema.UserMessage(s.contextManager.FormatContextForAgent(contextInfo))
	userContent := fmt.Sprintf("mapID: %s, nodeID: %s, msgID: %s", contextInfo.MapInfo.ID, req.NodeID, req.MsgID)
	if req.Clarification != "" {
		userContent += fmt.Sprintf(", clarification: %s", req.Clarification)
	}
	userMsg := schema.UserMessage(userContent)

	messages := []*schema.Message{ctxMsg, userMsg}

	option, future := react.WithMessageFuture()
	// 4. 调用意图识别Agent
	agent, err := intent.BuildIntentRecognitionAgent(ctx, option)
	if err != nil {
		return
	}

	go func() {
		sIter := future.GetMessageStreams()

		for {
			// First message should be the assistant message for tool calling
			stream, hasNext, err := sIter.Next()
			if err != nil {
				break
			}
			if !hasNext {
				break
			}
			// 开启新协程处理
			go func(ctx *gin.Context, mapID string, s *schema.StreamReader[*schema.Message]) {
				fullMsgs := make([]*schema.Message, 0)
				defer func() {
					s.Close()
					fullMsg, err := schema.ConcatMessages(fullMsgs)
					if err != nil {
						logger.Warn("concat message failed", zap.Error(err))
						return
					}
					fullMsg.Content = strings.ReplaceAll(fullMsg.Content, "&nbsp;", " ")
					fmt.Println("fullMsg.Content", fullMsg.Content)
				}()
			outer:
				for {
					select {
					case <-ctx.Done():
						fmt.Println("context done", ctx.Err())
						return
					default:
						chunk, err := s.Recv()
						if err != nil {
							if errors.Is(err, io.EOF) {
								break outer
							}
						}

						fullMsgs = append(fullMsgs, chunk)
						msgPart, err := schema.ConcatMessages(fullMsgs)
						global.GetBroker().Publish(mapID, sse.Event{
							ID:   mapID,
							Type: dto.MsgTextEventType,
							Data: msgPart,
						})
					}
				}
			}(ctx, contextInfo.MapInfo.ID, stream)
		}
	}()

	// 5. 执行意图识别
	sr, err = agent.Stream(ctx, messages, compose.WithCallbacks(callback.LogCbHandler))
	if err != nil {
		return
	}
	srs := sr.Copy(2)
	sr = srs[0]
	// 6. 保存消息记录
	go s.msgManager.SaveStreamMessage(ctx, srs[1], req.MsgID, "")
	return
}
