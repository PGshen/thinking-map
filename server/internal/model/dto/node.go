/*
 * @Date: 2025-06-19 00:02:51
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-06-22 21:39:39
 * @FilePath: /thinking-map/server/internal/model/dto/node.go
 */
package dto

import (
	"time"

	"github.com/PGshen/thinking-map/server/internal/model"
)

// CreateNodeRequest represents the request body for creating a node
type CreateNodeRequest struct {
	ParentID string         `json:"parent_id" binding:"required,uuid"`
	NodeType string         `json:"node_type" binding:"required"`
	Question string         `json:"question" binding:"required,max=500"`
	Target   string         `json:"target" binding:"max=500"`
	Position model.Position `json:"position" binding:"required"`
}

// UpdateNodeRequest represents the request body for updating a node
type UpdateNodeRequest struct {
	NodeID   string         `json:"id" binding:"required,uuid"`
	Question string         `json:"question" binding:"max=500"`
	Target   string         `json:"target" binding:"max=500"`
	Position model.Position `json:"position"`
}

// NodeResponse represents the node data in responses
type NodeResponse struct {
	ID           string             `json:"id"`
	MapID        string             `json:"map_id,omitempty"`
	ParentID     string             `json:"parent_id"`
	NodeType     string             `json:"node_type"`
	Question     string             `json:"question"`
	Target       string             `json:"target"`
	Context      string             `json:"context"`
	Status       int                `json:"status"`
	Position     model.Position     `json:"position"`
	Dependencies model.Dependencies `json:"dependencies"`
	NodeDetails  []model.NodeDetail `json:"node_details"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
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
}

// DependencyResponse represents the dependencies of a node
type DependencyResponse struct {
	Dependencies   []DependencyInfo `json:"dependencies"`
	DependentNodes []DependencyInfo `json:"dependent_nodes"`
}
