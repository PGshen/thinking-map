package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Position 节点位置信息
type Position struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width,omitempty"`
	Height float64 `json:"height,omitempty"`
}

// Dependency 节点依赖信息
type Dependency struct {
	NodeID         uuid.UUID `json:"node_id"`
	DependencyType string    `json:"dependency_type"`
	Required       bool      `json:"required"`
}

// ThinkingNode 思维节点模型
type ThinkingNode struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	MapID     uuid.UUID `gorm:"type:uuid;not null;index"`
	ParentID  uuid.UUID `gorm:"type:uuid;index"`
	NodeType  int       `gorm:"type:int;not null;default:1"` // 1:question, 2:answer, 3:idea
	Content   string    `gorm:"type:text;not null"`
	Position  JSONB     `gorm:"type:jsonb;not null;default:'{}'"`
	Status    int       `gorm:"type:int;not null;default:1"`
	CreatedAt time.Time `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	CreatedBy uuid.UUID `gorm:"type:uuid;not null"`
	UpdatedBy uuid.UUID `gorm:"type:uuid;not null"`
}

func (t *ThinkingNode) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}
