package thinking

import (
	"github.com/PGshen/thinking-map/server/internal/model/dto"
	"github.com/PGshen/thinking-map/server/internal/service"
	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// IntentRecognitionHandler handles intent recognition HTTP requests
type IntentRecognitionHandler struct {
	intentService *service.IntentService
}

// NewIntentRecognitionHandler creates a new intent handler
func NewIntentRecognitionHandler(intentService *service.IntentService) *IntentRecognitionHandler {
	return &IntentRecognitionHandler{
		intentService: intentService,
	}
}

func (h *IntentRecognitionHandler) Handle(c *gin.Context) (msgID string, event string, sr *schema.StreamReader[*schema.Message], err error) {
	var req dto.IntentRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		return
	}
	// 生成msgID
	msgID = uuid.NewString()
	req.MsgID = msgID
	event, sr, err = h.intentService.RecognizeIntent(c, req)
	return
}
