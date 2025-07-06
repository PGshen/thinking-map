package thinking

import (
	"github.com/PGshen/thinking-map/server/internal/model/dto"
	"github.com/PGshen/thinking-map/server/internal/service"
	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// 问题理解
type UnderstandingHandler struct {
	understandingService *service.UnderstandingService
}

func NewUnderstandingHandler(us *service.UnderstandingService) *UnderstandingHandler {
	return &UnderstandingHandler{
		understandingService: us,
	}
}

func (h *UnderstandingHandler) Handle(c *gin.Context) (msgID string, event string, sr *schema.StreamReader[*schema.Message], err error) {
	var req dto.UnderstandingRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		return
	}
	// 生产msgID
	msgID = uuid.NewString()
	req.MsgID = msgID
	event, sr, err = h.understandingService.Understanding(c, req)
	return
}
