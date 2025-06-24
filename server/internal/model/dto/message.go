package dto

import (
	"time"

	"github.com/PGshen/thinking-map/server/internal/model"
)

// CreateMessageRequest represents the request body for creating a message
type CreateMessageRequest struct {
	NodeID      string               `json:"node_id" binding:"required,uuid"`
	ParentID    string               `json:"parent_id" binding:"omitempty,uuid"`
	MessageType string               `json:"message_type" binding:"required,oneof=text rag notice"`
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
	NodeID      string               `json:"node_id"`
	ParentID    string               `json:"parent_id"`
	MessageType string               `json:"message_type"`
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
