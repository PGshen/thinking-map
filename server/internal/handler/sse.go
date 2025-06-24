package handler

import (
	"io"
	"net/http"

	"github.com/PGshen/thinking-map/server/internal/pkg/sse"

	"github.com/gin-gonic/gin"
)

type SSEHandler struct {
	eventManager *sse.EventManager
}

func NewSSEHandler(eventManager *sse.EventManager) *SSEHandler {
	return &SSEHandler{
		eventManager: eventManager,
	}
}

// Connect handles SSE connection requests
func (h *SSEHandler) Connect(c *gin.Context) {
	mapID := c.Param("mapId")
	if mapID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "map ID is required",
		})
		return
	}

	// Set headers for SSE
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// Create SSE connection
	connectionID, eventChan := h.eventManager.Connect(mapID)

	// Clean up connection when client disconnects
	c.Stream(func(w io.Writer) bool {
		select {
		case msg, ok := <-eventChan:
			if !ok {
				return false
			}
			c.SSEvent("message", string(msg))
			return true
		case <-c.Request.Context().Done():
			h.eventManager.Disconnect(mapID, connectionID)
			return false
		}
	})
}
