package conclusion

import "github.com/cloudwego/eino/schema"

type UserMessage struct {
	Reference   string            `json:"reference"`
	Instruction string            `json:"instruction"`
	Conclusion  string            `json:"conclusion"`
	History     []*schema.Message `json:"history"`
}
