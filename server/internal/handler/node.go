package handler

import (
	"net/http"
	"time"

	"github.com/PGshen/thinking-map/server/internal/model/dto"
	"github.com/PGshen/thinking-map/server/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// NodeHandler 节点相关接口
type NodeHandler struct {
	NodeService *service.NodeService
}

func NewNodeHandler(nodeService *service.NodeService) *NodeHandler {
	return &NodeHandler{NodeService: nodeService}
}

// ListNodes handles retrieving all nodes in a map
func (h *NodeHandler) ListNodes(c *gin.Context) {
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
	// 调用service层获取节点列表
	nodes, err := h.NodeService.ListNodes(c.Request.Context(), mapID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Code:      http.StatusInternalServerError,
			Message:   err.Error(),
			Data:      nil,
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}
	c.JSON(http.StatusOK, dto.Response{
		Code:      http.StatusOK,
		Message:   "success",
		Data:      dto.NodeListResponse{Nodes: nodes},
		Timestamp: time.Now(),
		RequestID: uuid.New().String(),
	})
}

// CreateNode handles creating a new node
func (h *NodeHandler) CreateNode(c *gin.Context) {
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
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Code:      http.StatusBadRequest,
			Message:   "invalid request parameters",
			Data:      dto.ErrorData{Error: err.Error()},
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}
	// TODO: 获取userID，暂用uuid.Nil
	userID := uuid.Nil.String()
	resp, err := h.NodeService.CreateNode(c.Request.Context(), mapID, req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Code:      http.StatusInternalServerError,
			Message:   err.Error(),
			Data:      nil,
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}
	c.JSON(http.StatusOK, dto.Response{
		Code:      http.StatusOK,
		Message:   "success",
		Data:      resp,
		Timestamp: time.Now(),
		RequestID: uuid.New().String(),
	})
}

// UpdateNode handles updating a node
func (h *NodeHandler) UpdateNode(c *gin.Context) {
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
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Code:      http.StatusBadRequest,
			Message:   "invalid request parameters",
			Data:      dto.ErrorData{Error: err.Error()},
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}

	resp, err := h.NodeService.UpdateNode(c.Request.Context(), nodeID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Code:      http.StatusInternalServerError,
			Message:   err.Error(),
			Data:      nil,
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}
	c.JSON(http.StatusOK, dto.Response{
		Code:      http.StatusOK,
		Message:   "success",
		Data:      resp,
		Timestamp: time.Now(),
		RequestID: uuid.New().String(),
	})
}

// DeleteNode handles deleting a node
func (h *NodeHandler) DeleteNode(c *gin.Context) {
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
	if err := h.NodeService.DeleteNode(c.Request.Context(), nodeID); err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Code:      http.StatusInternalServerError,
			Message:   err.Error(),
			Data:      nil,
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}
	c.JSON(http.StatusOK, dto.Response{
		Code:      http.StatusOK,
		Message:   "success",
		Data:      nil,
		Timestamp: time.Now(),
		RequestID: uuid.New().String(),
	})
}

// GetDependencies handles retrieving node dependencies
func (h *NodeHandler) GetDependencies(c *gin.Context) {
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
				},
			},
			DependentNodes: []dto.DependencyInfo{
				{
					NodeID:         uuid.New().String(),
					DependencyType: "dependent",
					Required:       true,
				},
			},
		},
		Timestamp: time.Now(),
		RequestID: uuid.New().String(),
	})
}
