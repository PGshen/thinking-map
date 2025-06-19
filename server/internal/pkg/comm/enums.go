/*
 * @Date: 2025-06-18 23:15:18
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-06-18 23:37:54
 * @FilePath: /thinking-map/server/internal/pkg/comm/enums.go
 */
package comm

import "errors"

// 用户角色
const (
	RoleUser  = 1 // 普通用户
	RoleAdmin = 2 // 管理员
)

// 用户状态
const (
	UserStatusActive   = 1 // 正常
	UserStatusInactive = 0 // 禁用
)

// 思维导图状态
const (
	MapStatusActive   = 1 // 正常
	MapStatusArchived = 0 // 归档
	MapStatusDeleted  = 2 // 删除
)

// 思维节点类型
const (
	NodeTypeQuestion = 1 // 问题
	NodeTypeAnswer   = 2 // 答案
	NodeTypeIdea     = 3 // 想法
)

// 思维节点状态
const (
	NodeStatusActive   = 1 // 正常
	NodeStatusArchived = 0 // 归档
	NodeStatusDeleted  = 2 // 删除
)

// 节点详情类型
const (
	DetailTypeText  = 1 // 文本
	DetailTypeImage = 2 // 图片
	DetailTypeFile  = 3 // 文件
	DetailTypeLink  = 4 // 链接
)

// 节点详情状态
const (
	DetailStatusActive   = 1 // 正常
	DetailStatusArchived = 0 // 归档
	DetailStatusDeleted  = 2 // 删除
)

// 消息类型
const (
	MessageTypeText  = 1 // 文本
	MessageTypeImage = 2 // 图片
	MessageTypeFile  = 3 // 文件
	MessageTypeLink  = 4 // 链接
)

// 消息状态
const (
	MessageStatusActive   = 1 // 正常
	MessageStatusArchived = 0 // 归档
	MessageStatusDeleted  = 2 // 删除
)

// RAG 记录状态
const (
	RAGStatusActive   = 1 // 正常
	RAGStatusArchived = 0 // 归档
	RAGStatusDeleted  = 2 // 删除
)

var (
	ErrMessageNotFound     = errors.New("message not found")
	ErrNoPermission        = errors.New("no permission")
	ErrThinkingMapNotFound = errors.New("thinking map not found")
)
