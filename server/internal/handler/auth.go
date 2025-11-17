package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/PGshen/thinking-map/server/internal/model/dto"
	"github.com/PGshen/thinking-map/server/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
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

	authData, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Code:      http.StatusInternalServerError,
			Message:   err.Error(),
			Data:      dto.ErrorData{Error: err.Error()},
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:      http.StatusOK,
		Message:   "success",
		Data:      authData,
		Timestamp: time.Now(),
		RequestID: uuid.New().String(),
	})
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
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

	authData, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusOK, dto.Response{
			Code:      http.StatusBadRequest,
			Message:   "invalid email or password",
			Data:      dto.ErrorData{Error: err.Error()},
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:      http.StatusOK,
		Message:   "success",
		Data:      authData,
		Timestamp: time.Now(),
		RequestID: uuid.New().String(),
	})
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	refreshToken := c.GetHeader("X-Refresh-Token")
	if refreshToken == "" {
		c.JSON(http.StatusUnauthorized, dto.Response{
			Code:      http.StatusUnauthorized,
			Message:   "invalid refresh token",
			Data:      nil,
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}

	authData, err := h.authService.RefreshToken(c.Request.Context(), refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.Response{
			Code:      http.StatusUnauthorized,
			Message:   "invalid refresh token",
			Data:      dto.ErrorData{Error: err.Error()},
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:      http.StatusOK,
		Message:   "success",
		Data:      authData,
		Timestamp: time.Now(),
		RequestID: uuid.New().String(),
	})
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	accessTokenHeader := c.GetHeader("Authorization")
	refreshToken := c.GetHeader("X-Refresh-Token")
	if accessTokenHeader == "" || refreshToken == "" {
		c.JSON(http.StatusUnauthorized, dto.Response{
			Code:      http.StatusUnauthorized,
			Message:   "invalid access or refresh token",
			Data:      nil,
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}
	accessToken := ""
	if strings.HasPrefix(accessTokenHeader, "Bearer ") {
		accessToken = strings.TrimPrefix(accessTokenHeader, "Bearer ")
		accessToken = strings.TrimSpace(accessToken)
	} else {
		c.JSON(http.StatusUnauthorized, dto.Response{
			Code:      http.StatusUnauthorized,
			Message:   "invalid Authorization header format",
			Data:      nil,
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}

	if err := h.authService.Logout(c.Request.Context(), accessToken, refreshToken); err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Code:      http.StatusInternalServerError,
			Message:   "failed to logout",
			Data:      dto.ErrorData{Error: err.Error()},
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:      http.StatusOK,
		Message:   "success",
		Data:      nil,
		Timestamp: time.Now(),
		RequestID: uuid.New().String(),
	})
}
