/*
 * @Date: 2025-06-18 22:26:28
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-06-23 22:58:20
 * @FilePath: /thinking-map/server/internal/model/rag_record.go
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

type RagSource string

const (
	RagTavily RagSource = "tavily"
)

// RAGRecord RAG 记录模型
type RAGRecord struct {
	SerialID  int64          `gorm:"primaryKey;autoIncrement;column:serial_id" json:"-"`
	ID        string         `gorm:"type:uuid;uniqueIndex" json:"id"`
	Query     string         `gorm:"type:text;not null" json:"query"`
	Answer    string         `gorm:"type:text;not null" json:"answer"`
	Sources   RagSource      `gorm:"type:varchar(128);not null" json:"sources"`
	Results   Results        `gorm:"type:jsonb;default:'[]'" json:"results"`
	Metadata  datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"metadata"`
	CreatedAt time.Time      `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (r *RAGRecord) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil.String() {
		r.ID = uuid.NewString()
	}
	return nil
}

func (RAGRecord) TableName() string {
	return "rag_records"
}

type Results []Result

type Result struct {
	Title      string  `json:"title"`
	URL        string  `json:"url"`
	Content    string  `json:"content"`
	Score      float64 `json:"score"`
	RawContent string  `json:"raw_content,omitempty"`
	Favicon    string  `json:"favicon,omitempty"`
}

// Scan implements the Scanner interface for Results
func (c *Results) Scan(value interface{}) error {
	if value == nil {
		*c = Results{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSON value: %v", value)
	}
	return json.Unmarshal(bytes, c)
}

// Value implements the Valuer interface for Results
func (c Results) Value() (driver.Value, error) {
	return json.Marshal(c)
}
