package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/PGshen/thinking-map/server/internal/model/dto"
)

func TestAuthService_Register_Login_Logout_ValidateToken(t *testing.T) {
	ctx := context.Background()
	email := fmt.Sprintf("test%d@example.com", time.Now().UnixNano())
	username := fmt.Sprintf("user%d", time.Now().UnixNano())
	password := "password123"
	fullName := "Test User"

	// 注册
	regReq := &dto.RegisterRequest{
		Username: username,
		Email:    email,
		Password: password,
		FullName: fullName,
	}
	authData, err := authSvc.Register(ctx, regReq)
	assert.NoError(t, err)
	assert.Equal(t, username, authData.Username)
	assert.Equal(t, email, authData.Email)
	assert.NotEmpty(t, authData.AccessToken)
	assert.NotEmpty(t, authData.RefreshToken)

	// 登录
	loginReq := &dto.LoginRequest{
		Email:    email,
		Password: password,
	}
	loginData, err := authSvc.Login(ctx, loginReq)
	assert.NoError(t, err)
	assert.Equal(t, username, loginData.Username)
	assert.Equal(t, email, loginData.Email)
	assert.NotEmpty(t, loginData.AccessToken)
	assert.NotEmpty(t, loginData.RefreshToken)

	// 校验 token
	tokenInfo, err := authSvc.ValidateToken(ctx, loginData.AccessToken)
	assert.NoError(t, err)
	assert.Equal(t, username, tokenInfo.Username)

	// 登出
	err = authSvc.Logout(ctx, loginData.AccessToken, loginData.RefreshToken)
	assert.NoError(t, err)

	// 校验 token 失效
	_, err = authSvc.ValidateToken(ctx, loginData.AccessToken)
	assert.Error(t, err)

	// refresh token 也不能再用
	_, err = authSvc.RefreshToken(ctx, loginData.RefreshToken)
	assert.Error(t, err)
}

func TestAuthService_RefreshToken(t *testing.T) {
	ctx := context.Background()
	email := fmt.Sprintf("test%d@example.com", time.Now().UnixNano())
	username := fmt.Sprintf("user%d", time.Now().UnixNano())
	password := "password123"
	fullName := "Test User"

	// 注册
	regReq := &dto.RegisterRequest{
		Username: username,
		Email:    email,
		Password: password,
		FullName: fullName,
	}
	authData, err := authSvc.Register(ctx, regReq)
	assert.NoError(t, err)

	// 刷新 token
	newAuthData, err := authSvc.RefreshToken(ctx, authData.RefreshToken)
	assert.NoError(t, err)
	assert.Equal(t, username, newAuthData.Username)
	assert.NotEqual(t, authData.AccessToken, newAuthData.AccessToken)
	assert.NotEqual(t, authData.RefreshToken, newAuthData.RefreshToken)
}
