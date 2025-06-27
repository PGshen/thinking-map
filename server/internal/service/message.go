package service

import (
	"context"
	"time"

	"github.com/PGshen/thinking-map/server/internal/model"
	"github.com/PGshen/thinking-map/server/internal/model/dto"
	"github.com/PGshen/thinking-map/server/internal/repository"
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
func (s *MessageService) CreateMessage(ctx context.Context, req dto.CreateMessageRequest) (*dto.MessageResponse, error) {
	msg := &model.Message{
		NodeID:      req.NodeID,
		ParentID:    req.ParentID,
		MessageType: req.MessageType,
		Content:     req.Content,
		Metadata:    nil, // 这里可以根据需要序列化 req.Metadata
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := s.messageRepo.Create(ctx, msg); err != nil {
		return nil, err
	}
	resp := dto.ToMessageResponse(msg)
	return &resp, nil
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

// ListMessagesByNodeID 根据节点ID分页获取消息
func (s *MessageService) ListMessagesByNodeID(ctx context.Context, nodeID string, page, limit int) (*dto.MessageListResponse, error) {
	offset := (page - 1) * limit
	msgs, total, err := s.messageRepo.FindByNodeID(ctx, nodeID, offset, limit)
	if err != nil {
		return nil, err
	}
	items := make([]dto.MessageResponse, len(msgs))
	for i, m := range msgs {
		items[i] = dto.ToMessageResponse(m)
	}
	return &dto.MessageListResponse{
		Total: int(total),
		Page:  page,
		Limit: limit,
		Items: items,
	}, nil
}

// isZeroMessageContent 判断 MessageContent 是否为零值
func isZeroMessageContent(mc model.MessageContent) bool {
	return mc.Text == "" && len(mc.RAG) == 0 && len(mc.Notice) == 0
}
