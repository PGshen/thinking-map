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

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Message 消息模型
type Message struct {
	SerialID    int64          `gorm:"primaryKey;autoIncrement;column:serial_id" json:"-"`
	ID          string         `gorm:"type:uuid;uniqueIndex"`
	NodeID      string         `gorm:"type:uuid;not null;index"`
	ParentID    string         `gorm:"type:uuid;index"`
	MessageType string         `gorm:"type:varchar(20);not null;default:1"` // text, rag, notice
	Content     MessageContent `gorm:"type:jsonb;not null"`
	Metadata    datatypes.JSON `gorm:"type:jsonb;default:'{}'"`
	CreatedAt   time.Time      `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time      `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (m *Message) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil.String() {
		m.ID = uuid.NewString()
	}
	return nil
}

func (Message) TableName() string {
	return "messages"
}

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
