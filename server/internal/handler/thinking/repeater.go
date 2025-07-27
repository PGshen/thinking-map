package thinking

import (
	"github.com/PGshen/thinking-map/server/internal/agent/llmmodel"
	"github.com/PGshen/thinking-map/server/internal/pkg/comm"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// 复读机
type RepeaterHandler struct{}

func NewRepeaterHandler() *RepeaterHandler {
	return &RepeaterHandler{}
}

func (r *RepeaterHandler) Handle(c *gin.Context) (msgID, event string, sr *schema.StreamReader[*schema.Message], err error) {
	var param map[string]any
	c.ShouldBindJSON(&param)
	message := param["message"].(string)
	var messageList []*schema.Message
	messageList, err = prompt.FromMessages(schema.Jinja2, schema.SystemMessage("你是一个复读机，用户输入什么，你就复读什么。除此之外不要尤其任何多余声明解释！")).Format(c, map[string]any{})
	if err != nil {
		return
	}
	messageList = append(messageList, &schema.Message{
		Role:    schema.User,
		Content: message,
	})
	var cm model.ToolCallingChatModel
	cm, err = llmmodel.NewOpenAIModel(c, nil)
	if err != nil {
		return
	}
	msgID = uuid.NewString()
	event = comm.EventText
	sr, err = cm.Stream(c, messageList)
	return
}
