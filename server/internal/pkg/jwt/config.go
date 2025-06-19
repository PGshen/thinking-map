package jwt

import "time"

// Config JWT 配置
type Config struct {
	SecretKey string        // JWT 密钥
	ExpiresIn time.Duration // token 过期时间
}
