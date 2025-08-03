package dto

// DecompositionRequest represents the request for intent recognition
type DecompositionRequest struct {
	NodeID        string `json:"nodeID" binding:"required,uuid"`
	MsgID         string `json:"msgID"`
	IsDecompose   bool   `json:"isDecompose"`
	Clarification string `json:"clarification"`
}

// DecompositionResponse represents the response from intent recognition
type DecompositionResponse struct {
	IntentType   string   `json:"intentType"`   // decompose, conclude, explore, clarify
	Confidence   float64  `json:"confidence"`   // 0-1之间的置信度
	Reasoning    string   `json:"reasoning"`    // 识别理由
	Suggestion   string   `json:"suggestion"`   // 处理建议
	NextAction   string   `json:"nextAction"`   // 下一步行动
	RequiredInfo []string `json:"requiredInfo"` // 需要的额外信息
}
