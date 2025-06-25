/*
 * @Date: 2025-06-18 23:52:27
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-06-22 17:49:33
 * @FilePath: /thinking-map/server/internal/model/dto/auth.go
 */
package dto

import (
	"time"
)

// RegisterRequest represents the request body for user registration
type RegisterRequest struct {
	Username string `json:"username" binding:"min=3,max=32"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=32"`
	FullName string `json:"full_name" binding:"required,max=100"`
}

// LoginRequest represents the request body for user login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthData represents the data field in auth responses
type AuthData struct {
	UserID       string `json:"user_id,omitempty"`
	Username     string `json:"username,omitempty"`
	Email        string `json:"email,omitempty"`
	FullName     string `json:"full_name,omitempty"`
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in,omitempty"`
}

// ErrorData represents the error details in error responses
type ErrorData struct {
	Field string `json:"field,omitempty"`
	Error string `json:"error,omitempty"`
}

// TokenInfoDTO 用于返回 token 信息
// 与 model.TokenInfo 字段保持一致
type TokenInfoDTO struct {
	UserID      string    `json:"user_id"`
	Username    string    `json:"username"`
	AccessToken string    `json:"access_token"`
	ExpiresAt   time.Time `json:"expires_at"`
}
