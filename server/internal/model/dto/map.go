/*
 * @Date: 2025-06-18 23:59:38
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-06-23 22:56:26
 * @FilePath: /thinking-map/server/internal/model/dto/map.go
 */
package dto

import (
	"time"

	"github.com/PGshen/thinking-map/server/internal/model"
)

// CreateMapRequest represents the request body for creating a mind map
type CreateMapRequest struct {
	Problem     string            `json:"problem" binding:"required,max=1000"`
	ProblemType string            `json:"problem_type" binding:"max=50"`
	Target      string            `json:"target" binding:"max=1000"`
	KeyPoints   model.KeyPoints   `json:"key_points"`
	Constraints model.Constraints `json:"constraints"`
	Conclusion  string            `json:"conclusion" binding:"max=1000"`
}

// UpdateMapRequest represents the request body for updating a mind map
type UpdateMapRequest struct {
	Status      int               `json:"status" binding:"oneof=0 1 2"`
	Problem     string            `json:"problem" binding:"max=1000"`
	ProblemType string            `json:"problem_type" binding:"max=50"`
	Target      string            `json:"target" binding:"max=1000"`
	KeyPoints   model.KeyPoints   `json:"key_points"`
	Constraints model.Constraints `json:"constraints"`
	Conclusion  string            `json:"conclusion" binding:"max=1000"`
}

// MapResponse represents the mind map data in responses
type MapResponse struct {
	ID          string            `json:"id"`
	RootNodeID  string            `json:"root_node_id,omitempty"`
	Status      int               `json:"status"`
	Problem     string            `json:"problem"`
	ProblemType string            `json:"problem_type"`
	Target      string            `json:"target"`
	KeyPoints   model.KeyPoints   `json:"key_points"`
	Constraints model.Constraints `json:"constraints"`
	Conclusion  string            `json:"conclusion"`
	Metadata    interface{}       `json:"metadata"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
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
	Page   int `form:"page" binding:"required,min=1"`
	Limit  int `form:"limit" binding:"required,min=1,max=100"`
	Status int `form:"status" binding:"omitempty,oneof=0 1 2"`
}
