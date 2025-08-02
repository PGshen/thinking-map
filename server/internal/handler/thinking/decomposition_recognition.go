package thinking

import (
	"github.com/PGshen/thinking-map/server/internal/model/dto"
	"github.com/PGshen/thinking-map/server/internal/service"
	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// DecompositionRecognitionHandler handles intent recognition HTTP requests
type DecompositionRecognitionHandler struct {
	intentService *service.IntentService
}

// NewDecompositionRecognitionHandler creates a new intent handler
func NewDecompositionRecognitionHandler(intentService *service.IntentService) *DecompositionRecognitionHandler {
	return &DecompositionRecognitionHandler{
		intentService: intentService,
	}
}

func (h *DecompositionRecognitionHandler) Handle(c *gin.Context) (msgID string, event string, sr *schema.StreamReader[*schema.Message], err error) {
	var req dto.DecompositionRecognitionRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		return
	}
	// 生成msgID
	msgID = uuid.NewString()
	req.MsgID = msgID
	event, sr, err = h.intentService.RecognizeDecomposition(c, req)
	return
}
