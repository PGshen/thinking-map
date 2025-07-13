/*
 * @Date: 2025-06-18 23:59:38
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-06-26 00:03:55
 * @FilePath: /thinking-map/server/internal/model/dto/map.go
 */
package dto

import (
	"time"

	"encoding/json"

	"github.com/PGshen/thinking-map/server/internal/model"
)

// CreateMapRequest represents the request body for creating a mind map
type CreateMapRequest struct {
	Title       string            `json:"title" binding:"required,max=256"`
	Problem     string            `json:"problem" binding:"required,max=1000"`
	ProblemType string            `json:"problemType" binding:"max=50"`
	Target      string            `json:"target" binding:"max=1000"`
	KeyPoints   model.KeyPoints   `json:"keyPoints"`
	Constraints model.Constraints `json:"constraints"`
}

// UpdateMapRequest represents the request body for updating a mind map
type UpdateMapRequest struct {
	Status      int               `json:"status" binding:"oneof=0 1 2"`
	Title       string            `json:"title" binding:"max=256"`
	Problem     string            `json:"problem" binding:"max=1000"`
	ProblemType string            `json:"problemType" binding:"max=50"`
	Target      string            `json:"target" binding:"max=1000"`
	KeyPoints   model.KeyPoints   `json:"keyPoints"`
	Constraints model.Constraints `json:"constraints"`
	Conclusion  string            `json:"conclusion" binding:"max=1000"`
}

// MapResponse represents the mind map data in responses
type MapResponse struct {
	ID          string            `json:"id"`
	Status      string            `json:"status"`
	Title       string            `json:"title"`
	Problem     string            `json:"problem"`
	ProblemType string            `json:"problemType"`
	Target      string            `json:"target"`
	KeyPoints   model.KeyPoints   `json:"keyPoints"`
	Constraints model.Constraints `json:"constraints"`
	Conclusion  string            `json:"conclusion"`
	Progress    float64           `json:"progress"`
	Metadata    model.JSONB       `json:"metadata"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
}

// MapListResponse represents the paginated list of mind maps
type MapListResponse struct {
	Total int           `json:"total"`
	Page  int           `json:"page"`
	Limit int           `json:"limit"`
	Items []MapResponse `json:"items"`
}

// MapListQuery represents the query parameters for listing mind maps
type MapListQuery struct {
	Page        int    `form:"page" binding:"required,min=1"`
	Limit       int    `form:"limit" binding:"required,min=1,max=100"`
	Status      string `form:"status"`
	ProblemType string `form:"problemType"`
	DateRange   string `form:"dateRange"` // this-week, last-week, this-month, all-time
	Search      string `form:"search"`
}

// ToMapResponse converts a model.ThinkingMap to a MapResponse
func ToMapResponse(m *model.ThinkingMap) MapResponse {
	var meta model.JSONB
	if m.Metadata != nil {
		_ = json.Unmarshal(m.Metadata, &meta)
	}
	return MapResponse{
		ID:          m.ID,
		Status:      m.Status,
		Title:       m.Title,
		Problem:     m.Problem,
		ProblemType: m.ProblemType,
		Target:      m.Target,
		KeyPoints:   m.KeyPoints,
		Constraints: m.Constraints,
		Conclusion:  m.Conclusion,
		Metadata:    meta,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}
