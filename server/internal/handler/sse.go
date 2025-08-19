package handler

import (
	"net/http"
	"time"

	"github.com/PGshen/thinking-map/server/internal/model/dto"
	"github.com/PGshen/thinking-map/server/internal/pkg/sse"
	"github.com/PGshen/thinking-map/server/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SSEHandler struct {
	broker  *sse.Broker
	mapRepo repository.ThinkingMap
}

func NewSSEHandler(broker *sse.Broker, mapRepo repository.ThinkingMap) *SSEHandler {
	return &SSEHandler{
		broker:  broker,
		mapRepo: mapRepo,
	}
}

// Connect handles SSE connection requests, with map ownership check
func (h *SSEHandler) Connect(c *gin.Context) {
	mapID := c.Param("mapID")
	if mapID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "map ID is required",
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
		return
	}

	mapObj, err := h.mapRepo.FindByID(c.Request.Context(), mapID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "map not found"})
		return
	}
	if mapObj.UserID != userIDStr {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden: map does not belong to user"})
		return
	}

	// mapID作为sessionID
	// 以 userID 作为 clientID，或可自定义
	h.broker.HandleSSE(c, mapID, userIDStr)
}

// SendEvent handles SSE test event requests
func (h *SSEHandler) SendEvent(c *gin.Context) {
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

	// 解析请求
	var req dto.TestEventRequest
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

	// 生成事件ID
	eventID := uuid.New().String()

	// 创建SSE事件
	event := sse.Event{
		ID:   eventID,
		Type: req.EventType,
		Data: req.Data,
	}

	// 如果有延迟，异步发送
	if req.Delay > 0 {
		go func() {
			time.Sleep(time.Duration(req.Delay) * time.Millisecond)
			h.broker.PublishToSession(mapID, event)
		}()
	} else {
		// 立即发送
		h.broker.PublishToSession(mapID, event)
	}

	// 返回响应
	response := dto.TestEventResponse{
		EventID:   eventID,
		EventType: req.EventType,
		SentAt:    time.Now(),
		Message:   "Test event sent successfully",
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:      http.StatusOK,
		Message:   "success",
		Data:      response,
		Timestamp: time.Now(),
		RequestID: uuid.New().String(),
	})
}
