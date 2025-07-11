package service

import (
	"context"
	"time"

	"github.com/PGshen/thinking-map/server/internal/model"
	"github.com/PGshen/thinking-map/server/internal/model/dto"
	"github.com/PGshen/thinking-map/server/internal/pkg/comm"
	"github.com/PGshen/thinking-map/server/internal/repository"
	"github.com/gin-gonic/gin"

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
func (s *NodeService) ListNodes(ctx *gin.Context, mapID string) ([]dto.NodeResponse, error) {
	nodes, err := s.nodeRepo.FindByMapID(ctx, mapID)
	if err != nil {
		return nil, err
	}
	var res []dto.NodeResponse
	for _, n := range nodes {
		if len(n.Context.ParentProblem) == 0 && len(n.Context.SubProblem) == 0 {
			n.Context = s.GetNodeContext(ctx, n.ID)
		}
		resp := dto.ToNodeResponse(n)
		res = append(res, resp)
	}
	return res, nil
}

func (s *NodeService) GetNodeContext(ctx *gin.Context, nodeID string) model.NodeContext {
	// 获取节点上下文，parentProblem是所有祖先节点的问题和目标，subProblem是所有直接子节点的问题、目标和结论
	var parentProblems []model.Problem
	var subProblems []model.SubProblem

	// 获取所有祖先节点的问题和目标
	parentProblems = s.getAncestorProblems(ctx, nodeID)

	// 获取所有直接子节点的问题、目标和结论
	subProblems = s.getChildrenProblems(ctx, nodeID)

	return model.NodeContext{
		ParentProblem: parentProblems,
		SubProblem:    subProblems,
	}
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
	resp := dto.ToNodeResponse(node)
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

// getAncestorProblems 递归获取所有祖先节点的问题和目标
func (s *NodeService) getAncestorProblems(ctx *gin.Context, nodeID string) []model.Problem {
	var problems []model.Problem

	// 获取当前节点
	node, err := s.nodeRepo.FindByID(ctx, nodeID)
	if err != nil || node.ParentID == "" {
		return problems
	}

	// 获取父节点
	parentNode, err := s.nodeRepo.FindByID(ctx, node.ParentID)
	if err != nil {
		return problems
	}

	// 添加父节点的问题和目标
	problem := model.Problem{
		Question: parentNode.Question,
		Target:   parentNode.Target,
		Abstract: "", // 可以根据需要添加摘要逻辑
	}
	problems = append(problems, problem)

	// 递归获取更上层的祖先节点
	ancestorProblems := s.getAncestorProblems(ctx, parentNode.ID)
	problems = append(ancestorProblems, problem)

	return problems
}

// getChildrenProblems 获取所有直接子节点的问题、目标和结论
func (s *NodeService) getChildrenProblems(ctx *gin.Context, nodeID string) []model.SubProblem {
	var subProblems []model.SubProblem

	// 获取所有直接子节点
	childNodes, err := s.nodeRepo.FindByParentID(ctx, nodeID)
	if err != nil {
		return subProblems
	}

	for _, childNode := range childNodes {
		subProblem := model.SubProblem{
			Question:   childNode.Question,
			Target:     childNode.Target,
			Conclusion: childNode.Conclusion,
			Abstract:   "", // 可以根据需要添加摘要逻辑
		}
		subProblems = append(subProblems, subProblem)
	}

	return subProblems
}
