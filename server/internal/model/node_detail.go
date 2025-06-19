package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ContextInfo 上下文信息
type ContextInfo struct {
	Type       string `json:"type"`
	Question   string `json:"question"`
	Target     string `json:"target"`
	Conclusion string `json:"conclusion"`
}

// DecomposeResult 分解结果
type DecomposeResult struct {
	Question string `json:"question"`
	Target   string `json:"target"`
}

// TabContent 标签页内容
type TabContent struct {
	// Info tab
	Context  []ContextInfo `json:"context,omitempty"`
	Question string        `json:"question,omitempty"`
	Target   string        `json:"target,omitempty"`

	// Decompose tab
	Message         [][]int           `json:"message,omitempty"`
	DecomposeResult []DecomposeResult `json:"decompose_result,omitempty"`

	// Conclusion tab
	Conclusion string `json:"conclusion,omitempty"`
}

// NodeDetail 节点详情模型
type NodeDetail struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	NodeID     uuid.UUID `gorm:"type:uuid;not null;index"`
	DetailType int       `gorm:"type:int;not null;default:1"` // 1:text, 2:image, 3:file, 4:link
	Content    JSONB     `gorm:"type:jsonb;not null;default:'{}'"`
	Status     int       `gorm:"type:int;not null;default:1"`
	CreatedAt  time.Time `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt  time.Time `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	CreatedBy  uuid.UUID `gorm:"type:uuid;not null"`
	UpdatedBy  uuid.UUID `gorm:"type:uuid;not null"`
}

func (n *NodeDetail) BeforeCreate(tx *gorm.DB) error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	return nil
}
