package dto

import "time"

// CreateMapRequest represents the request body for creating a mind map
type CreateMapRequest struct {
	Title        string `json:"title" binding:"required,max=100"`
	Description  string `json:"description" binding:"max=500"`
	RootQuestion string `json:"root_question" binding:"required,max=200"`
}

// UpdateMapRequest represents the request body for updating a mind map
type UpdateMapRequest struct {
	Title       string `json:"title" binding:"max=100"`
	Description string `json:"description" binding:"max=500"`
	Status      int    `json:"status" binding:"oneof=0 1 2"`
}

// MapResponse represents the mind map data in responses
type MapResponse struct {
	ID           string                 `json:"id"`
	Title        string                 `json:"title"`
	Description  string                 `json:"description"`
	RootQuestion string                 `json:"root_question"`
	RootNodeID   string                 `json:"root_node_id,omitempty"`
	Status       int                    `json:"status"`
	Metadata     map[string]interface{} `json:"metadata"`
	NodeCount    int                    `json:"node_count,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
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
