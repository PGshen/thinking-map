/*
 * @Date: 2025-06-19 00:10:27
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-06-19 09:48:34
 * @FilePath: /thinking-map/server/internal/handler/sse.go
 */
package handler

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/thinking-map/server/internal/pkg/sse"
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
func (h *SSEHandler) Connect(ctx context.Context, c *app.RequestContext) {
	mapID := c.Param("mapId")
	if mapID == "" {
		c.JSON(400, map[string]string{"error": "map ID is required"})
		return
	}
	c.Response.Header.Set("Content-Type", "text/event-stream")
	c.Response.Header.Set("Cache-Control", "no-cache")
	c.Response.Header.Set("Connection", "keep-alive")
	c.Response.Header.Set("Access-Control-Allow-Origin", "*")
	connectionID, eventChan := h.eventManager.Connect(mapID)
	for {
		select {
		case msg, ok := <-eventChan:
			if !ok {
				return
			}
			c.Write([]byte("event: message\ndata: " + string(msg) + "\n\n"))
			c.Flush()
		case <-ctx.Done():
			h.eventManager.Disconnect(mapID, connectionID)
			return
		}
	}
}
