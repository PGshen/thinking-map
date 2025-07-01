package dto

type UnderstandingRequest struct {
	ParentMsgID   string `json:"parent_msg_id"`
	Problem       string `json:"problem"`
	ProblemType   string `json:"problem_type"`
	Supplementary string `json:"supplementary"`
}
