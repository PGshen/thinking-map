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

const (
	ConversationTypeDecomposition = "decomposition"
	ConversationTypeConclusion    = "conclusion"
)

// CreateNodeRequest represents the request body for creating a node
type CreateNodeRequest struct {
	MapID    string         `json:"MapID" binding:"required,uuid"`
	ParentID string         `json:"parentID" binding:"required,uuid"`
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

type UpdateNodeContextRequest struct {
	Context model.DependentContext `json:"context"`
}

// NodeResponse represents the node data in responses
type NodeResponse struct {
	ID            string                 `json:"id"`
	NodeID        string                 `json:"nodeID"`
	MapID         string                 `json:"mapID,omitempty"`
	ParentID      string                 `json:"parentID"`
	NodeType      string                 `json:"nodeType"`
	Question      string                 `json:"question"`
	Target        string                 `json:"target"`
	Context       model.DependentContext `json:"context"`
	Decomposition model.Decomposition    `json:"decomposition"`
	Conclusion    model.Conclusion       `json:"conclusion"`
	Status        string                 `json:"status"`
	Position      model.Position         `json:"position"`
	Metadata      interface{}            `json:"metadata"`
	Dependencies  model.Dependencies     `json:"dependencies"`
	CreatedAt     time.Time              `json:"createdAt"`
	UpdatedAt     time.Time              `json:"updatedAt"`
}

// NodeListResponse represents the list of nodes in a map
type NodeListResponse struct {
	Nodes []NodeResponse `json:"nodes"`
}

// modelToNodeResponse 将model.ThinkingNode转为dto.NodeResponse
func ToNodeResponse(n *model.ThinkingNode) NodeResponse {
	return NodeResponse{
		ID:            n.ID,
		MapID:         n.MapID,
		ParentID:      n.ParentID,
		NodeType:      n.NodeType,
		Question:      n.Question,
		Target:        n.Target,
		Context:       n.Context,
		Decomposition: n.Decomposition,
		Conclusion:    n.Conclusion,
		Status:        n.Status,
		Position:      n.Position,
		Metadata:      n.Metadata,
		Dependencies:  n.Dependencies,
		CreatedAt:     n.CreatedAt,
		UpdatedAt:     n.UpdatedAt,
	}
}
