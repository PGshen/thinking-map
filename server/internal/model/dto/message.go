package dto

import (
	"time"

	"github.com/PGshen/thinking-map/server/internal/model"
	"github.com/cloudwego/eino/schema"
)

// CreateMessageRequest represents the request body for creating a message
type CreateMessageRequest struct {
	ID          string               `json:"ID"`
	ParentID    string               `json:"parentID" binding:"omitempty,uuid"`
	UserID      string               `json:"userID" binding:"omitempty,uuid"`
	MessageType model.MsgType        `json:"messageType" binding:"required,oneof=text rag notice action"`
	Role        schema.RoleType      `json:"role" binding:"required,oneof=system assistant user"`
	Content     model.MessageContent `json:"content" binding:"required"`
	Metadata    interface{}          `json:"metadata"`
}

// UpdateMessageRequest represents the request body for updating a message
type UpdateMessageRequest struct {
	ID          string               `json:"id" binding:"required,uuid"`
	MessageType model.MsgType        `json:"messageType" binding:"omitempty,oneof=text rag notice action"`
	Content     model.MessageContent `json:"content" binding:"omitempty"`
	Metadata    interface{}          `json:"metadata"`
}

// MessageResponse represents the message data in responses
type MessageResponse struct {
	ID             string               `json:"id"`
	ParentID       string               `json:"parentID"`
	ConversationID string               `json:"conversationID"`
	MessageType    model.MsgType        `json:"messageType"`
	Role           schema.RoleType      `json:"role"`
	Content        model.MessageContent `json:"content"`
	Metadata       interface{}          `json:"metadata"`
	CreatedAt      time.Time            `json:"createdAt"`
	UpdatedAt      time.Time            `json:"updatedAt"`
}

// MessageListResponse represents the paginated list of messages
type MessageListResponse struct {
	Total int               `json:"total"`
	Page  int               `json:"page"`
	Limit int               `json:"limit"`
	Items []MessageResponse `json:"items"`
}

// MessageStatus 消息状态
type MessageStatus struct {
	ID        string     `json:"id"`
	Status    string     `json:"status"` // active, deleted, archived
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt,omitempty"`
}

// ToMessageResponse 将 model.Message 转为 dto.MessageResponse
func ToMessageResponse(m *model.Message) MessageResponse {
	return MessageResponse{
		ID:             m.ID,
		ParentID:       m.ParentID,
		ConversationID: m.ConversationID,
		MessageType:    m.MessageType,
		Role:           m.Role,
		Content:        m.Content,
		Metadata:       m.Metadata,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}

// 用户选择消息
type ActionChoice struct {
	Introduction string   `json:"introduction"`
	Actions      []string `json:"actions"`
}

type ActionMsgResp []model.Action
