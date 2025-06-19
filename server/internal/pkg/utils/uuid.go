package utils

import (
	"github.com/google/uuid"
)

// NewUUID 生成新的 UUID
func NewUUID() string {
	return uuid.New().String()
}

// NewUUIDV4 生成新的 UUID v4
func NewUUIDV4() string {
	return uuid.New().String()
}

// ParseUUID 解析 UUID 字符串
func ParseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

// IsValidUUID 检查字符串是否是有效的 UUID
func IsValidUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}

// MustUUID 生成新的 UUID，如果失败则 panic
func MustUUID() string {
	return uuid.New().String()
}

// MustParseUUID 解析 UUID 字符串，如果失败则 panic
func MustParseUUID(s string) uuid.UUID {
	return uuid.Must(uuid.Parse(s))
}
