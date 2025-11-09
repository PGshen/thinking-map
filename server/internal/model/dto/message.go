package dto

import (
	"fmt"
	"strings"
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
	ID             string          `json:"id"`
	ParentID       string          `json:"parentID"`
	ConversationID string          `json:"conversationID"`
	MessageType    model.MsgType   `json:"messageType"`
	Role           schema.RoleType `json:"role"`
	Content        MessageContent  `json:"content"`
	Metadata       interface{}     `json:"metadata"`
	CreatedAt      time.Time       `json:"createdAt"`
	UpdatedAt      time.Time       `json:"updatedAt"`
}

type MessageContent struct {
	Text      string           `json:"text,omitempty"`
	Thought   string           `json:"thought,omitempty"`
	RagRecord *model.RAGRecord `json:"rag,omitempty"`
	Notice    *model.Notice    `json:"notice,omitempty"`
	Action    []model.Action   `json:"action,omitempty"`
	Plan      *model.Plan      `json:"plan,omitempty"`
}

// String()
func (m MessageContent) String() string {
	contentList := []string{}
	if m.Text != "" {
		contentList = append(contentList, m.Text)
	}
	if m.Thought != "" {
		contentList = append(contentList, fmt.Sprintf("\n思考：%s", m.Thought))
	}
	if m.Notice != nil && m.Notice.Content != "" {
		contentList = append(contentList, fmt.Sprintf("\n通知：%s：%s", m.Notice.Name, m.Notice.Content))
	}
	if m.Plan != nil && len(m.Plan.Steps) > 0 {
		for _, step := range m.Plan.Steps {
			if step.Status == "running" {
				contentList = append(contentList, fmt.Sprintf("\n计划：%s, %s", step.Name, step.Description))
			}
		}
	}
	if m.RagRecord != nil {
		contentList = append(contentList, fmt.Sprintf("\n检索结果：%s", m.RagRecord.Answer))
	}
	return strings.Join(contentList, "\n")
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
func ToMessageResponse(m *model.Message, rag *model.RAGRecord) MessageResponse {
	content := MessageContent{
		Text:    m.Content.Text,
		Thought: m.Content.Thought,
		Notice:  m.Content.Notice,
		Action:  m.Content.Action,
		Plan:    m.Content.Plan,
	}
	if rag != nil {
		content.RagRecord = rag
	}
	return MessageResponse{
		ID:             m.ID,
		ParentID:       m.ParentID,
		ConversationID: m.ConversationID,
		MessageType:    m.MessageType,
		Role:           m.Role,
		Content:        content,
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
