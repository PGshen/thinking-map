package service

import (
	"errors"
	"fmt"
	"io"

	"github.com/PGshen/thinking-map/server/internal/agent/callback"
	"github.com/PGshen/thinking-map/server/internal/agent/conclusion"
	"github.com/PGshen/thinking-map/server/internal/global"
	"github.com/PGshen/thinking-map/server/internal/model/dto"
	"github.com/PGshen/thinking-map/server/internal/pkg/logger"
	"github.com/PGshen/thinking-map/server/internal/repository"
	"github.com/cloudwego/eino/compose"
	"github.com/gin-gonic/gin"
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
	return s.Conclude(ctx, contextInfo, userMessage)
}

// Conclude 生成节点结论
func (s *ConclusionService) Conclude(ctx *gin.Context, contextInfo *ContextInfo, userMessage *conclusion.UserMessage) (err error) {
	defer func() {
		if err != nil {
			logger.Error("Generate conclusion failed", zap.Error(err))
		}
	}()

	// 创建消息处理器，用于处理最终结果
	// 注意：由于我们不再使用WithFinalAnswerHandler，这个处理器暂时不需要实现Handle方法
	// 如果需要处理最终结果，可以在流式输出处理中进行

	// 将mapID, nodeID保存至ctx, 工具调用时会用到
	ctx.Set("mapID", contextInfo.MapInfo.ID)
	ctx.Set("nodeID", contextInfo.NodeInfo.ID)

	// 调用结论生成Agent
	agent, err := conclusion.BuildConclusionAgent(ctx)
	if err != nil {
		return
	}

	// 执行结论生成
	// 使用基于Eino Graph的结论生成Agent，不再需要multiagent相关选项
	opts := []compose.Option{
		compose.WithCallbacks(callback.LogCbHandler),
	}

	// 流式执行结论生成
	sr, err := agent.Stream(ctx, userMessage, opts...)
	if err != nil {
		return
	}

	// 处理流式输出
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

	// 结论生成完成后的处理
	// 注意：如果需要更新节点的结论生成状态，需要在ThinkingNode接口中添加UpdateConclusionGenerated方法

	return
}
