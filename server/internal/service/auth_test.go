package service

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/thinking-map/server/internal/model/dto"
)

var (
	testDB    *gorm.DB
	testRedis *redis.Client
	authSvc   AuthService
)

func TestMain(m *testing.M) {
	// 设置测试环境
	testConfig, err := SetupTestEnvironment()
	if err != nil {
		panic(fmt.Sprintf("failed to setup test environment: %v", err))
	}

	// 设置全局变量
	testDB = testConfig.DB
	testRedis = testConfig.Redis

	// 解析 JWT 配置
	cfg, _ := LoadTestConfig()
	expireDuration, err := time.ParseDuration(cfg.JWT.Expire)
	if err != nil {
		expireDuration = time.Minute * 10 // 默认值
	}

	// 初始化 AuthService
	authSvc = NewAuthService(testDB, testRedis, JWTConfig{
		SecretKey:       cfg.JWT.Secret,
		AccessTokenTTL:  expireDuration,
		RefreshTokenTTL: expireDuration * 2,
		TokenIssuer:     "test",
	})

	code := m.Run()

	// 清理测试数据
	CleanupTestEnvironment(testConfig)

	os.Exit(code)
}

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
		Username: username,
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
	err = authSvc.Logout(ctx, loginData.AccessToken)
	assert.NoError(t, err)

	// 校验 token 失效
	_, err = authSvc.ValidateToken(ctx, loginData.AccessToken)
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
