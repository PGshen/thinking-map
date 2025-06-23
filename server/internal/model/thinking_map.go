/*
 * @Date: 2025-06-18 22:25:29
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-06-23 22:58:29
 * @FilePath: /thinking-map/server/internal/model/thinking_map.go
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

// ThinkingMap 思维导图模型
type ThinkingMap struct {
	SerialID    int64          `gorm:"primaryKey;autoIncrement;column:serial_id"`
	ID          string         `gorm:"type:uuid;uniqueIndex" json:"id"`
	UserID      string         `json:"user_id" gorm:"type:uuid;not null"`
	Problem     string         `json:"problem" gorm:"type:text;not null"`
	ProblemType string         `json:"problem_type" gorm:"type:varchar(50)"`
	Target      string         `json:"target" gorm:"type:text"`
	KeyPoints   KeyPoints      `json:"key_points" gorm:"type:jsonb"`
	Constraints Constraints    `json:"constraints" gorm:"type:jsonb"`
	Conclusion  string         `json:"conclusion" gorm:"type:text"`
	Status      int            `json:"status" gorm:"type:int;not null;default:1"`
	Metadata    datatypes.JSON `json:"metadata" gorm:"type:jsonb"`
	CreatedAt   time.Time      `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time      `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (t *ThinkingMap) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil.String() {
		t.ID = uuid.NewString()
	}
	return nil
}

// TableName 定义表名
func (ThinkingMap) TableName() string {
	return "thinking_maps"
}

type KeyPoints []string

// 实现Scanner接口
func (k *KeyPoints) Scan(value interface{}) error {
	if value == nil {
		*k = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSON value: %v", value)
	}
	return json.Unmarshal(bytes, k)
}

// 实现Valuer接口
func (k KeyPoints) Value() (driver.Value, error) {
	if len(k) == 0 {
		return nil, nil
	}
	return json.Marshal(k)
}

type Constraints []string

// 实现Scanner接口
func (c *Constraints) Scan(value interface{}) error {
	if value == nil {
		*c = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSON value: %v", value)
	}
	return json.Unmarshal(bytes, c)
}

// 实现Valuer接口
func (c Constraints) Value() (driver.Value, error) {
	if len(c) == 0 {
		return nil, nil
	}
	return json.Marshal(c)
}
