/*
 * @Date: 2025-06-19 00:09:56
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-06-22 21:37:47
 * @FilePath: /thinking-map/server/internal/model/dto/sse.go
 */
package dto

import (
	"time"

	"github.com/PGshen/thinking-map/server/internal/agent/base/multiagent"
	"github.com/PGshen/thinking-map/server/internal/model"
)

// 事件类型
const (
	ConnectionEstablishedEventType = "connectionEstablished"
	NodeCreatedEventType           = "nodeCreated"
	NodeUpdatedEventType           = "nodeUpdated"
	ThinkingProgressEventType      = "thinkingProgress"
	MessageTextEventType           = "messageText"
	MessageThoughtEventType        = "messageThought"
	MessageNoticeEventType         = "messageNotice"
	MessageActionEventType         = "messageAction"
	MessagePlanEventType           = "messagePlan"
	MessageRagEventType            = "messageRag"
	ErrorEventType                 = "error"
	CustomEventType                = "custom"
)

type ConnectionEstablishedEvent struct {
	SessionID string `json:"sessionID"`
	ClientID  string `json:"clientID"`
	Message   string `json:"message"`
}

// NodeCreatedEvent represents the node creation event
type NodeCreatedEvent struct {
	NodeID   string         `json:"nodeID"`
	ParentID string         `json:"parentID"`
	NodeType string         `json:"nodeType"`
	Question string         `json:"question"`
	Target   string         `json:"target"`
	Position model.Position `json:"position"`
}

// NodeUpdatedEvent represents the node update event
type NodeUpdatedEvent struct {
	NodeID  string                 `json:"nodeID"`
	Mode    string                 `json:"mode"` // 更新模式：repeace/append
	Updates map[string]interface{} `json:"updates"`
}

// ThinkingProgressEvent represents the thinking progress event
type ThinkingProgressEvent struct {
	NodeID   string `json:"nodeID"`
	Stage    string `json:"stage"`
	Progress int    `json:"progress"`
	Message  string `json:"message"`
}

type MessageActionEvent struct {
	NodeID    string         `json:"nodeID"`
	MessageID string         `json:"messageID"`
	Actions   []model.Action `json:"actions"`
}

// MsgTextEvent represents the text event
type MessageTextEvent struct {
	NodeID    string `json:"nodeID"`
	MessageID string `json:"messageID"`
	Message   string `json:"message"`
	Mode      string `json:"mode"`
}

// MsgThoughtEvent represents the thought event
type MessageThoughtEvent struct {
	NodeID    string `json:"nodeID"`
	MessageID string `json:"messageID"`
	Message   string `json:"message"`
	Mode      string `json:"mode"`
}

type MessagePlanEvent struct {
	NodeID    string              `json:"nodeID"`
	MessageID string              `json:"messageID"`
	Plan      multiagent.TaskPlan `json:"plan"`
	IsEnd     bool                `json:"isEnd"`
}

// ErrorEvent represents the error event
type ErrorEvent struct {
	NodeID       string `json:"nodeID"`
	ErrorCode    string `json:"errorCode"`
	ErrorMessage string `json:"errorMessage"`
}

// TestEventRequest represents the request for testing SSE events
type TestEventRequest struct {
	EventType string                 `json:"eventType" binding:"required,oneof=nodeCreated nodeUpdated thinkingProgress error custom"`
	Data      map[string]interface{} `json:"data" binding:"required"`
	Delay     int                    `json:"delay" binding:"min=0,max=10000"` // 延迟发送时间（毫秒）
}

// TestEventResponse represents the response for testing SSE events
type TestEventResponse struct {
	EventID   string    `json:"eventID"`
	EventType string    `json:"eventType"`
	SentAt    time.Time `json:"sentAt"`
	Message   string    `json:"message"`
}
