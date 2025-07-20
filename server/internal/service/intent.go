package service

import (
	"fmt"

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

	// 4. 调用意图识别Agent
	agent, err := intent.BuildIntentRecognitionAgent(ctx)
	if err != nil {
		return
	}

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
