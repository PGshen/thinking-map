package dto

import "time"

// SSEConnectionResponse represents the SSE connection response
type SSEConnectionResponse struct {
	ConnectionID string    `json:"connection_id"`
	MapID        string    `json:"map_id"`
	Timestamp    time.Time `json:"timestamp"`
}

// SSEDisconnectionResponse represents the SSE disconnection response
type SSEDisconnectionResponse struct {
	ConnectionID string    `json:"connection_id"`
	Reason       string    `json:"reason"`
	Timestamp    time.Time `json:"timestamp"`
}

// NodeCreatedEvent represents the node creation event
type NodeCreatedEvent struct {
	NodeID    string    `json:"node_id"`
	ParentID  string    `json:"parent_id"`
	NodeType  string    `json:"node_type"`
	Question  string    `json:"question"`
	Target    string    `json:"target"`
	Position  Position  `json:"position"`
	Timestamp time.Time `json:"timestamp"`
}

// NodeUpdatedEvent represents the node update event
type NodeUpdatedEvent struct {
	NodeID    string                 `json:"node_id"`
	Updates   map[string]interface{} `json:"updates"`
	Timestamp time.Time              `json:"timestamp"`
}

// ThinkingProgressEvent represents the thinking progress event
type ThinkingProgressEvent struct {
	NodeID    string    `json:"node_id"`
	Stage     string    `json:"stage"`
	Progress  int       `json:"progress"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// ErrorEvent represents the error event
type ErrorEvent struct {
	NodeID       string    `json:"node_id"`
	ErrorCode    string    `json:"error_code"`
	ErrorMessage string    `json:"error_message"`
	Timestamp    time.Time `json:"timestamp"`
}
