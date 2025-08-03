package service

import (
	"fmt"

	"github.com/PGshen/thinking-map/server/internal/agent/callback"
	"github.com/PGshen/thinking-map/server/internal/agent/understanding"
	"github.com/PGshen/thinking-map/server/internal/global"
	"github.com/PGshen/thinking-map/server/internal/model"
	"github.com/PGshen/thinking-map/server/internal/model/dto"
	"github.com/PGshen/thinking-map/server/internal/pkg/comm"
	"github.com/PGshen/thinking-map/server/internal/pkg/utils"
	"github.com/PGshen/thinking-map/server/internal/repository"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UnderstandingService struct {
	messageRepo repository.Message
	nodeRepo    repository.ThinkingNode
}

func NewUnderstandingService(messageRepo repository.Message, nodeRepo repository.ThinkingNode) *UnderstandingService {
	return &UnderstandingService{
		messageRepo: messageRepo,
		nodeRepo:    nodeRepo,
	}
}

func (s *UnderstandingService) Understanding(ctx *gin.Context, req dto.UnderstandingRequest) (event string, sr *schema.StreamReader[*schema.Message], err error) {
	event = comm.EventJson
	userID := ctx.GetString("user_id")
	// 1. 构建理解agent
	var agent compose.Runnable[[]*schema.Message, *schema.Message]
	agent, err = understanding.BuildUnderstandingAgent(ctx)
	if err != nil {
		return
	}
	var userContent string
	if req.Supplementary != "" {
		// 补充内容消息
		userContent = req.Supplementary
	} else if req.Problem != "" && req.ProblemType != "" {
		userContent = fmt.Sprintf("问题：%s\n类型：%s", req.Problem, req.ProblemType)
	}

	// 2. 加载历史消息
	msgManager := global.GetMessageManager()
	var msgs []*dto.MessageResponse
	if req.ParentMsgID != "" {
		// 先获取父消息以得到conversationID
		parentMsg, err := msgManager.GetMessageByID(ctx, req.ParentMsgID)
		if err != nil {
			return "", nil, err
		}
		msgs, err = msgManager.GetMessageChain(ctx, req.ParentMsgID, parentMsg.ConversationID)
		if err != nil {
			return "", nil, err
		}
	} else {
		// 没有父消息，使用空的消息链
		msgs = []*dto.MessageResponse{}
	}
	// 消息转换为schema.Message
	schemaMsgs := global.ConvertToSchemaMsg(msgs)
	schemaMsgs = append(schemaMsgs, schema.UserMessage(userContent))

	// 3. 调用agent理解
	sr, err = agent.Stream(ctx, schemaMsgs, compose.WithCallbacks(callback.LogCbHandler))
	// 4. 保存消息
	// 4.2 保存用户消息
	msgRequest := dto.CreateMessageRequest{
		ParentID:    utils.Ternary(req.ParentMsgID == "", uuid.Nil.String(), req.ParentMsgID),
		MessageType: comm.MessageTypeText,
		Role:        schema.User,
		Content: model.MessageContent{
			Text: userContent,
		},
		Metadata: map[string]any{
			"agent": "Understanding",
		},
	}
	var msgResp *dto.MessageResponse
	msgResp, err = msgManager.CreateMessage(ctx, userID, msgRequest)
	if err != nil {
		return
	}
	// 4.2 保存模型消息
	srs := sr.Copy(2)
	sr = srs[0]
	// 流式消息，提前确定消息ID
	go msgManager.SaveStreamMessage(ctx, srs[1], req.MsgID, msgResp.ID)
	return
}
