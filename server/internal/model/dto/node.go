package dto

import "time"

// Position represents the x,y coordinates of a node
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// CreateNodeRequest represents the request body for creating a node
type CreateNodeRequest struct {
	ParentID string   `json:"parent_id" binding:"required,uuid"`
	NodeType string   `json:"node_type" binding:"required,oneof=analysis question target"`
	Question string   `json:"question" binding:"required,max=500"`
	Target   string   `json:"target" binding:"max=500"`
	Context  string   `json:"context" binding:"max=1000"`
	Position Position `json:"position" binding:"required"`
}

// UpdateNodeRequest represents the request body for updating a node
type UpdateNodeRequest struct {
	Question string   `json:"question" binding:"max=500"`
	Target   string   `json:"target" binding:"max=500"`
	Context  string   `json:"context" binding:"max=1000"`
	Position Position `json:"position"`
}

// NodeResponse represents the node data in responses
type NodeResponse struct {
	ID        string    `json:"id"`
	MapID     string    `json:"map_id,omitempty"`
	ParentID  string    `json:"parent_id"`
	NodeType  string    `json:"node_type"`
	Question  string    `json:"question"`
	Target    string    `json:"target"`
	Context   string    `json:"context"`
	Status    int       `json:"status"`
	Position  Position  `json:"position"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NodeListResponse represents the list of nodes in a map
type NodeListResponse struct {
	Nodes []NodeResponse `json:"nodes"`
}

// DependencyInfo represents information about a node dependency
type DependencyInfo struct {
	NodeID         string `json:"node_id"`
	DependencyType string `json:"dependency_type"`
	Required       bool   `json:"required"`
	Status         int    `json:"status"`
}

// DependencyResponse represents the dependencies of a node
type DependencyResponse struct {
	Dependencies   []DependencyInfo `json:"dependencies"`
	DependentNodes []DependencyInfo `json:"dependent_nodes"`
}
