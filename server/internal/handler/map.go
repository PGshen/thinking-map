package handler

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
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
func (h *MapHandler) CreateMap(ctx context.Context, c *app.RequestContext) {
	var req dto.CreateMapRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(400, dto.Response{
			Code:      400,
			Message:   "invalid request parameters",
			Data:      dto.ErrorData{Error: err.Error()},
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, dto.Response{
			Code:      401,
			Message:   "unauthorized",
			Data:      nil,
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}
	mapResponse, err := h.mapService.CreateMap(ctx, req, userID.(uuid.UUID))
	if err != nil {
		c.JSON(500, dto.Response{
			Code:      500,
			Message:   "failed to create map",
			Data:      dto.ErrorData{Error: err.Error()},
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}
	c.JSON(200, dto.Response{
		Code:      200,
		Message:   "success",
		Data:      mapResponse,
		Timestamp: time.Now(),
		RequestID: uuid.New().String(),
	})
}

// ListMaps handles retrieving a list of mind maps
func (h *MapHandler) ListMaps(c *app.RequestContext) {
	var query dto.MapListQuery
	if err := c.BindAndValidate(&query); err != nil {
		c.JSON(400, dto.Response{
			Code:      400,
			Message:   "invalid query parameters",
			Data:      dto.ErrorData{Error: err.Error()},
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, dto.Response{
			Code:      401,
			Message:   "unauthorized",
			Data:      nil,
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}
	maps, err := h.mapService.ListMaps(c.Request.Context(), query, userID.(uuid.UUID))
	if err != nil {
		c.JSON(500, dto.Response{
			Code:      500,
			Message:   "failed to list maps",
			Data:      dto.ErrorData{Error: err.Error()},
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}
	c.JSON(200, dto.Response{
		Code:      200,
		Message:   "success",
		Data:      maps,
		Timestamp: time.Now(),
		RequestID: uuid.New().String(),
	})
}

// GetMap handles retrieving a specific mind map
func (h *MapHandler) GetMap(c *app.RequestContext) {
	mapID := c.Param("mapId")
	if mapID == "" {
		c.JSON(400, dto.Response{
			Code:      400,
			Message:   "map ID is required",
			Data:      nil,
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}
	parsedMapID, err := uuid.Parse(mapID)
	if err != nil {
		c.JSON(400, dto.Response{
			Code:      400,
			Message:   "invalid map ID format",
			Data:      dto.ErrorData{Error: err.Error()},
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, dto.Response{
			Code:      401,
			Message:   "unauthorized",
			Data:      nil,
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}
	mapResponse, err := h.mapService.GetMap(c.Request.Context(), parsedMapID, userID.(uuid.UUID))
	if err != nil {
		c.JSON(500, dto.Response{
			Code:      500,
			Message:   "failed to get map",
			Data:      dto.ErrorData{Error: err.Error()},
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}
	c.JSON(200, dto.Response{
		Code:      200,
		Message:   "success",
		Data:      mapResponse,
		Timestamp: time.Now(),
		RequestID: uuid.New().String(),
	})
}

// UpdateMap handles updating a mind map
func (h *MapHandler) UpdateMap(c *app.RequestContext) {
	mapID := c.Param("mapId")
	if mapID == "" {
		c.JSON(400, dto.Response{
			Code:      400,
			Message:   "map ID is required",
			Data:      nil,
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}
	parsedMapID, err := uuid.Parse(mapID)
	if err != nil {
		c.JSON(400, dto.Response{
			Code:      400,
			Message:   "invalid map ID format",
			Data:      dto.ErrorData{Error: err.Error()},
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}
	var req dto.UpdateMapRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(400, dto.Response{
			Code:      400,
			Message:   "invalid request parameters",
			Data:      dto.ErrorData{Error: err.Error()},
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, dto.Response{
			Code:      401,
			Message:   "unauthorized",
			Data:      nil,
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}
	mapResponse, err := h.mapService.UpdateMap(c.Request.Context(), parsedMapID, req, userID.(uuid.UUID))
	if err != nil {
		c.JSON(500, dto.Response{
			Code:      500,
			Message:   "failed to update map",
			Data:      dto.ErrorData{Error: err.Error()},
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}
	c.JSON(200, dto.Response{
		Code:      200,
		Message:   "success",
		Data:      mapResponse,
		Timestamp: time.Now(),
		RequestID: uuid.New().String(),
	})
}

// DeleteMap handles deleting a mind map
func (h *MapHandler) DeleteMap(c *app.RequestContext) {
	mapID := c.Param("mapId")
	if mapID == "" {
		c.JSON(400, dto.Response{
			Code:      400,
			Message:   "map ID is required",
			Data:      nil,
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}
	parsedMapID, err := uuid.Parse(mapID)
	if err != nil {
		c.JSON(400, dto.Response{
			Code:      400,
			Message:   "invalid map ID format",
			Data:      dto.ErrorData{Error: err.Error()},
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, dto.Response{
			Code:      401,
			Message:   "unauthorized",
			Data:      nil,
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}
	if err := h.mapService.DeleteMap(c.Request.Context(), parsedMapID, userID.(uuid.UUID)); err != nil {
		c.JSON(500, dto.Response{
			Code:      500,
			Message:   "failed to delete map",
			Data:      dto.ErrorData{Error: err.Error()},
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}
	c.JSON(200, dto.Response{
		Code:      200,
		Message:   "success",
		Data:      nil,
		Timestamp: time.Now(),
		RequestID: uuid.New().String(),
	})
}
