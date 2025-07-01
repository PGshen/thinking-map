package service

import (
	"fmt"

	"github.com/PGshen/thinking-map/server/internal/agent/callback"
	"github.com/PGshen/thinking-map/server/internal/agent/understanding"
	"github.com/PGshen/thinking-map/server/internal/model"
	"github.com/PGshen/thinking-map/server/internal/model/dto"
	"github.com/PGshen/thinking-map/server/internal/pkg/comm"
	"github.com/PGshen/thinking-map/server/internal/repository"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
)

type UnderstandingService struct {
	messageRepo repository.Message
}

func NewUnderstandingService(messageRepo repository.Message) *UnderstandingService {
	return &UnderstandingService{
		messageRepo: messageRepo,
	}
}

func (s *UnderstandingService) Understanding(ctx *gin.Context, req dto.UnderstandingRequest) (*schema.StreamReader[*schema.Message], error) {
	userID := ctx.GetString("user_id")
	// 1. 构建理解agent
	agent, err := understanding.BuildUnderstandingAgent(ctx)
	if err != nil {
		return nil, err
	}

	// 2. 加载历史消息
	msgService := NewMessageService(s.messageRepo)
	msgs, err := msgService.GetMessageByParentID(ctx, req.ParentMsgID)
	if err != nil {
		return nil, err
	}
	// 消息转换为schema.Message
	schemaMsgs := make([]*schema.Message, len(msgs))
	for i, msg := range msgs {
		schemaMsgs[i] = &schema.Message{
			Role:    msg.Role,
			Content: msg.Content.Text,
		}
	}

	// 3. 调用agent理解
	sr, err := agent.Stream(ctx, schemaMsgs, compose.WithCallbacks(callback.LogCbHandler))
	// 4. 保存消息
	// 4.2 保存用户消息
	msgRequest := dto.CreateMessageRequest{
		ParentID:    req.ParentMsgID,
		MessageType: comm.MessageTypeText,
		Role:        schema.User,
		Content: model.MessageContent{
			Text: fmt.Sprintf("problem: %s\nproblem_type: %s", req.Problem, req.ProblemType),
		},
		Metadata: map[string]any{
			"agent": "Understanding",
		},
	}
	msgResp, err := msgService.CreateMessage(ctx, userID, msgRequest)
	if err != nil {
		return sr, err
	}
	// 4.2 保存模型消息
	srs := sr.Copy(2)
	msgService.SaveStreamMessage(ctx, srs[1], msgResp.ID)
	return srs[0], nil
}
