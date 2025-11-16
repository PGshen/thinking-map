package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/PGshen/thinking-map/server/internal/agent/base/react"
	"github.com/PGshen/thinking-map/server/internal/agent/callback"
	conclusionv3 "github.com/PGshen/thinking-map/server/internal/agent/conclusion"
	"github.com/PGshen/thinking-map/server/internal/global"
	"github.com/PGshen/thinking-map/server/internal/model"
	"github.com/PGshen/thinking-map/server/internal/model/dto"
	"github.com/PGshen/thinking-map/server/internal/pkg/comm"
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

type ConclusionService struct {
	contextManager *ContextManager
	msgManager     *global.MessageManager
	nodeRepo       repository.ThinkingNode
}

func NewConclusionV3Service(contextManager *ContextManager, nodeRepo repository.ThinkingNode) *ConclusionService {
	return &ConclusionService{
		contextManager: contextManager,
		msgManager:     global.GetMessageManager(),
		nodeRepo:       nodeRepo,
	}
}

func (c *ConclusionService) Conclusion(ctx *gin.Context, req dto.ConclusionRequest) error {
	// 查询上下文
	contextInfo, err := c.contextManager.GetNodeContextWithConversation(ctx, req.NodeID, "")
	if err != nil {
		return err
	}
	ctx.Set("mapID", contextInfo.MapInfo.ID)
	ctx.Set("nodeID", req.NodeID)
	ctx.Set("operation", "conclusion")

	// 2. 构建用户消息
	//  2.1 上下文消息
	ctxMsg := schema.UserMessage(c.contextManager.FormatContextForAgent(contextInfo))
	messages := []*schema.Message{ctxMsg}

	// 2.2 查询当前节点的和子节点列表。作为上下文消息，用于后续操作节点
	childrenMessages, err := c.msgManager.GetNodeChildren(ctx, contextInfo.NodeInfo.ID)
	if err != nil {
		return err
	}
	messages = append(messages, childrenMessages...)
	// 2.3 用户指令
	if req.Instruction != "" {
		instruction := req.Instruction
		if req.Reference != "" {
			instruction = fmt.Sprintf("引用内容：%s\n指令要求：%s", req.Reference, instruction)
		}
		messages = append(messages, schema.UserMessage(instruction))
	}
	if contextInfo.NodeInfo.Conclusion.Content != "" {
		// 有结论，优化结论
		go c.Optimize(ctx, messages)
	}
	// 无结论，生成结论
	go c.Generate(ctx, contextInfo, messages)
	return nil
}

