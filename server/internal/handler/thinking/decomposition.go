package thinking

import (
	"net/http"
	"time"

	"github.com/PGshen/thinking-map/server/internal/model/dto"
	"github.com/PGshen/thinking-map/server/internal/service"
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

func (h *DecompositionHandler) Handle(c *gin.Context) {
	var req dto.DecompositionRequest
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
	// 生成msgID
	h.decompositionService.Decomposition(c, req)
	// 响应
	c.JSON(http.StatusOK, dto.Response{
		Code:      http.StatusOK,
		Message:   "success",
		Data:      nil,
		Timestamp: time.Now(),
		RequestID: uuid.NewString(),
	})
}
