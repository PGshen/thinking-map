package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/thinking-map/server/internal/model/dto"
	"github.com/thinking-map/server/internal/service"
)

type MapHandler struct {
	mapService *service.MapService
}

func NewMapHandler(mapService *service.MapService) *MapHandler {
	return &MapHandler{
		mapService: mapService,
	}
}

// CreateMap handles the creation of a new mind map
func (h *MapHandler) CreateMap(c *gin.Context) {
	var req dto.CreateMapRequest
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

	// Get user ID from context (assuming it's set by auth middleware)
	userID, _ := c.Get("user_id")

	// Call service to create map
	mapResponse, err := h.mapService.CreateMap(c.Request.Context(), req, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Code:      http.StatusInternalServerError,
			Message:   "failed to create map",
			Data:      dto.ErrorData{Error: err.Error()},
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:      http.StatusOK,
		Message:   "success",
		Data:      mapResponse,
		Timestamp: time.Now(),
		RequestID: uuid.New().String(),
	})
}

// ListMaps handles retrieving a list of mind maps
func (h *MapHandler) ListMaps(c *gin.Context) {
	var query dto.MapListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Code:      http.StatusBadRequest,
			Message:   "invalid query parameters",
			Data:      dto.ErrorData{Error: err.Error()},
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}

	// Get user ID from context
	userID, _ := c.Get("user_id")

	// Call service to list maps
	maps, err := h.mapService.ListMaps(c.Request.Context(), query, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Code:      http.StatusInternalServerError,
			Message:   "failed to list maps",
			Data:      dto.ErrorData{Error: err.Error()},
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:      http.StatusOK,
		Message:   "success",
		Data:      maps,
		Timestamp: time.Now(),
		RequestID: uuid.New().String(),
	})
}

// GetMap handles retrieving a specific mind map
func (h *MapHandler) GetMap(c *gin.Context) {
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

	// Call service to get map
	mapResponse, err := h.mapService.GetMap(c.Request.Context(), mapID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Code:      http.StatusInternalServerError,
			Message:   "failed to get map",
			Data:      dto.ErrorData{Error: err.Error()},
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:      http.StatusOK,
		Message:   "success",
		Data:      mapResponse,
		Timestamp: time.Now(),
		RequestID: uuid.New().String(),
	})
}

// UpdateMap handles updating a mind map
func (h *MapHandler) UpdateMap(c *gin.Context) {
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

	var req dto.UpdateMapRequest
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

	// Get user ID from context
	userID, _ := c.Get("user_id")

	// Call service to update map
	mapResponse, err := h.mapService.UpdateMap(c.Request.Context(), mapID, req, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Code:      http.StatusInternalServerError,
			Message:   "failed to update map",
			Data:      dto.ErrorData{Error: err.Error()},
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:      http.StatusOK,
		Message:   "success",
		Data:      mapResponse,
		Timestamp: time.Now(),
		RequestID: uuid.New().String(),
	})
}

// DeleteMap handles deleting a mind map
func (h *MapHandler) DeleteMap(c *gin.Context) {
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

	// Get user ID from context
	userID, _ := c.Get("user_id")

	// Call service to delete map
	if err := h.mapService.DeleteMap(c.Request.Context(), mapID, userID.(string)); err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Code:      http.StatusInternalServerError,
			Message:   "failed to delete map",
			Data:      dto.ErrorData{Error: err.Error()},
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