func (c *ConclusionService) Generate(ctx *gin.Context, contextInfo *ContextInfo, messages []*schema.Message) error {
	userID := ctx.GetString("user_id")
	messageHandler := &messageHandler{
		mapID:      contextInfo.MapInfo.ID,
		nodeID:     contextInfo.NodeInfo.ID,
		userID:     userID,
		msgManager: c.msgManager,
	}
	agent, err := conclusionv3.BuildGenerationAgent(ctx, react.WithMessageHandler(messageHandler))
	if err != nil {
		return err
	}
	sr, err := agent.Stream(ctx, messages, compose.WithCallbacks(callback.LogCbHandler))
	if err != nil {
		return err
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
	global.GetBroker().PublishToSession(contextInfo.MapInfo.ID, sse.Event{
		ID:   contextInfo.NodeInfo.ID,
		Type: dto.ConclusionCompletedEventType,
		Data: dto.ConclusionCompletedEvent{
			NodeID: contextInfo.NodeInfo.ID,
			Mode:   "generate",
			Status: "completed",
		},
	})
	return nil
}

func (c ConclusionService) Optimize(ctx *gin.Context, messages []*schema.Message) error {
	agent, err := conclusionv3.BuildOptimizationAgent(ctx)
	if err != nil {
		return err
	}
	sr, err := agent.Stream(ctx, messages, compose.WithCallbacks(callback.LogCbHandler))
	if err != nil {
		return err
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
	mapID := ctx.GetString("mapID")
	nodeID := ctx.GetString("nodeID")
	global.GetBroker().PublishToSession(mapID, sse.Event{
		ID:   nodeID,
		Type: dto.ConclusionCompletedEventType,
		Data: dto.ConclusionCompletedEvent{
			NodeID: nodeID,
			Mode:   "optimize",
			Status: "completed",
		},
	})
	return nil
}

type messageHandler struct {
	mapID      string
	nodeID     string
	userID     string
	msgManager *global.MessageManager
}

func (m *messageHandler) OnMessage(ctx context.Context, message *schema.Message) (context.Context, error) {
	return ctx, nil
}

func (m *messageHandler) OnStreamMessage(ctx context.Context, sr *schema.StreamReader[*schema.Message]) (context.Context, error) {
	messageID := uuid.NewString()
	sseBroker := global.GetBroker()
	matcher := utils.NewSimplePathMatcher()
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
			fmt.Print(str)
		}
	})

	matcher.On("final_answer", func(value interface{}, path []interface{}) {
		if str, ok := value.(string); ok {
			sseBroker.PublishToSession(m.mapID, sse.Event{
				ID:   m.nodeID,
				Type: dto.MessageConclusionEventType,
				Data: dto.MessageThoughtEvent{
					NodeID:    m.nodeID,
					MessageID: messageID,
					Message:   str,
					Mode:      "generate",
				},
			})
			finalAnswer.WriteString(str)
			fmt.Print(str)
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
			_, err2 := m.msgManager.SaveConclusionMessage(ctx, m.nodeID, msgReq)
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
			// fmt.Print(chunk.Content)
			if err := parser.Write(chunk.Content); err != nil {
				logger.Error("parse reasoning response failed", zap.Error(err))
			}
		}
	}
	return ctx, nil
}

// SaveConclusion 保存结论
func (c *ConclusionService) SaveConclusion(ctx *gin.Context, nodeID string, req dto.SaveConclusionRequest) error {
	// 获取当前节点
	node, err := c.nodeRepo.FindByID(ctx, nodeID)
	if err != nil {
		logger.Error("Failed to get node", zap.String("nodeID", nodeID), zap.Error(err))
		return fmt.Errorf("failed to get node: %w", err)
	}

	// 更新结论内容，保留原有的 conversationID 和 lastMessageID
	node.Conclusion.Content = req.Content
	node.Status = comm.NodeStatusCompleted

	// 更新数据库
	if err := c.nodeRepo.Update(ctx, node); err != nil {
		logger.Error("Failed to update node conclusion", zap.String("nodeID", nodeID), zap.Error(err))
		return fmt.Errorf("failed to update node conclusion: %w", err)
	}

	logger.Info("Conclusion saved successfully", zap.String("nodeID", nodeID))
	return nil
}

// ResetConclusion 重置结论
func (c *ConclusionService) ResetConclusion(ctx *gin.Context, nodeID string) error {
	// 获取当前节点
	node, err := c.nodeRepo.FindByID(ctx, nodeID)
	if err != nil {
		logger.Error("Failed to get node", zap.String("nodeID", nodeID), zap.Error(err))
		return fmt.Errorf("failed to get node: %w", err)
	}

	// 重置结论相关字段
	node.Conclusion = model.Conclusion{
		ConversationID: "",
		LastMessageID:  "",
		Content:        "",
	}
	node.Status = comm.NodeStatusInConclusion

	// 更新数据库
	if err := c.nodeRepo.Update(ctx, node); err != nil {
		logger.Error("Failed to reset node conclusion", zap.String("nodeID", nodeID), zap.Error(err))
		return fmt.Errorf("failed to reset node conclusion: %w", err)
	}

	logger.Info("Conclusion reset successfully", zap.String("nodeID", nodeID))
	return nil
}
