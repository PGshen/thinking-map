package thinking

import (
	"github.com/PGshen/thinking-map/server/internal/agent/llmmodel"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
)

// 复读机
type RepeaterHandler struct{}

func NewRepeaterHandler() *RepeaterHandler {
	return &RepeaterHandler{}
}

func (r *RepeaterHandler) Handle(c *gin.Context) (*schema.StreamReader[*schema.Message], error) {
	message := c.GetString("message")
	messageList, err := prompt.FromMessages(schema.Jinja2, schema.SystemMessage("你是一个复读机，用户输入什么，你就复读什么。除此之外不要尤其任何多余声明解释！")).Format(c, map[string]any{})
	if err != nil {
		return nil, err
	}
	messageList = append(messageList, &schema.Message{
		Role:    schema.User,
		Content: message,
	})
	cm, err := llmmodel.NewOpenAIModel(c, nil)
	if err != nil {
		return nil, err
	}
	return cm.Stream(c, messageList)
}
