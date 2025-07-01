package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/PGshen/thinking-map/server/internal/model"
	"github.com/PGshen/thinking-map/server/internal/model/dto"
	"github.com/PGshen/thinking-map/server/internal/pkg/comm"
	"github.com/PGshen/thinking-map/server/internal/pkg/logger"
	"github.com/PGshen/thinking-map/server/internal/repository"
	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type MessageService struct {
	messageRepo repository.Message
}

func NewMessageService(messageRepo repository.Message) *MessageService {
	return &MessageService{
		messageRepo: messageRepo,
	}
}

// CreateMessage 创建消息
func (s *MessageService) CreateMessage(ctx context.Context, userID string, req dto.CreateMessageRequest) (*dto.MessageResponse, error) {
	// 获取chatID
	var chatID string
	if req.ParentID == "" {
		chatID = uuid.NewString()
	} else {
		// 通过parentID获取
		if parentMsg, err := s.messageRepo.FindByID(ctx, req.ParentID); err == nil {
			chatID = parentMsg.ChatID
		} else {
			chatID = uuid.NewString()
		}
	}
	msg := &model.Message{
		ParentID:    req.ParentID,
		ChatID:      chatID,
		UserID:      userID,
		MessageType: req.MessageType,
		Role:        req.Role,
		Content:     req.Content,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := s.messageRepo.Create(ctx, msg); err != nil {
		return nil, err
	}
	resp := dto.ToMessageResponse(msg)
	return &resp, nil
}

func (s *MessageService) SaveStreamMessage(ctx *gin.Context, sr *schema.StreamReader[*schema.Message], parentID string) {
	useID := ctx.GetString("user_id")

	fullMsgs := make([]*schema.Message, 0)
	defer func() {
		sr.Close()
		fullMsg, err := schema.ConcatMessages(fullMsgs)
		if err != nil {
			logger.Warn("concat message failed", zap.Error(err))
			return
		}
		fullMsg.Content = strings.ReplaceAll(fullMsg.Content, "&nbsp;", " ")
		createMessageRequest := dto.CreateMessageRequest{
			ParentID:    parentID,
			MessageType: comm.MessageTypeText,
			Role:        schema.Assistant,
			Content: model.MessageContent{
				Text: fullMsg.Content,
			},
			Metadata: map[string]any{
				"creater": "ai",
			},
		}
		if _, err := s.CreateMessage(ctx, useID, createMessageRequest); err != nil {
			logger.Error("save message failed", zap.Error(err))
		}
	}()
outer:
	for {
		select {
		case <-ctx.Done():
			fmt.Println("context done", ctx.Err())
			return
		default:
			chunk, err := sr.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					break outer
				}
			}

			fullMsgs = append(fullMsgs, chunk)
		}
	}
}

// UpdateMessage 更新消息
func (s *MessageService) UpdateMessage(ctx context.Context, req dto.UpdateMessageRequest) (*dto.MessageResponse, error) {
	msg, err := s.messageRepo.FindByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if req.MessageType != "" {
		msg.MessageType = req.MessageType
	}
	if !isZeroMessageContent(req.Content) {
		msg.Content = req.Content
	}
	if req.Metadata != nil {
		// 这里可以根据需要序列化 req.Metadata
	}
	msg.UpdatedAt = time.Now()
	if err := s.messageRepo.Update(ctx, msg); err != nil {
		return nil, err
	}
	resp := dto.ToMessageResponse(msg)
	return &resp, nil
}

// DeleteMessage 删除消息
func (s *MessageService) DeleteMessage(ctx context.Context, id string) error {
	return s.messageRepo.Delete(ctx, id)
}

// GetMessageByID 根据ID获取消息
func (s *MessageService) GetMessageByID(ctx context.Context, id string) (*dto.MessageResponse, error) {
	msg, err := s.messageRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := dto.ToMessageResponse(msg)
	return &resp, nil
}

func (s *MessageService) GetMessageByChatID(ctx context.Context, chatID string) ([]*dto.MessageResponse, error) {
	msgs, err := s.messageRepo.FindByChatID(ctx, chatID)
	if err != nil {
		return nil, err
	}
	responses := make([]*dto.MessageResponse, 0, len(msgs))
	for _, msg := range msgs {
		resp := dto.ToMessageResponse(msg)
		responses = append(responses, &resp)
	}
	return responses, nil
}

// GetMessageByParentID
func (s *MessageService) GetMessageByParentID(ctx context.Context, parentID string) ([]*dto.MessageResponse, error) {
	var result []*dto.MessageResponse
	var fetch func(ctx context.Context, pid string) error
	fetch = func(ctx context.Context, pid string) error {
		if pid == uuid.Nil.String() {
			return nil
		}
		msg, err := s.messageRepo.FindByID(ctx, pid)
		if err != nil {
			return err
		}
		resp := dto.ToMessageResponse(msg)
		result = append([]*dto.MessageResponse{&resp}, result...)
		// 递归查找消息
		if err := fetch(ctx, msg.ParentID); err != nil {
			return err
		}
		return nil
	}
	if err := fetch(ctx, parentID); err != nil {
		return nil, err
	}
	return result, nil
}

// isZeroMessageContent 判断 MessageContent 是否为零值
func isZeroMessageContent(mc model.MessageContent) bool {
	return mc.Text == "" && len(mc.RAG) == 0 && len(mc.Notice) == 0
}
