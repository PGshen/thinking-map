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
	ParentID string         `json:"parentId" binding:"required,uuid"`
	NodeType string         `json:"nodeType" binding:"required"`
	Question string         `json:"question" binding:"required,max=500"`
	Target   string         `json:"target" binding:"max=500"`
	Position model.Position `json:"position" binding:"required"`
}

// UpdateNodeRequest represents the request body for updating a node
type UpdateNodeRequest struct {
	Question string         `json:"question" binding:"max=500"`
	Target   string         `json:"target" binding:"max=500"`
	Position model.Position `json:"position"`
}

// NodeResponse represents the node data in responses
type NodeResponse struct {
	ID           string               `json:"id"`
	MapID        string               `json:"mapId,omitempty"`
	ParentID     string               `json:"parentId"`
	NodeType     string               `json:"nodeType"`
	Question     string               `json:"question"`
	Target       string               `json:"target"`
	Context      string               `json:"context"`
	Status       int                  `json:"status"`
	Position     model.Position       `json:"position"`
	Dependencies model.Dependencies   `json:"dependencies"`
	NodeDetails  []NodeDetailResponse `json:"nodeDetails"`
	Metadata     interface{}          `json:"metadata"`
	CreatedAt    time.Time            `json:"createdAt"`
	UpdatedAt    time.Time            `json:"updatedAt"`
}

// NodeListResponse represents the list of nodes in a map
type NodeListResponse struct {
	Nodes []NodeResponse `json:"nodes"`
}

// DependencyInfo represents information about a node dependency
type DependencyInfo struct {
	NodeID         string `json:"nodeId"`
	DependencyType string `json:"dependencyType"`
	Required       bool   `json:"required"`
}

// DependencyResponse represents the dependencies of a node
type DependencyResponse struct {
	Dependencies   []DependencyInfo `json:"dependencies"`
	DependentNodes []DependencyInfo `json:"dependentNodes"`
}

// AddDependencyRequest represents the request body for adding a dependency
type AddDependencyRequest struct {
	DependencyNodeID string `json:"dependencyNodeId" binding:"required,uuid"`
	DependencyType   string `json:"dependencyType" binding:"required,oneof=prerequisite dependent"`
	Required         bool   `json:"required"`
}

// modelToNodeResponse 将model.ThinkingNode转为dto.NodeResponse
func ToNodeResponse(n *model.ThinkingNode) NodeResponse {
	return NodeResponse{
		ID:           n.ID,
		MapID:        n.MapID,
		ParentID:     n.ParentID,
		NodeType:     n.NodeType,
		Question:     n.Question,
		Target:       n.Target,
		Context:      n.Context,
		Status:       n.Status,
		Position:     n.Position,
		Dependencies: n.Dependencies,
		Metadata:     n.Metadata,
		CreatedAt:    n.CreatedAt,
		UpdatedAt:    n.UpdatedAt,
	}
}
