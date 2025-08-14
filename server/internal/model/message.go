/*
 * @Date: 2025-06-18 22:26:13
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-06-23 22:43:13
 * @FilePath: /thinking-map/server/internal/model/message.go
 */
package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Message 消息模型
type Message struct {
	SerialID       int64           `gorm:"primaryKey;autoIncrement;column:serial_id" json:"-"`
	ID             string          `gorm:"type:uuid;uniqueIndex"`
	ParentID       string          `gorm:"type:uuid;index"`
	ConversationID string          `gorm:"type:uuid;index"`
	UserID         string          `json:"user_id" gorm:"type:uuid;not null"`
	MessageType    MsgType         `gorm:"type:varchar(20);not null;default:text"` // text, rag, notice, action
	Role           schema.RoleType `gorm:"type:varchar(48)"`
	Content        MessageContent  `gorm:"type:jsonb;not null"`
	Metadata       datatypes.JSON  `gorm:"type:jsonb;default:'{}'"`
	CreatedAt      time.Time       `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	UpdatedAt      time.Time       `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	DeletedAt      gorm.DeletedAt  `gorm:"index" json:"-"`
}

func (m *Message) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil.String() || m.ID == "" {
		m.ID = uuid.NewString()
	}
	return nil
}

func (Message) TableName() string {
	return "messages"
}

// 消息类型
type MsgType string

const (
	MsgTypeText   MsgType = "text"
	MsgTypeRAG    MsgType = "rag"
	MsgTypeNotice MsgType = "notice"
	MsgTypeAction MsgType = "action"
)

// Notice 通知信息
type Notice struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
}

type Action struct {
	Name   string         `json:"name"`
	URL    string         `json:"url"`
	Method string         `json:"method"`
	Param  map[string]any `json:"param,omitempty"`
}

// MessageContent 消息内容
type MessageContent struct {
	Text   string   `json:"text,omitempty"`
	RAG    []string `json:"rag,omitempty"`
	Notice []Notice `json:"notice,omitempty"`
	Action []Action `json:"action,omitempty"`
}

// MessageContent 实现 Scanner 接口
func (m *MessageContent) Scan(value interface{}) error {
	if value == nil {
		*m = MessageContent{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSON value: %v", value)
	}
	return json.Unmarshal(bytes, m)
}

// MessageContent 实现 Valuer 接口
func (m MessageContent) Value() (driver.Value, error) {
	return json.Marshal(m)
}

// String()
func (m MessageContent) String() string {
	if m.Text != "" {
		return m.Text
	}
	if len(m.Notice) > 0 {
		noticeStr := ""
		for _, notice := range m.Notice {
			noticeStr += fmt.Sprintf("%s: %s\n", notice.Type, notice.Content)
		}
		return noticeStr
	}
	return ""
}
