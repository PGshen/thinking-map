package dto

type UnderstandingRequest struct {
	MsgID         string `json:"msgID"`
	ParentMsgID   string `json:"parentMsgID"`
	Problem       string `json:"problem"`
	ProblemType   string `json:"problemType"`
	Supplementary string `json:"supplementary"`
}

type UnderstandingResponse struct {
	Title       string   `json:"title"`
	Problem     string   `json:"problem"`
	ProblemType string   `json:"problemType"`
	Goal        string   `json:"goal"`
	KeyPoints   []string `json:"keyPoints"`
	Constraints []string `json:"constraints"`
	Suggestion  string   `json:"suggestion"`
}
