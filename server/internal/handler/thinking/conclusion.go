package thinking

import (
	"github.com/PGshen/thinking-map/server/internal/model/dto"
	"github.com/PGshen/thinking-map/server/internal/pkg/comm"
	"github.com/PGshen/thinking-map/server/internal/service"
	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ConclusionHandler handles conclusion generation HTTP requests
type ConclusionHandler struct {
	conclusionService *service.ConclusionService
}

// NewConclusionHandler creates a new conclusion handler
func NewConclusionHandler(conclusionService *service.ConclusionService) *ConclusionHandler {
	return &ConclusionHandler{
		conclusionService: conclusionService,
	}
}

func (h *ConclusionHandler) Handle(c *gin.Context) (msgID string, event string, sr *schema.StreamReader[*schema.Message], err error) {
	var req dto.ConclusionRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		return
	}
	msgID = uuid.New().String()
	event = comm.EventText
	sr, err = h.conclusionService.Conclusion(c, req)
	return
}
