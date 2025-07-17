package dto

import (
	"encoding/json"
	"time"

	"github.com/PGshen/thinking-map/server/internal/model"
)

// CreateNodeDetailRequest 创建节点详情请求
// POST /api/v1/nodes/{nodeID}/details
// 只允许部分字段
// 参考API文档
// detail_type, content, status, metadata
// content结构体直接复用model.DetailContent
// metadata用map[string]interface{}即可

type CreateNodeDetailRequest struct {
	DetailType string                 `json:"detailType" binding:"required"`
	Content    model.DetailContent    `json:"content" binding:"required"`
	Status     int                    `json:"status"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// UpdateNodeDetailRequest 更新节点详情请求
// PUT /api/v1/node-details/{detailID}
type UpdateNodeDetailRequest struct {
	Content  model.DetailContent    `json:"content" binding:"required"`
	Status   int                    `json:"status"`
	Metadata map[string]interface{} `json:"metadata"`
}

// NodeDetailResponse 节点详情响应
// GET /api/v1/nodes/{nodeID}/details
// POST /api/v1/nodes/{nodeID}/details
// PUT /api/v1/node-details/{detailID}
type NodeDetailResponse struct {
	ID         string                 `json:"id"`
	NodeID     string                 `json:"nodeID"`
	DetailType string                 `json:"detailType"`
	Content    model.DetailContent    `json:"content"`
	Status     int                    `json:"status"`
	Metadata   map[string]interface{} `json:"metadata"`
	CreatedAt  time.Time              `json:"createdAt"`
	UpdatedAt  time.Time              `json:"updatedAt"`
}

// NodeDetailListResponse 节点详情列表响应
type NodeDetailListResponse struct {
	Details []NodeDetailResponse `json:"details"`
}

// ToNodeDetailResponse converts a model.NodeDetail to a NodeDetailResponse
func ToNodeDetailResponse(detail *model.NodeDetail) NodeDetailResponse {
	var meta map[string]interface{}
	if detail.Metadata != nil {
		_ = json.Unmarshal(detail.Metadata, &meta)
	}
	return NodeDetailResponse{
		ID:         detail.ID,
		NodeID:     detail.NodeID,
		DetailType: detail.DetailType,
		Content:    detail.Content,
		Status:     detail.Status,
		Metadata:   meta,
		CreatedAt:  detail.CreatedAt,
		UpdatedAt:  detail.UpdatedAt,
	}
}
