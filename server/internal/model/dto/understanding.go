package dto

type UnderstandingRequest struct {
	ParentMsgID   string `json:"parentMsgId"`
	Problem       string `json:"problem"`
	ProblemType   string `json:"problemType"`
	Supplementary string `json:"supplementary"`
}
