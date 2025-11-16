package thinking

import (
	"net/http"
	"time"

	"github.com/PGshen/thinking-map/server/internal/model/dto"
	"github.com/PGshen/thinking-map/server/internal/service"
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

func (h *ConclusionHandler) Handle(c *gin.Context) {
	var req dto.ConclusionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Code:      http.StatusBadRequest,
			Message:   "invalid request parameters",
			Data:      dto.ErrorData{Error: err.Error()},
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}
	h.conclusionService.Conclusion(c, req)
	// 响应
	c.JSON(http.StatusOK, dto.Response{
		Code:      http.StatusOK,
		Message:   "success",
		Data:      nil,
		Timestamp: time.Now(),
		RequestID: uuid.NewString(),
	})
}
