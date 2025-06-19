package errors

import "errors"

var (
	// 通用错误
	ErrNoPermission = errors.New("no permission")

	// 用户相关错误
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidPassword   = errors.New("invalid password")

	// 思维导图相关错误
	ErrThinkingMapNotFound = errors.New("thinking map not found")

	// 思维节点相关错误
	ErrThinkingNodeNotFound = errors.New("thinking node not found")

	// 节点详情相关错误
	ErrNodeDetailNotFound = errors.New("node detail not found")

	// 消息相关错误
	ErrMessageNotFound = errors.New("message not found")

	// RAG 相关错误
	ErrRAGRecordNotFound = errors.New("RAG record not found")
)
