/*
 * @Date: 2025-06-19 00:09:56
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-06-22 21:37:47
 * @FilePath: /thinking-map/server/internal/model/dto/sse.go
 */
package dto

import (
	"time"

	"github.com/PGshen/thinking-map/server/internal/model"
)

// NodeCreatedEvent represents the node creation event
type NodeCreatedEvent struct {
	NodeID    string         `json:"nodeId"`
	ParentID  string         `json:"parentId"`
	NodeType  string         `json:"nodeType"`
	Question  string         `json:"question"`
	Target    string         `json:"target"`
	Position  model.Position `json:"position"`
	Timestamp time.Time      `json:"timestamp"`
}

// NodeUpdatedEvent represents the node update event
type NodeUpdatedEvent struct {
	NodeID    string                 `json:"nodeId"`
	Updates   map[string]interface{} `json:"updates"`
	Timestamp time.Time              `json:"timestamp"`
}

// ThinkingProgressEvent represents the thinking progress event
type ThinkingProgressEvent struct {
	NodeID    string    `json:"nodeId"`
	Stage     string    `json:"stage"`
	Progress  int       `json:"progress"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// ErrorEvent represents the error event
type ErrorEvent struct {
	NodeID       string    `json:"nodeId"`
	ErrorCode    string    `json:"errorCode"`
	ErrorMessage string    `json:"errorMessage"`
	Timestamp    time.Time `json:"timestamp"`
}

// TestEventRequest represents the request for testing SSE events
type TestEventRequest struct {
	EventType string                 `json:"eventType" binding:"required,oneof=node_created node_updated thinking_progress error custom"`
	Data      map[string]interface{} `json:"data" binding:"required"`
	Delay     int                    `json:"delay" binding:"min=0,max=10000"` // 延迟发送时间（毫秒）
}

// TestEventResponse represents the response for testing SSE events
type TestEventResponse struct {
	EventID   string    `json:"eventId"`
	EventType string    `json:"eventType"`
	SentAt    time.Time `json:"sentAt"`
	Message   string    `json:"message"`
}
