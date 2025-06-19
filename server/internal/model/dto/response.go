/*
 * @Date: 2025-06-18 22:28:50
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-06-18 23:57:34
 * @FilePath: /thinking-map/server/internal/model/dto/response.go
 */
package dto

import (
	"time"
)

// 通用响应
type Response struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	RequestID string      `json:"request_id"`
}

// 分页响应
type PaginationResponse struct {
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
	Data     interface{} `json:"data"`
}
