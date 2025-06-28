package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/PGshen/thinking-map/server/internal/model"
	"github.com/PGshen/thinking-map/server/internal/model/dto"
	"github.com/PGshen/thinking-map/server/internal/repository"
	"github.com/google/uuid"
	"gorm.io/datatypes"
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
		res = append(res, dto.ToNodeDetailResponse(d))
	}
	return res, nil
}

// CreateNodeDetail 创建节点详情
func (s *NodeDetailService) CreateNodeDetail(ctx context.Context, nodeID string, req dto.CreateNodeDetailRequest) (*dto.NodeDetailResponse, error) {
	meta := datatypes.JSON{}
	if req.Metadata != nil {
		meta, _ = json.Marshal(req.Metadata)
	}
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
	resp := dto.ToNodeDetailResponse(detail)
	return &resp, nil
}

// UpdateNodeDetail 更新节点详情
func (s *NodeDetailService) UpdateNodeDetail(ctx context.Context, detailID string, req dto.UpdateNodeDetailRequest) (*dto.NodeDetailResponse, error) {
	detail, err := s.repo.FindByID(ctx, detailID)
	if err != nil {
		return nil, err
	}
	meta := datatypes.JSON{}
	if req.Metadata != nil {
		meta, _ = json.Marshal(req.Metadata)
	}
	detail.Content = req.Content
	detail.Status = req.Status
	detail.Metadata = meta
	detail.UpdatedAt = time.Now()
	if err := s.repo.Update(ctx, detail); err != nil {
		return nil, err
	}
	resp := dto.ToNodeDetailResponse(detail)
	return &resp, nil
}

// DeleteNodeDetail 删除节点详情
func (s *NodeDetailService) DeleteNodeDetail(ctx context.Context, detailID string) error {
	return s.repo.Delete(ctx, detailID)
}
