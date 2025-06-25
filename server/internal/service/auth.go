package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/PGshen/thinking-map/server/internal/model"
	"github.com/PGshen/thinking-map/server/internal/model/dto"
	"github.com/PGshen/thinking-map/server/internal/pkg/comm"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthService defines the interface for authentication operations
type AuthService interface {
	Register(ctx context.Context, req *dto.RegisterRequest) (*dto.AuthData, error)
	Login(ctx context.Context, req *dto.LoginRequest) (*dto.AuthData, error)
	RefreshToken(ctx context.Context, refreshToken string) (*dto.AuthData, error)
	Logout(ctx context.Context, accessToken string, refreshToken string) error
	ValidateToken(ctx context.Context, token string) (*model.TokenInfo, error)
}

// authService implements the AuthService interface
type authService struct {
	db    *gorm.DB
	redis *redis.Client
	jwt   JWTConfig
}

// JWTConfig holds the JWT configuration
type JWTConfig struct {
	SecretKey       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	TokenIssuer     string
}

// NewAuthService creates a new instance of AuthService
func NewAuthService(db *gorm.DB, redis *redis.Client, jwtConfig JWTConfig) AuthService {
	return &authService{
		db:    db,
		redis: redis,
		jwt:   jwtConfig,
	}
}

// Register implements user registration
func (s *authService) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.AuthData, error) {
	// Check if email exists
	var existingUser model.User
	if err := s.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return nil, comm.ErrUserAlreadyExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create user
	user := model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
		FullName: req.FullName,
		Status:   1,
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, err
	}

	// Generate tokens
	accessToken, refreshToken, err := s.generateTokens(user.ID, user.Username)
	if err != nil {
		return nil, err
	}

	// Store token info in Redis
	tokenInfo := model.TokenInfo{
		UserID:      user.ID,
		Username:    user.Username,
		AccessToken: accessToken,
		ExpiresAt:   time.Now().Add(s.jwt.AccessTokenTTL),
	}

	// Serialize token info to JSON
	tokenInfoJSON, err := json.Marshal(tokenInfo)
	if err != nil {
		return nil, err
	}

	if err := s.redis.Set(ctx, "token:"+accessToken, tokenInfoJSON, s.jwt.AccessTokenTTL).Err(); err != nil {
		return nil, err
	}

	return &dto.AuthData{
		UserID:       user.ID,
		Username:     user.Username,
		Email:        user.Email,
		FullName:     user.FullName,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(s.jwt.AccessTokenTTL.Seconds()),
	}, nil
}

// Login implements user login
func (s *authService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.AuthData, error) {
	var user model.User
	if err := s.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, comm.ErrInvalidCredentials
		}
		return nil, err
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, comm.ErrInvalidCredentials
	}

	// Generate tokens
	accessToken, refreshToken, err := s.generateTokens(user.ID, user.Username)
	if err != nil {
		return nil, err
	}

	// Store token info in Redis
	tokenInfo := model.TokenInfo{
		UserID:      user.ID,
		Username:    user.Username,
		AccessToken: accessToken,
		ExpiresAt:   time.Now().Add(s.jwt.AccessTokenTTL),
	}

	// Serialize token info to JSON
	tokenInfoJSON, err := json.Marshal(tokenInfo)
	if err != nil {
		return nil, err
	}

	if err := s.redis.Set(ctx, "token:"+accessToken, tokenInfoJSON, s.jwt.AccessTokenTTL).Err(); err != nil {
		return nil, err
	}

	return &dto.AuthData{
		UserID:       user.ID,
		Username:     user.Username,
		Email:        user.Email,
		FullName:     user.FullName,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(s.jwt.AccessTokenTTL.Seconds()),
	}, nil
}

