package handler

import (
	"net/http"

	"github.com/PGshen/thinking-map/server/internal/pkg/sse"
	"github.com/PGshen/thinking-map/server/internal/repository"
	"github.com/gin-gonic/gin"
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
	mapID := c.Param("mapId")
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

	// 以 userID 作为 clientID，或可自定义
	h.broker.HandleSSE(c, mapID, userIDStr)
}
