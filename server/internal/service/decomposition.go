package service

import (
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

func (s *DecompositionService) Decomposition(ctx *gin.Context, req dto.DecompositionRequest) (event string, sr *schema.StreamReader[*schema.Message], err error) {
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
func (s *DecompositionService) Recognize(ctx *gin.Context, req dto.DecompositionRequest) (event string, sr *schema.StreamReader[*schema.Message], err error) {
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
	agent, err := decomposition.BuildRecognitionAgent(ctx, option)
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
			messageID := uuid.NewString()
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
							Type: dto.MessageTextEventType,
							Data: dto.MessageTextEvent{
								NodeID:    nodeID,
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
						global.GetBroker().Publish(mapID, sse.Event{
							ID:   nodeID,
							Type: dto.MessageTextEventType,
							Data: dto.MessageTextEvent{
								NodeID:    nodeID,
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
						ParentID:    lastMessageID,
						MessageType: model.MsgTypeText,
						Role:        schema.Assistant,
						Content: model.MessageContent{
							Text: thought.String() + finalAnswer.String(),
						},
					}
					msg, err2 := s.msgManager.CreateMessage(ctx, userID, msgReq)
					if err2 != nil {
						logger.Error("create message failed", zap.Error(err2))
						return
					}
					// 将最新messageID挂载到节点上
					s.msgManager.LinkMessageToNode(ctx, req.NodeID, messageID, msg.ConversationID, dto.ConversationTypeDecomposition)
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

// Decompose 拆解节点
func (s *DecompositionService) Decompose(ctx *gin.Context, req dto.DecompositionRequest) (event string, sr *schema.StreamReader[*schema.Message], err error) {
	return
}
