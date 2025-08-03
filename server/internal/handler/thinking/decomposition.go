package thinking

import (
	"github.com/PGshen/thinking-map/server/internal/model/dto"
	"github.com/PGshen/thinking-map/server/internal/service"
	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// DecompositionHandler handles intent recognition HTTP requests
type DecompositionHandler struct {
	decompositionService *service.DecompositionService
}

// NewDecompositionHandler creates a new intent handler
func NewDecompositionHandler(decompositionService *service.DecompositionService) *DecompositionHandler {
	return &DecompositionHandler{
		decompositionService: decompositionService,
	}
}

func (h *DecompositionHandler) Handle(c *gin.Context) (msgID string, event string, sr *schema.StreamReader[*schema.Message], err error) {
	var req dto.DecompositionRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		return
	}
	// 生成msgID
	msgID = uuid.NewString()
	req.MsgID = msgID
	event, sr, err = h.decompositionService.Decomposition(c, req)
	return
}
