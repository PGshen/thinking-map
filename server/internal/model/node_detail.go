/*
 * @Date: 2025-06-18 22:25:59
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-06-23 22:58:11
 * @FilePath: /thinking-map/server/internal/model/node_detail.go
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

// NodeDetail 节点详情模型
type NodeDetail struct {
	SerialID   int64          `gorm:"primaryKey;autoIncrement;column:serial_id" json:"-"`
	ID         string         `gorm:"type:uuid;uniqueIndex"`
	NodeID     string         `gorm:"type:uuid;not null;index"`
	DetailType string         `gorm:"type:varchar(50);not null"`
	Content    DetailContent  `gorm:"type:jsonb;not null;default:'{}'"`
	Status     int            `gorm:"type:int;not null;default:1"`
	Metadata   datatypes.JSON `gorm:"type:jsonb;default:'{}'"`
	CreatedAt  time.Time      `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	UpdatedAt  time.Time      `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (n *NodeDetail) BeforeCreate(tx *gorm.DB) error {
	if n.ID == uuid.Nil.String() {
		n.ID = uuid.NewString()
	}
	return nil
}

func (NodeDetail) TableName() string {
	return "node_details"
}

// ContextInfo 上下文信息
type ContextInfo struct {
	NodeID     string `json:"node_id"`
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

// DetailContent 标签页内容
type DetailContent struct {
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

// DetailContent 实现 Scanner 接口
func (d *DetailContent) Scan(value interface{}) error {
	if value == nil {
		*d = DetailContent{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSON value: %v", value)
	}
	return json.Unmarshal(bytes, d)
}

// DetailContent 实现 Valuer 接口
func (d DetailContent) Value() (driver.Value, error) {
	return json.Marshal(d)
}
