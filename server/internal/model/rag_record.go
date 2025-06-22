/*
 * @Date: 2025-06-18 22:26:28
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-06-23 00:01:58
 * @FilePath: /thinking-map/server/internal/model/rag_record.go
 */
package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RAGRecord RAG 记录模型
type RAGRecord struct {
	SerialID  int64          `gorm:"primaryKey;autoIncrement;column:serial_id"`
	ID        uuid.UUID      `gorm:"type:uuid;uniqueIndex"`
	Query     string         `gorm:"type:text;not null"`
	Answer    string         `gorm:"type:text;not null"`
	Sources   JSONB          `gorm:"type:jsonb;not null;default:'[]'"`
	Status    int            `gorm:"type:int;not null;default:1"`
	CreatedAt time.Time      `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time      `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (r *RAGRecord) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

func (RAGRecord) TableName() string {
	return "rag_records"
}
