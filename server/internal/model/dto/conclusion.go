package dto

type ConclusionRequest struct {
	NodeID      string `json:"nodeID" binding:"required,uuid"`
	Reference   string `json:"reference"`
	Instruction string `json:"instruction"`
}
