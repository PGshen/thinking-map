/*
 * @Date: 2025-06-18 22:25:44
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-06-23 22:58:36
 * @FilePath: /thinking-map/server/internal/model/thinking_node.go
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

// ThinkingNode 思维节点模型
type ThinkingNode struct {
	SerialID     int64            `gorm:"primaryKey;autoIncrement;column:serial_id" json:"-"`
	ID           string           `gorm:"type:uuid;uniqueIndex"`
	MapID        string           `gorm:"type:uuid;not null;index"`
	ParentID     string           `gorm:"type:uuid;index"`
	NodeType     string           `gorm:"type:varchar(50);not null"` // root, analysis, conclusion, custom
	Question     string           `gorm:"type:text;not null"`
	Target       string           `gorm:"type:text"`
	Context      DependentContext `gorm:"type:text;default:'{}'"` // 上下文
	Conclusion   string           `gorm:"type:text"`
	Status       int              `gorm:"type:int;default:0"` // 0:pending, 1:processing, 2:completed, -1:failed
	Position     Position         `gorm:"type:jsonb;default:'{\"x\":0,\"y\":0}'"`
	Metadata     datatypes.JSON   `gorm:"type:jsonb;default:'{}'"`
	Dependencies Dependencies     `gorm:"type:jsonb;default:'[]'"`
	CreatedAt    time.Time        `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	UpdatedAt    time.Time        `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	DeletedAt    gorm.DeletedAt   `gorm:"index" json:"-"`
}

func (t *ThinkingNode) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil.String() {
		t.ID = uuid.NewString()
	}
	return nil
}

func (ThinkingNode) TableName() string {
	return "thinking_nodes"
}

// Position 节点位置信息
type Position struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width,omitempty"`
	Height float64 `json:"height,omitempty"`
}

// Dependency 节点依赖信息
type Dependencies []string

// DependentContext 依赖上下文
type DependentContext struct {
	Ancestor    []NodeContext `json:"ancestor,omitempty"`    // 祖先节点
	PrevSibling []NodeContext `json:"prevSibling,omitempty"` // 前置兄弟节点
	Children    []NodeContext `json:"children,omitempty"`    // 子节点
}

// NodeContext 节点上下文
type NodeContext struct {
	Question   string `json:"question"`
	Target     string `json:"target"`
	Conclusion string `json:"conclusion"`
	Abstract   string `json:"abstract"`
}

// Position 实现 Scanner 接口
func (p *Position) Scan(value interface{}) error {
	if value == nil {
		*p = Position{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSON value: %v", value)
	}
	return json.Unmarshal(bytes, p)
}

// Position 实现 Valuer 接口
func (p Position) Value() (driver.Value, error) {
	return json.Marshal(p)
}

// Dependencies 实现 Scanner 接口
func (d *Dependencies) Scan(value interface{}) error {
	if value == nil {
		*d = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSON value: %v", value)
	}
	return json.Unmarshal(bytes, d)
}

// Dependencies 实现 Valuer 接口
func (d Dependencies) Value() (driver.Value, error) {
	if len(d) == 0 {
		return nil, nil
	}
	return json.Marshal(d)
}

// Scan implements the Scanner interface for Context
func (c *DependentContext) Scan(value interface{}) error {
	if value == nil {
		*c = DependentContext{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSON value: %v", value)
	}
	return json.Unmarshal(bytes, c)
}

// Value implements the Valuer interface for Context
func (c DependentContext) Value() (driver.Value, error) {
	return json.Marshal(c)
}
