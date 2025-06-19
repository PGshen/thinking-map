/*
 * @Date: 2025-06-18 22:25:29
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-06-18 23:09:42
 * @FilePath: /thinking-map/server/internal/model/thinking_map.go
 */
package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ThinkingMap 思维导图模型
type ThinkingMap struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Title        string    `gorm:"type:varchar(100);not null"`
	Description  string    `gorm:"type:text"`
	RootQuestion string    `gorm:"type:text;not null"`
	Status       int       `gorm:"type:int;not null;default:1"`
	CreatedAt    time.Time `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt    time.Time `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	CreatedBy    uuid.UUID `gorm:"type:uuid;not null"`
	UpdatedBy    uuid.UUID `gorm:"type:uuid;not null"`
}

func (t *ThinkingMap) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}
