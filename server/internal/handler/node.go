package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/google/uuid"
	"github.com/thinking-map/server/internal/model/dto"
)

type NodeHandler struct {
	// TODO: Add service dependencies here
}

func NewNodeHandler() *NodeHandler {
	return &NodeHandler{}
}

// ListNodes handles retrieving all nodes in a map
func (h *NodeHandler) ListNodes(ctx context.Context, c *app.RequestContext) {
	mapID := c.Param("mapId")
	if mapID == "" {
		c.JSON(http.StatusBadRequest, dto.Response{
			Code:      http.StatusBadRequest,
			Message:   "map ID is required",
			Data:      nil,
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}

	// TODO: Call service layer to get nodes
	// For now, return mock response
	c.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "success",
		Data: dto.NodeListResponse{
			Nodes: []dto.NodeResponse{
				{
					ID:        uuid.New().String(),
					MapID:     mapID,
					ParentID:  uuid.New().String(),
					NodeType:  "analysis",
					Question:  "Sample Question",
					Target:    "Sample Target",
					Context:   "Sample Context",
					Status:    0,
					Position:  dto.Position{X: 100, Y: 200},
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
		},
		Timestamp: time.Now(),
		RequestID: uuid.New().String(),
	})
}

// CreateNode handles creating a new node
func (h *NodeHandler) CreateNode(c *app.RequestContext) {
	mapID := c.Param("mapId")
	if mapID == "" {
		c.JSON(http.StatusBadRequest, dto.Response{
			Code:      http.StatusBadRequest,
			Message:   "map ID is required",
			Data:      nil,
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}

	var req dto.CreateNodeRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Code:      http.StatusBadRequest,
			Message:   "invalid request parameters",
			Data:      dto.ErrorData{Error: err.Error()},
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}

	// TODO: Call service layer to create node
	// For now, return mock response
	c.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "success",
		Data: dto.NodeResponse{
			ID:        uuid.New().String(),
			MapID:     mapID,
			ParentID:  req.ParentID,
			NodeType:  req.NodeType,
			Question:  req.Question,
			Target:    req.Target,
			Context:   req.Context,
			Status:    0,
			Position:  req.Position,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Timestamp: time.Now(),
		RequestID: uuid.New().String(),
	})
}

// UpdateNode handles updating a node
func (h *NodeHandler) UpdateNode(c *app.RequestContext) {
	nodeID := c.Param("nodeId")
	if nodeID == "" {
		c.JSON(http.StatusBadRequest, dto.Response{
			Code:      http.StatusBadRequest,
			Message:   "node ID is required",
			Data:      nil,
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}

	var req dto.UpdateNodeRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Code:      http.StatusBadRequest,
			Message:   "invalid request parameters",
			Data:      dto.ErrorData{Error: err.Error()},
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}

	// TODO: Call service layer to update node
	// For now, return mock response
	c.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "success",
		Data: dto.NodeResponse{
			ID:        nodeID,
			Question:  req.Question,
			Target:    req.Target,
			Context:   req.Context,
			Position:  req.Position,
			UpdatedAt: time.Now(),
		},
		Timestamp: time.Now(),
		RequestID: uuid.New().String(),
	})
}

// DeleteNode handles deleting a node
func (h *NodeHandler) DeleteNode(c *app.RequestContext) {
	nodeID := c.Param("nodeId")
	if nodeID == "" {
		c.JSON(http.StatusBadRequest, dto.Response{
			Code:      http.StatusBadRequest,
			Message:   "node ID is required",
			Data:      nil,
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}

	// TODO: Call service layer to delete node
	// For now, return success response
	c.JSON(http.StatusOK, dto.Response{
		Code:      http.StatusOK,
		Message:   "success",
		Data:      nil,
		Timestamp: time.Now(),
		RequestID: uuid.New().String(),
	})
}

// GetDependencies handles retrieving node dependencies
func (h *NodeHandler) GetDependencies(c *app.RequestContext) {
	nodeID := c.Param("nodeId")
	if nodeID == "" {
		c.JSON(http.StatusBadRequest, dto.Response{
			Code:      http.StatusBadRequest,
			Message:   "node ID is required",
			Data:      nil,
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}

	// TODO: Call service layer to get dependencies
	// For now, return mock response
	c.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "success",
		Data: dto.DependencyResponse{
			Dependencies: []dto.DependencyInfo{
				{
					NodeID:         uuid.New().String(),
					DependencyType: "prerequisite",
					Required:       true,
					Status:         2,
				},
			},
			DependentNodes: []dto.DependencyInfo{
				{
					NodeID:         uuid.New().String(),
					DependencyType: "dependent",
					Required:       true,
					Status:         0,
				},
			},
		},
		Timestamp: time.Now(),
		RequestID: uuid.New().String(),
	})
}
