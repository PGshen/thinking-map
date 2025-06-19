package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Notice 通知信息
type Notice struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

// MessageContent 消息内容
type MessageContent struct {
	Text   string   `json:"text,omitempty"`
	RAG    []string `json:"rag,omitempty"`
	Notice []Notice `json:"notice,omitempty"`
}

// Message 消息模型
type Message struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	NodeID      uuid.UUID `gorm:"type:uuid;not null;index"`
	ParentID    uuid.UUID `gorm:"type:uuid;index"`
	MessageType int       `gorm:"type:int;not null;default:1"` // 1:text, 2:image, 3:file, 4:link
	Content     string    `gorm:"type:text;not null"`
	Status      int       `gorm:"type:int;not null;default:1"`
	CreatedAt   time.Time `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	CreatedBy   uuid.UUID `gorm:"type:uuid;not null"`
	UpdatedBy   uuid.UUID `gorm:"type:uuid;not null"`
}

func (m *Message) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
