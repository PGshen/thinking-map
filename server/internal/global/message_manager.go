/*
 * @Date: 2025-01-27 00:00:00
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-01-27 00:00:00
 * @FilePath: /thinking-map/server/internal/global/message_manager.go
 */
package global

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
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
	"gorm.io/gorm"
)

var (
	// GlobalMessageManager 全局消息管理器实例
	GlobalMessageManager *MessageManager
	messageManagerOnce   sync.Once
)

// MessageManager 消息管理器
type MessageManager struct {
	messageRepo repository.Message
	nodeRepo    repository.ThinkingNode
}

// InitMessageManager 初始化全局消息管理器
func InitMessageManager(messageRepo repository.Message, nodeRepo repository.ThinkingNode) {
	messageManagerOnce.Do(func() {
		GlobalMessageManager = &MessageManager{
			messageRepo: messageRepo,
			nodeRepo:    nodeRepo,
		}
	})
}

// GetMessageManager 获取全局消息管理器实例
func GetMessageManager() *MessageManager {
	if GlobalMessageManager == nil {
		panic("message manager not initialized, call InitMessageManager first")
	}
	return GlobalMessageManager
}

// CreateMessage 创建消息
func (s *MessageManager) CreateMessage(ctx context.Context, userID string, req dto.CreateMessageRequest) (*dto.MessageResponse, error) {
	// 获取conversationID
	var conversationID string
	if req.ParentID == "" {
		conversationID = uuid.NewString()
	} else {
		// 通过parentID获取
		if parentMsg, err := s.messageRepo.FindByID(ctx, req.ParentID); err == nil {
			conversationID = parentMsg.ConversationID
			if userID == "" { // 没有传userID,则继承父消息的
				userID = parentMsg.UserID
			}
		} else {
			conversationID = uuid.NewString()
		}
	}
	// 处理空ParentID
	parentID := req.ParentID
	if parentID == "" {
		parentID = uuid.Nil.String()
	}

	msg := &model.Message{
		ID:             req.ID,
		ParentID:       parentID,
		ConversationID: conversationID,
		UserID:         userID,
		MessageType:    req.MessageType,
		Role:           req.Role,
		Content:        req.Content,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	if err := s.messageRepo.Create(ctx, msg); err != nil {
		return nil, err
	}
	resp := dto.ToMessageResponse(msg)
	return &resp, nil
}

// SaveDecompositionMessage 保存分解消息
func (s *MessageManager) SaveDecompositionMessage(ctx context.Context, nodeID string, req dto.CreateMessageRequest) (*dto.MessageResponse, error) {
	// 0. 找到节点
	node, err := s.nodeRepo.FindByID(ctx, nodeID)
	if err != nil {
		return nil, err
	}
	lastMessageID := node.Decomposition.LastMessageID
	// 1. 保存消息
	req.ParentID = lastMessageID
	msg, err := s.CreateMessage(ctx, "", req)
	if err != nil {
		return nil, err
	}
	// 2. 更新节点的最后消息ID
	if err := s.LinkMessageToNode(ctx, nodeID, msg.ID, dto.ConversationTypeDecomposition); err != nil {
		return nil, err
	}
	return msg, nil
}

func (s *MessageManager) SaveStreamMessage(ctx *gin.Context, sr *schema.StreamReader[*schema.Message], ID, parentID string) {
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
			ID:          ID,
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
func (s *MessageManager) UpdateMessage(ctx context.Context, req dto.UpdateMessageRequest) (*dto.MessageResponse, error) {
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
	if err = s.messageRepo.Update(ctx, msg); err != nil {
		return nil, err
	}
	// 重新从数据库获取更新后的记录
	updatedMsg, err := s.messageRepo.FindByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	resp := dto.ToMessageResponse(updatedMsg)
	return &resp, nil
}

// DeleteMessage 删除消息
func (s *MessageManager) DeleteMessage(ctx context.Context, id string) error {
	return s.messageRepo.Delete(ctx, id)
}

// GetMessageByID 根据ID获取消息
func (s *MessageManager) GetMessageByID(ctx context.Context, id string) (*dto.MessageResponse, error) {
	msg, err := s.messageRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := dto.ToMessageResponse(msg)
	return &resp, nil
}

func (s *MessageManager) GetMessageByConversationID(ctx context.Context, conversationID string) ([]*dto.MessageResponse, error) {
	msgs, err := s.messageRepo.FindByConversationID(ctx, conversationID)
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

func ConvertToSchemaMsg(list []*dto.MessageResponse) []*schema.Message {
	result := make([]*schema.Message, 0)
	for _, item := range list {
		result = append(result, &schema.Message{
			Role:    item.Role,
			Content: item.Content.String(),
		})
	}
	return result
}

// CreateConversation 创建新会话
func (s *MessageManager) CreateConversation(ctx context.Context, nodeID string) (string, error) {
	conversationID := uuid.NewString()
	return conversationID, nil
}

// RollbackConversation 回退会话到指定消息
func (s *MessageManager) RollbackConversation(ctx context.Context, conversationID string, targetMessageID string) error {
	// 获取会话中的所有消息
	messages, err := s.messageRepo.FindByConversationID(ctx, conversationID)
	if err != nil {
		return fmt.Errorf("failed to get conversation messages: %w", err)
	}

	// 找到目标消息的位置
	targetIndex := -1
	for i, msg := range messages {
		if msg.ID == targetMessageID {
			targetIndex = i
			break
		}
	}

	if targetIndex == -1 {
		return fmt.Errorf("target message not found in conversation")
	}

	// 删除目标消息之后的所有消息
	for i := targetIndex + 1; i < len(messages); i++ {
		if err := s.messageRepo.Delete(ctx, messages[i].ID); err != nil {
			return err
		}
	}

	return nil
}

// ClearConversation 清空会话
func (s *MessageManager) ClearConversation(ctx context.Context, conversationID string) error {
	messages, err := s.messageRepo.FindByConversationID(ctx, conversationID)
	if err != nil {
		return fmt.Errorf("failed to get conversation messages: %w", err)
	}

	for _, msg := range messages {
		if err := s.messageRepo.Delete(ctx, msg.ID); err != nil {
			return err
		}
	}

	return nil
}

// GetConversationMessages 获取会话中的所有消息
func (s *MessageManager) GetConversationMessages(ctx context.Context, conversationID string) ([]*dto.MessageResponse, error) {
	messages, err := s.messageRepo.FindByConversationID(ctx, conversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation messages: %w", err)
	}

	responses := make([]*dto.MessageResponse, 0, len(messages))
	for _, msg := range messages {
		resp := dto.ToMessageResponse(msg)
		responses = append(responses, &resp)
	}

	return responses, nil
}

// GetMessageChain 获取消息链（从根消息到指定消息的完整路径）
func (s *MessageManager) GetMessageChain(ctx context.Context, messageID string) ([]*dto.MessageResponse, error) {
	result := []*dto.MessageResponse{}
	currentID := messageID

	for currentID != "" {
		msg, err := s.messageRepo.FindByID(ctx, currentID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				break
			}
			return nil, fmt.Errorf("failed to get message: %w", err)
		}

		resp := dto.ToMessageResponse(msg)
		result = append([]*dto.MessageResponse{&resp}, result...)
		currentID = msg.ParentID
	}

	return result, nil
}

// LinkMessageToNode 将消息关联到节点
func (s *MessageManager) LinkMessageToNode(ctx context.Context, nodeID string, messageID string, conversationType string) error {
	node, err := s.nodeRepo.FindByID(ctx, nodeID)
	if err != nil {
		return fmt.Errorf("failed to get node: %w", err)
	}

	// 根据对话类型更新节点的相应字段
	switch conversationType {
	case dto.ConversationTypeDecomposition:
		decomposition := node.Decomposition
		decomposition.LastMessageID = messageID
		node.Decomposition = decomposition
	case dto.ConversationTypeConclusion:
		conclusion := node.Conclusion
		conclusion.LastMessageID = messageID
		node.Conclusion = conclusion
	default:
		return fmt.Errorf("unsupported message type: %s", conversationType)
	}

	if err := s.nodeRepo.Update(ctx, node); err != nil {
		return fmt.Errorf("failed to update node: %w", err)
	}

	return nil
}

// GetNodeMessages 获取节点相关的消息
func (s *MessageManager) GetNodeMessages(ctx context.Context, nodeID string, conversationType string) ([]*dto.MessageResponse, error) {
	node, err := s.nodeRepo.FindByID(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}

	var lastMessageID string
	switch conversationType {
	case dto.ConversationTypeDecomposition:
		lastMessageID = node.Decomposition.LastMessageID
	case dto.ConversationTypeConclusion:
		lastMessageID = node.Conclusion.LastMessageID
	default:
		return nil, fmt.Errorf("unsupported conversation type: %s", conversationType)
	}

	if lastMessageID == "" {
		return []*dto.MessageResponse{}, nil
	}

	// 获取消息链
	return s.GetMessageChain(ctx, lastMessageID)
}

// UpdateNodeLastMessage 更新节点的最后消息ID
func (s *MessageManager) UpdateNodeLastMessage(ctx context.Context, nodeID string, messageID string, conversationType string) error {
	return s.LinkMessageToNode(ctx, nodeID, messageID, conversationType)
}

// MarkMessageAsDeleted 标记消息为已删除
func (s *MessageManager) MarkMessageAsDeleted(ctx context.Context, messageID string) error {
	return s.messageRepo.Delete(ctx, messageID)
}

// RestoreMessage 恢复已删除的消息（如果支持软删除）
func (s *MessageManager) RestoreMessage(ctx context.Context, messageID string) error {
	// 这里需要根据实际的软删除实现来处理
	// 当前的Delete方法可能是硬删除，需要根据实际情况调整
	return fmt.Errorf("restore message not implemented")
}

// GetMessageStatus 获取消息状态
func (s *MessageManager) GetMessageStatus(ctx context.Context, messageID string) (dto.MessageStatus, error) {
	msg, err := s.messageRepo.FindByID(ctx, messageID)
	if err != nil {
		return dto.MessageStatus{}, fmt.Errorf("failed to get message: %w", err)
	}

	status := dto.MessageStatus{
		ID:        msg.ID,
		Status:    "active",
		CreatedAt: msg.CreatedAt,
		UpdatedAt: msg.UpdatedAt,
	}

	if msg.DeletedAt.Valid {
		status.Status = "deleted"
		status.DeletedAt = &msg.DeletedAt.Time
	}

	return status, nil
}

// isZeroMessageContent 判断 MessageContent 是否为零值
func isZeroMessageContent(mc model.MessageContent) bool {
	return mc.Text == "" && len(mc.RAG) == 0 && len(mc.Notice) == 0
}
