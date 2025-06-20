package validator

import (
	"regexp"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// RegisterValidators 注册自定义验证器
func RegisterValidators() error {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// 注册自定义验证标签
		if err := v.RegisterValidation("username", validateUsername); err != nil {
			return err
		}
		if err := v.RegisterValidation("password", validatePassword); err != nil {
			return err
		}
		if err := v.RegisterValidation("node_type", validateNodeType); err != nil {
			return err
		}
		if err := v.RegisterValidation("tab_type", validateTabType); err != nil {
			return err
		}
		if err := v.RegisterValidation("message_type", validateMessageType); err != nil {
			return err
		}
		if err := v.RegisterValidation("role", validateRole); err != nil {
			return err
		}
		if err := v.RegisterValidation("status", validateStatus); err != nil {
			return err
		}
	}
	return nil
}

// validateUsername 验证用户名
// 规则：3-50个字符，只能包含字母、数字、下划线和连字符
func validateUsername(fl validator.FieldLevel) bool {
	username := fl.Field().String()
	if len(username) < 3 || len(username) > 50 {
		return false
	}
	matched, err := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, username)
	if err != nil {
		return false
	}
	return matched
}

// validatePassword 验证密码
// 规则：至少6个字符，必须包含大小写字母和数字
func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	if len(password) < 6 {
		return false
	}
	hasUpper := strings.ContainsAny(password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	hasLower := strings.ContainsAny(password, "abcdefghijklmnopqrstuvwxyz")
	hasNumber := strings.ContainsAny(password, "0123456789")
	return hasUpper && hasLower && hasNumber
}

// validateNodeType 验证节点类型
func validateNodeType(fl validator.FieldLevel) bool {
	nodeType := fl.Field().String()
	validTypes := map[string]bool{
		"root":       true,
		"analysis":   true,
		"conclusion": true,
		"custom":     true,
	}
	return validTypes[nodeType]
}

// validateTabType 验证标签页类型
func validateTabType(fl validator.FieldLevel) bool {
	tabType := fl.Field().String()
	validTypes := map[string]bool{
		"info":       true,
		"decompose":  true,
		"conclusion": true,
	}
	return validTypes[tabType]
}

// validateMessageType 验证消息类型
func validateMessageType(fl validator.FieldLevel) bool {
	messageType := fl.Field().String()
	validTypes := map[string]bool{
		"text":   true,
		"rag":    true,
		"notice": true,
	}
	return validTypes[messageType]
}

// validateRole 验证角色
func validateRole(fl validator.FieldLevel) bool {
	role := fl.Field().String()
	validRoles := map[string]bool{
		"user":      true,
		"assistant": true,
		"system":    true,
	}
	return validRoles[role]
}

// validateStatus 验证状态
func validateStatus(fl validator.FieldLevel) bool {
	status := fl.Field().Int()
	validStatus := map[int64]bool{
		-1: true, // failed
		0:  true, // pending/archived
		1:  true, // active/processing
		2:  true, // completed
	}
	return validStatus[status]
}
