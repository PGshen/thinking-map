package thinking

import (
	"github.com/PGshen/thinking-map/server/internal/model/dto"
	"github.com/PGshen/thinking-map/server/internal/service"
	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
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

func (h *UnderstandingHandler) Handle(c *gin.Context) (*schema.StreamReader[*schema.Message], error) {
	var req dto.UnderstandingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, err
	}
	return h.understandingService.Understanding(c, req)
}