// Logout implements user logout
func (s *authService) Logout(ctx context.Context, accessToken string, refreshToken string) error {
	// accessToken黑名单
	token, _ := jwt.ParseWithClaims(accessToken, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwt.SecretKey), nil
	})
	if token != nil && token.Valid {
		claims := token.Claims.(jwt.MapClaims)
		if jti, ok := claims["jti"].(string); ok {
			if exp, ok := claims["exp"].(float64); ok {
				expTime := time.Unix(int64(exp), 0)
				dur := time.Until(expTime)
				s.redis.Set(ctx, "blacklist:access:"+jti, 1, dur)
			}
		}
	}
	// refreshToken黑名单
	if refreshToken != "" {
		rt, _ := jwt.ParseWithClaims(refreshToken, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(s.jwt.SecretKey), nil
		})
		if rt != nil && rt.Valid {
			claims := rt.Claims.(jwt.MapClaims)
			if jti, ok := claims["jti"].(string); ok {
				if exp, ok := claims["exp"].(float64); ok {
					expTime := time.Unix(int64(exp), 0)
					dur := time.Until(expTime)
					s.redis.Set(ctx, "blacklist:refresh:"+jti, 1, dur)
				}
			}
		}
	}
	// 删除accessToken
	return s.redis.Del(ctx, "token:"+accessToken).Err()
}

// RefreshToken implements token refresh
func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*dto.AuthData, error) {
	// Parse and validate refresh token
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwt.SecretKey), nil
	})

	if err != nil || !token.Valid {
		return nil, comm.ErrInvalidToken
	}

	// 检查refreshToken是否在黑名单
	if jti, ok := claims["jti"].(string); ok {
		val, _ := s.redis.Get(ctx, "blacklist:refresh:"+jti).Result()
		if val != "" {
			return nil, comm.ErrInvalidToken
		}
	}

	// Get user info from claims
	userID, ok := claims["user_id"].(string)
	if !ok {
		return nil, comm.ErrInvalidToken
	}

	username, ok := claims["username"].(string)
	if !ok {
		return nil, comm.ErrInvalidToken
	}

	// Generate new tokens
	accessToken, newRefreshToken, err := s.generateTokens(userID, username)
	if err != nil {
		return nil, err
	}

	// Store new token info in Redis
	tokenInfo := model.TokenInfo{
		UserID:      userID,
		Username:    username,
		AccessToken: accessToken,
		ExpiresAt:   time.Now().Add(s.jwt.AccessTokenTTL),
	}

	// Serialize token info to JSON
	tokenInfoJSON, err := json.Marshal(tokenInfo)
	if err != nil {
		return nil, err
	}

	if err := s.redis.Set(ctx, "token:"+accessToken, tokenInfoJSON, s.jwt.AccessTokenTTL).Err(); err != nil {
		return nil, err
	}

	return &dto.AuthData{
		UserID:       userID,
		Username:     username,
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int(s.jwt.AccessTokenTTL.Seconds()),
	}, nil
}

// ValidateToken validates the access token
func (s *authService) ValidateToken(ctx context.Context, token string) (*model.TokenInfo, error) {
	// Get token info from Redis
	tokenInfoJSON, err := s.redis.Get(ctx, "token:"+token).Result()
	if err != nil {
		return nil, comm.ErrInvalidToken
	}

	// Deserialize token info from JSON
	var tokenInfo model.TokenInfo
	if err := json.Unmarshal([]byte(tokenInfoJSON), &tokenInfo); err != nil {
		return nil, comm.ErrInvalidToken
	}

	if time.Now().After(tokenInfo.ExpiresAt) {
		return nil, comm.ErrInvalidToken
	}

	return &tokenInfo, nil
}

// generateTokens generates access and refresh tokens
func (s *authService) generateTokens(userID, username string) (string, string, error) {
	// Generate access token
	accessTokenClaims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"exp":      time.Now().Add(s.jwt.AccessTokenTTL).Unix(),
		"iat":      time.Now().Unix(),
		"iss":      s.jwt.TokenIssuer,
		"type":     "access",
		"jti":      uuid.NewString(),
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)
	accessTokenString, err := accessToken.SignedString([]byte(s.jwt.SecretKey))
	if err != nil {
		return "", "", err
	}

	// Generate refresh token
	refreshTokenClaims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"exp":      time.Now().Add(s.jwt.RefreshTokenTTL).Unix(),
		"iat":      time.Now().Unix(),
		"iss":      s.jwt.TokenIssuer,
		"type":     "refresh",
		"jti":      uuid.NewString(),
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(s.jwt.SecretKey))
	if err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}
