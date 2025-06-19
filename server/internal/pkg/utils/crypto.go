package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword 使用 bcrypt 对密码进行哈希
func HashPassword(password string) (string, error) {
	// 生成随机盐值
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// 使用 bcrypt 进行哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	// 将盐值和哈希值组合并编码为 base64
	combined := append(salt, hashedPassword...)
	return base64.StdEncoding.EncodeToString(combined), nil
}

// VerifyPassword 验证密码是否匹配
func VerifyPassword(hashedPassword, password string) bool {
	// 解码 base64 字符串
	decoded, err := base64.StdEncoding.DecodeString(hashedPassword)
	if err != nil {
		return false
	}

	// 提取哈希值（跳过盐值）
	hashedBytes := decoded[16:]

	// 验证密码
	err = bcrypt.CompareHashAndPassword(hashedBytes, []byte(password))
	return err == nil
}

// GenerateToken 生成随机令牌
func GenerateToken(length int) (string, error) {
	// 生成随机字节
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// 编码为 base64
	return base64.URLEncoding.EncodeToString(b), nil
}

// EncryptString 加密字符串
func EncryptString(text string, key []byte) (string, error) {
	// 生成随机 IV
	iv := make([]byte, 12)
	if _, err := rand.Read(iv); err != nil {
		return "", fmt.Errorf("failed to generate IV: %w", err)
	}

	// 使用 AES-GCM 加密
	// TODO: 实现 AES-GCM 加密
	return "", nil
}

// DecryptString 解密字符串
func DecryptString(ciphertext string, key []byte) (string, error) {
	// 使用 AES-GCM 解密
	// TODO: 实现 AES-GCM 解密
	return "", nil
}
