package service

import (
	"context"
	"time"

	"github.com/PGshen/thinking-map/server/internal/model"
	"github.com/PGshen/thinking-map/server/internal/model/dto"
	"github.com/PGshen/thinking-map/server/internal/pkg/comm"
	"github.com/PGshen/thinking-map/server/internal/repository"

	"github.com/google/uuid"
)

type NodeService struct {
	nodeRepo       repository.ThinkingNode
	nodeDetailRepo repository.NodeDetail
	mapRepo        repository.ThinkingMap
}

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
		resp := dto.ToNodeResponse(n)
		if details != nil {
			resp.NodeDetails = make([]dto.NodeDetailResponse, len(details))
			for i, d := range details {
				resp.NodeDetails[i] = dto.ToNodeDetailResponse(d)
			}
		}
		res = append(res, resp)
	}
	return res, nil
}

// CreateNode 创建节点
func (s *NodeService) CreateNode(ctx context.Context, mapID string, req dto.CreateNodeRequest) (*dto.NodeResponse, error) {
	node := &model.ThinkingNode{
		ID:       uuid.NewString(),
		MapID:    mapID,
		ParentID: req.ParentID,
		NodeType: req.NodeType,
		Question: req.Question,
		Target:   req.Target,
		Status:   comm.NodeStatusPending,
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
	resp := dto.ToNodeResponse(node)
	resp.NodeDetails = make([]dto.NodeDetailResponse, 0)
	resp.NodeDetails = append(resp.NodeDetails, dto.ToNodeDetailResponse(infoDetail))
	resp.NodeDetails = append(resp.NodeDetails, dto.ToNodeDetailResponse(conclusionDetail))
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
	resp := dto.ToNodeResponse(node)
	return &resp, nil
}

// DeleteNode 删除节点
func (s *NodeService) DeleteNode(ctx context.Context, nodeID string) error {
	return s.nodeRepo.Delete(ctx, nodeID)
}

// AddDependency 添加节点依赖
func (s *NodeService) AddDependency(ctx context.Context, nodeID string, req dto.AddDependencyRequest) (*dto.DependencyInfo, error) {
	node, err := s.nodeRepo.FindByID(ctx, nodeID)
	if err != nil {
		return nil, err
	}
	// 检查是否已存在该依赖
	for _, dep := range node.Dependencies {
		if dep.NodeID == req.DependencyNodeID && dep.DependencyType == req.DependencyType {
			return nil, nil // 已存在，直接返回
		}
	}
	dep := model.Dependency{
		NodeID:         req.DependencyNodeID,
		DependencyType: req.DependencyType,
		Required:       req.Required,
	}
	node.Dependencies = append(node.Dependencies, dep)
	node.UpdatedAt = time.Now()
	if err := s.nodeRepo.Update(ctx, node); err != nil {
		return nil, err
	}
	return &dto.DependencyInfo{
		NodeID:         dep.NodeID,
		DependencyType: dep.DependencyType,
		Required:       dep.Required,
	}, nil
}

// DeleteDependency 删除节点依赖
func (s *NodeService) DeleteDependency(ctx context.Context, nodeID string, dependencyNodeID string) error {
	node, err := s.nodeRepo.FindByID(ctx, nodeID)
	if err != nil {
		return err
	}
	newDeps := make(model.Dependencies, 0, len(node.Dependencies))
	for _, dep := range node.Dependencies {
		if dep.NodeID != dependencyNodeID {
			newDeps = append(newDeps, dep)
		}
	}
	node.Dependencies = newDeps
	node.UpdatedAt = time.Now()
	return s.nodeRepo.Update(ctx, node)
}
