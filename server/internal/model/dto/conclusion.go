package dto

type ConclusionRequest struct {
	NodeID      string `json:"nodeID" binding:"required,uuid"`
	Reference   string `json:"reference"`
	Instruction string `json:"instruction"`
}

type SaveConclusionRequest struct {
	Content string `json:"content"`
}

type ResetConclusionRequest struct {
	NodeID string `json:"nodeID" binding:"required,uuid"`
}
