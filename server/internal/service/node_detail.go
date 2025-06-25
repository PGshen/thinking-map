package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/PGshen/thinking-map/server/internal/model"
	"github.com/PGshen/thinking-map/server/internal/model/dto"
	"github.com/PGshen/thinking-map/server/internal/repository"
	"github.com/google/uuid"
)

type NodeDetailService struct {
	repo repository.NodeDetail
}

func NewNodeDetailService(repo repository.NodeDetail) *NodeDetailService {
	return &NodeDetailService{repo: repo}
}

// GetNodeDetails 获取节点详情列表
func (s *NodeDetailService) GetNodeDetails(ctx context.Context, nodeID string) ([]dto.NodeDetailResponse, error) {
	details, err := s.repo.FindByNodeID(ctx, nodeID)
	if err != nil {
		return nil, err
	}
	var res []dto.NodeDetailResponse
	for _, d := range details {
		var meta map[string]interface{}
		if d.Metadata != nil {
			_ = json.Unmarshal(d.Metadata, &meta)
		}
		res = append(res, dto.NodeDetailResponse{
			ID:         d.ID,
			NodeID:     d.NodeID,
			DetailType: d.DetailType,
			Content:    d.Content,
			Status:     d.Status,
			Metadata:   meta,
			CreatedAt:  d.CreatedAt,
			UpdatedAt:  d.UpdatedAt,
		})
	}
	return res, nil
}

// CreateNodeDetail 创建节点详情
func (s *NodeDetailService) CreateNodeDetail(ctx context.Context, nodeID string, req dto.CreateNodeDetailRequest) (*dto.NodeDetailResponse, error) {
	meta, _ := json.Marshal(req.Metadata)
	detail := &model.NodeDetail{
		ID:         uuid.NewString(),
		NodeID:     nodeID,
		DetailType: req.DetailType,
		Content:    req.Content,
		Status:     req.Status,
		Metadata:   meta,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if err := s.repo.Create(ctx, detail); err != nil {
		return nil, err
	}
	return &dto.NodeDetailResponse{
		ID:         detail.ID,
		NodeID:     detail.NodeID,
		DetailType: detail.DetailType,
		Content:    detail.Content,
		Status:     detail.Status,
		Metadata:   req.Metadata,
		CreatedAt:  detail.CreatedAt,
		UpdatedAt:  detail.UpdatedAt,
	}, nil
}

// UpdateNodeDetail 更新节点详情
func (s *NodeDetailService) UpdateNodeDetail(ctx context.Context, detailID string, req dto.UpdateNodeDetailRequest) (*dto.NodeDetailResponse, error) {
	detail, err := s.repo.FindByID(ctx, detailID)
	if err != nil {
		return nil, err
	}
	meta, _ := json.Marshal(req.Metadata)
	detail.Content = req.Content
	detail.Status = req.Status
	detail.Metadata = meta
	detail.UpdatedAt = time.Now()
	if err := s.repo.Update(ctx, detail); err != nil {
		return nil, err
	}
	return &dto.NodeDetailResponse{
		ID:         detail.ID,
		NodeID:     detail.NodeID,
		DetailType: detail.DetailType,
		Content:    detail.Content,
		Status:     detail.Status,
		Metadata:   req.Metadata,
		CreatedAt:  detail.CreatedAt,
		UpdatedAt:  detail.UpdatedAt,
	}, nil
}

// DeleteNodeDetail 删除节点详情
func (s *NodeDetailService) DeleteNodeDetail(ctx context.Context, detailID string) error {
	return s.repo.Delete(ctx, detailID)
}
