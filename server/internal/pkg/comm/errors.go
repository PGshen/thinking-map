/*
 * @Date: 2025-06-18 23:01:52
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-06-24 09:13:17
 * @FilePath: /thinking-map/server/internal/pkg/comm/errors.go
 */
package comm

import "errors"

var (
	// 通用错误
	ErrNoPermission = errors.New("no permission")

	// 用户相关错误
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidToken       = errors.New("invalid token")

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
