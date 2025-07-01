package dto

import (
	"time"

	"github.com/PGshen/thinking-map/server/internal/model"
	"github.com/cloudwego/eino/schema"
)

// CreateMessageRequest represents the request body for creating a message
type CreateMessageRequest struct {
	ParentID    string               `json:"parent_id" binding:"omitempty,uuid"`
	MessageType string               `json:"message_type" binding:"required,oneof=text rag notice"`
	Role        schema.RoleType      `json:"role" binding:"required,oneof=system assistant user"`
	Content     model.MessageContent `json:"content" binding:"required"`
	Metadata    interface{}          `json:"metadata"`
}

// UpdateMessageRequest represents the request body for updating a message
type UpdateMessageRequest struct {
	ID          string               `json:"id" binding:"required,uuid"`
	MessageType string               `json:"message_type" binding:"omitempty,oneof=text rag notice"`
	Content     model.MessageContent `json:"content" binding:"omitempty"`
	Metadata    interface{}          `json:"metadata"`
}

// MessageResponse represents the message data in responses
type MessageResponse struct {
	ID          string               `json:"id"`
	ParentID    string               `json:"parent_id"`
	ChatID      string               `json:"chat_id"`
	MessageType string               `json:"message_type"`
	Role        schema.RoleType      `json:"role"`
	Content     model.MessageContent `json:"content"`
	Metadata    interface{}          `json:"metadata"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
}

// MessageListResponse represents the paginated list of messages
type MessageListResponse struct {
	Total int               `json:"total"`
	Page  int               `json:"page"`
	Limit int               `json:"limit"`
	Items []MessageResponse `json:"items"`
}

// ToMessageResponse 将 model.Message 转为 dto.MessageResponse
func ToMessageResponse(m *model.Message) MessageResponse {
	return MessageResponse{
		ID:          m.ID,
		ParentID:    m.ParentID,
		ChatID:      m.ChatID,
		MessageType: m.MessageType,
		Role:        m.Role,
		Content:     m.Content,
		Metadata:    m.Metadata,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}
