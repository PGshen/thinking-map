package handler

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/google/uuid"
	"github.com/thinking-map/server/internal/model/dto"
	"github.com/thinking-map/server/internal/service"
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
func (h *AuthHandler) Register(ctx context.Context, c *app.RequestContext) {
	var req dto.RegisterRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(400, dto.Response{
			Code:      400,
			Message:   "invalid request parameters",
			Data:      dto.ErrorData{Error: err.Error()},
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}
	authData, err := h.authService.Register(ctx, &req)
	if err != nil {
		c.JSON(500, dto.Response{
			Code:      500,
			Message:   "failed to register user",
			Data:      dto.ErrorData{Error: err.Error()},
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}
	c.JSON(200, dto.Response{
		Code:      200,
		Message:   "success",
		Data:      authData,
		Timestamp: time.Now(),
		RequestID: uuid.New().String(),
	})
}

// Login handles user login
func (h *AuthHandler) Login(ctx context.Context, c *app.RequestContext) {
	var req dto.LoginRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(400, dto.Response{
			Code:      400,
			Message:   "invalid request parameters",
			Data:      dto.ErrorData{Error: err.Error()},
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}
	authData, err := h.authService.Login(ctx, &req)
	if err != nil {
		c.JSON(401, dto.Response{
			Code:      401,
			Message:   "invalid credentials",
			Data:      dto.ErrorData{Error: err.Error()},
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}
	c.JSON(200, dto.Response{
		Code:      200,
		Message:   "success",
		Data:      authData,
		Timestamp: time.Now(),
		RequestID: uuid.New().String(),
	})
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(ctx context.Context, c *app.RequestContext) {
	refreshToken := string(c.GetHeader("Authorization"))
	if refreshToken == "" {
		c.JSON(401, dto.Response{
			Code:      401,
			Message:   "invalid refresh token",
			Data:      nil,
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}
	authData, err := h.authService.RefreshToken(ctx, refreshToken)
	if err != nil {
		c.JSON(401, dto.Response{
			Code:      401,
			Message:   "invalid refresh token",
			Data:      dto.ErrorData{Error: err.Error()},
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}
	c.JSON(200, dto.Response{
		Code:      200,
		Message:   "success",
		Data:      authData,
		Timestamp: time.Now(),
		RequestID: uuid.New().String(),
	})
}

// Logout handles user logout
func (h *AuthHandler) Logout(ctx context.Context, c *app.RequestContext) {
	accessToken := string(c.GetHeader("Authorization"))
	if accessToken == "" {
		c.JSON(401, dto.Response{
			Code:      401,
			Message:   "invalid access token",
			Data:      nil,
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}
	if err := h.authService.Logout(ctx, accessToken); err != nil {
		c.JSON(500, dto.Response{
			Code:      500,
			Message:   "failed to logout",
			Data:      dto.ErrorData{Error: err.Error()},
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		})
		return
	}
	c.JSON(200, dto.Response{
		Code:      200,
		Message:   "success",
		Data:      nil,
		Timestamp: time.Now(),
		RequestID: uuid.New().String(),
	})
}
