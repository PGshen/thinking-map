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
	NodeService    *service.NodeService
	MessageManager *service.MessageManager
}

func NewNodeHandler(nodeService *service.NodeService, messageManager *service.MessageManager) *NodeHandler {
	return &NodeHandler{NodeService: nodeService, MessageManager: messageManager}
}

// ListNodes handles retrieving all nodes in a map
func (h *NodeHandler) ListNodes(c *gin.Context) {
	mapID := c.Param("mapID")
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
	nodes, err := h.NodeService.ListNodes(c, mapID)
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
	mapID := c.Param("mapID")
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
	resp, err := h.NodeService.CreateNode(c.Request.Context(), mapID, req)
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
	nodeID := c.Param("nodeID")
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
func (h *NodeHandler) UpdateNodeContext(c *gin.Context) {
	nodeID := c.Param("nodeID")
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

	var req dto.UpdateNodeContextRequest
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

	resp, err := h.NodeService.UpdateNodeContext(c, nodeID, req)
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

// ResetNodeContext handles resetting a node's context
func (h *NodeHandler) ResetNodeContext(c *gin.Context) {
	nodeID := c.Param("nodeID")
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

	resp, err := h.NodeService.ResetNodeContext(c, nodeID)
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
	nodeID := c.Param("nodeID")
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

func (h *NodeHandler) GetNodeMessages(c *gin.Context) {
	nodeID := c.Param("nodeID")
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
	node, err := h.NodeService.GetNode(c.Request.Context(), nodeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Code:      http.StatusBadRequest,
			Message:   "node not found",
			Data:      nil,
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}
	conversationType := c.Query("conversationType")
	if conversationType == "" {
		c.JSON(http.StatusBadRequest, dto.Response{
			Code:      http.StatusBadRequest,
			Message:   "conversationType is required",
			Data:      nil,
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}
	var lastMessageID string
	if conversationType == dto.ConversationTypeDecomposition {
		lastMessageID = node.Decomposition.LastMessageID
	} else if conversationType == dto.ConversationTypeConclusion {
		lastMessageID = node.Conclusion.LastMessageID
	} else {
		c.JSON(http.StatusBadRequest, dto.Response{
			Code:      http.StatusBadRequest,
			Message:   "conversationType is invalid",
			Data:      nil,
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}

	messages, err := h.MessageManager.GetMessageChain(c.Request.Context(), lastMessageID)
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
		Data:      messages,
		Timestamp: time.Now(),
		RequestID: uuid.New().String(),
	})
}
