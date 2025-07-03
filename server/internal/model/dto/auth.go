/*
 * @Date: 2025-06-18 23:52:27
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-07-03 23:16:08
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
	FullName string `json:"fullName" binding:"required,max=100"`
}

// LoginRequest represents the request body for user login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthData represents the data field in auth responses
type AuthData struct {
	UserID       string `json:"userId,omitempty"`
	Username     string `json:"username,omitempty"`
	Email        string `json:"email,omitempty"`
	FullName     string `json:"fullName,omitempty"`
	AccessToken  string `json:"accessToken,omitempty"`
	RefreshToken string `json:"refreshToken,omitempty"`
	ExpiresIn    int    `json:"expiresIn,omitempty"`
}

// ErrorData represents the error details in error responses
type ErrorData struct {
	Field string `json:"field,omitempty"`
	Error string `json:"error,omitempty"`
}

// TokenInfoDTO 用于返回 token 信息
// 与 model.TokenInfo 字段保持一致
type TokenInfoDTO struct {
	UserID      string    `json:"userId"`
	Username    string    `json:"username"`
	AccessToken string    `json:"accessToken"`
	ExpiresAt   time.Time `json:"expiresAt"`
}
