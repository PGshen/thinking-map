package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RAGRecord RAG 记录模型
type RAGRecord struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	NodeID    uuid.UUID `gorm:"type:uuid;not null;index"`
	Query     string    `gorm:"type:text;not null"`
	Answer    string    `gorm:"type:text;not null"`
	Sources   JSONB     `gorm:"type:jsonb;not null;default:'[]'"`
	Status    int       `gorm:"type:int;not null;default:1"`
	CreatedAt time.Time `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	CreatedBy uuid.UUID `gorm:"type:uuid;not null"`
	UpdatedBy uuid.UUID `gorm:"type:uuid;not null"`
}

func (r *RAGRecord) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}
