package handler

import (
	"net/http"
	"time"

	"github.com/PGshen/thinking-map/server/internal/model/dto"
	"github.com/PGshen/thinking-map/server/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// NodeDetailHandler 节点详情相关接口
type NodeDetailHandler struct {
	NodeDetailService *service.NodeDetailService
}

func NewNodeDetailHandler(svc *service.NodeDetailService) *NodeDetailHandler {
	return &NodeDetailHandler{NodeDetailService: svc}
}

// GetNodeDetails handles GET /api/v1/nodes/{nodeID}/details
func (h *NodeDetailHandler) GetNodeDetails(c *gin.Context) {
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
	list, err := h.NodeDetailService.GetNodeDetails(c.Request.Context(), nodeID)
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
		Data:      dto.NodeDetailListResponse{Details: list},
		Timestamp: time.Now(),
		RequestID: uuid.New().String(),
	})
}

// CreateNodeDetail handles POST /api/v1/nodes/{nodeID}/details
func (h *NodeDetailHandler) CreateNodeDetail(c *gin.Context) {
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
	var req dto.CreateNodeDetailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Code:      http.StatusBadRequest,
			Message:   "invalid request parameters",
			Data:      nil,
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}
	resp, err := h.NodeDetailService.CreateNodeDetail(c.Request.Context(), nodeID, req)
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

// UpdateNodeDetail handles PUT /api/v1/node-details/{detailID}
func (h *NodeDetailHandler) UpdateNodeDetail(c *gin.Context) {
	detailID := c.Param("detailID")
	if detailID == "" {
		c.JSON(http.StatusBadRequest, dto.Response{
			Code:      http.StatusBadRequest,
			Message:   "detail ID is required",
			Data:      nil,
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}
	var req dto.UpdateNodeDetailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Code:      http.StatusBadRequest,
			Message:   "invalid request parameters",
			Data:      nil,
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}
	resp, err := h.NodeDetailService.UpdateNodeDetail(c.Request.Context(), detailID, req)
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

// DeleteNodeDetail handles DELETE /api/v1/node-details/{detailID}
func (h *NodeDetailHandler) DeleteNodeDetail(c *gin.Context) {
	detailID := c.Param("detailID")
	if detailID == "" {
		c.JSON(http.StatusBadRequest, dto.Response{
			Code:      http.StatusBadRequest,
			Message:   "detail ID is required",
			Data:      nil,
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}
	if err := h.NodeDetailService.DeleteNodeDetail(c.Request.Context(), detailID); err != nil {
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
