package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/thinking-map/server/internal/model"
	"github.com/thinking-map/server/internal/model/dto"
	"github.com/thinking-map/server/internal/pkg/comm"
	"github.com/thinking-map/server/internal/repository"
)

type NodeService struct {
	nodeRepo       repository.ThinkingNode
	nodeDetailRepo repository.NodeDetail
	mapRepo        repository.ThinkingMap
}

var (
	ErrNodeNotFound  = errors.New("node not found")
	ErrForbiddenNode = errors.New("forbidden: node does not belong to user")
)

func NewNodeService(nodeRepo repository.ThinkingNode, nodeDetailRepo repository.NodeDetail, mapRepo repository.ThinkingMap) *NodeService {
	return &NodeService{
		nodeRepo:       nodeRepo,
		nodeDetailRepo: nodeDetailRepo,
		mapRepo:        mapRepo,
	}
}

// ListNodes 获取某个map下的所有节点
func (s *NodeService) ListNodes(ctx context.Context, mapID string) ([]dto.NodeResponse, error) {
	nodes, err := s.nodeRepo.FindByMapID(ctx, mapID)
	if err != nil {
		return nil, err
	}
	var res []dto.NodeResponse
	for _, n := range nodes {
		details, err := s.nodeDetailRepo.FindByNodeID(ctx, n.ID)
		if err != nil {
			return nil, err
		}
		resp := modelToNodeResponse(n)
		if details != nil {
			resp.NodeDetails = make([]model.NodeDetail, len(details))
			for i, d := range details {
				resp.NodeDetails[i] = *d
			}
		}
		res = append(res, resp)
	}
	return res, nil
}

// CreateNode 创建节点
func (s *NodeService) CreateNode(ctx context.Context, mapID string, req dto.CreateNodeRequest, userID string) (*dto.NodeResponse, error) {
	node := &model.ThinkingNode{
		MapID:    mapID,
		ParentID: req.ParentID,
		NodeType: req.NodeType,
		Question: req.Question,
		Target:   req.Target,
		Status:   1,
		Position: model.Position{
			X:      req.Position.X,
			Y:      req.Position.Y,
			Width:  req.Position.Width,
			Height: req.Position.Height,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := s.nodeRepo.Create(ctx, node); err != nil {
		return nil, err
	}
	// 创建node同时创建node_detail记录，默认创建info, conclusion两种类型的节点详情，decompose类型在执行过程中有需要再创建
	infoDetail := &model.NodeDetail{
		ID:         uuid.NewString(),
		NodeID:     node.ID,
		DetailType: comm.DetailTypeInfo,
		Content: model.DetailContent{
			Question: req.Question,
			Target:   req.Target,
		},
		Status:    comm.DetailStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	conclusionDetail := &model.NodeDetail{
		ID:         uuid.NewString(),
		NodeID:     node.ID,
		DetailType: comm.DetailTypeConclusion,
		Content:    model.DetailContent{},
		Status:     comm.DetailStatusPending,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if err := s.nodeDetailRepo.Create(ctx, infoDetail); err != nil {
		return nil, err
	}
	if err := s.nodeDetailRepo.Create(ctx, conclusionDetail); err != nil {
		return nil, err
	}
	resp := modelToNodeResponse(node)
	return &resp, nil
}

// UpdateNode 更新节点
func (s *NodeService) UpdateNode(ctx context.Context, nodeID string, req dto.UpdateNodeRequest) (*dto.NodeResponse, error) {
	node, err := s.nodeRepo.FindByID(ctx, nodeID)
	if err != nil {
		return nil, err
	}
	if req.Question != "" {
		node.Question = req.Question
	}
	if req.Target != "" {
		node.Target = req.Target
	}
	if (req.Position != model.Position{}) {
		node.Position = model.Position{
			X:      req.Position.X,
			Y:      req.Position.Y,
			Width:  req.Position.Width,
			Height: req.Position.Height,
		}
	}
	node.UpdatedAt = time.Now()
	if err := s.nodeRepo.Update(ctx, node); err != nil {
		return nil, err
	}
	resp := modelToNodeResponse(node)
	return &resp, nil
}

// DeleteNode 删除节点
func (s *NodeService) DeleteNode(ctx context.Context, nodeID string) error {
	return s.nodeRepo.Delete(ctx, nodeID)
}

// modelToNodeResponse 将model.ThinkingNode转为dto.NodeResponse
func modelToNodeResponse(n *model.ThinkingNode) dto.NodeResponse {
	return dto.NodeResponse{
		ID:           n.ID,
		MapID:        n.MapID,
		ParentID:     n.ParentID,
		NodeType:     n.NodeType,
		Question:     n.Question,
		Target:       n.Target,
		Context:      n.Context,
		Status:       n.Status,
		Position:     n.Position,
		Dependencies: n.Dependencies,
		CreatedAt:    n.CreatedAt,
		UpdatedAt:    n.UpdatedAt,
	}
}
