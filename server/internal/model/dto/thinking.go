package dto

import "time"

// ThinkingOptions represents the options for AI thinking process
type ThinkingOptions struct {
	Model       string  `json:"model" binding:"required,oneof=gpt-4 gpt-3.5-turbo"`
	Temperature float64 `json:"temperature" binding:"required,min=0,max=1"`
}

// AnalyzeRequest represents the request body for starting problem analysis
type AnalyzeRequest struct {
	NodeID  string          `json:"nodeID" binding:"required,uuid"`
	Context string          `json:"context" binding:"required,max=2000"`
	Options ThinkingOptions `json:"options" binding:"required"`
}

// DecomposeRequest represents the request body for starting problem decomposition
type DecomposeRequest struct {
	NodeID            string `json:"nodeID" binding:"required,uuid"`
	DecomposeStrategy string `json:"decomposeStrategy" binding:"required,oneof=breadth_first depth_first"`
	MaxDepth          int    `json:"maxDepth" binding:"required,min=1,max=5"`
}

// ConcludeRequest represents the request body for generating conclusions
type ConcludeRequest struct {
	NodeID        string   `json:"nodeID" binding:"required,uuid"`
	Evidence      []string `json:"evidence" binding:"required,min=1"`
	ReasoningType string   `json:"reasoningType" binding:"required,oneof=deductive inductive abductive"`
}

// ChatRequest represents the request body for chat interaction
type ChatRequest struct {
	NodeID  string `json:"nodeID" binding:"required,uuid"`
	Message string `json:"message" binding:"required,max=1000"`
	Context string `json:"context" binding:"required,oneof=decompose conclude"`
}

// TaskResponse represents the response for async tasks
type TaskResponse struct {
	TaskID        string `json:"taskID"`
	NodeID        string `json:"nodeID"`
	Status        string `json:"status"`
	EstimatedTime int    `json:"estimatedTime"`
}

// ChatResponse represents the response for chat messages
type ChatResponse struct {
	MessageID string    `json:"messageID"`
	Content   string    `json:"content"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
}
