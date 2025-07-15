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
		n.Context = s.GetNodeContext(ctx, n)
		resp := dto.ToNodeResponse(n)
		res = append(res, resp)
	}
	return res, nil
}

func (s *NodeService) GetNodeContext(ctx *gin.Context, node *model.ThinkingNode) model.DependentContext {
	// 获取节点上下文，parentProblem是所有祖先节点的问题和目标，subProblem是所有直接子节点的问题、目标和结论
	ancestor := node.Context.Ancestor
	prevSibling := node.Context.PrevSibling
	children := node.Context.Children

	// 获取所有祖先节点的问题和目标
	if len(ancestor) == 0 {
		ancestor = s.getAncestor(ctx, node.ID)
	}
	// 获取所有前一个兄弟节点的问题、目标和结论
	if len(prevSibling) == 0 {
		prevSibling = s.getPreSibling(ctx, node)
	}
	// 获取所有直接子节点的问题、目标和结论
	if len(children) == 0 {
		children = s.getChildren(ctx, node.ID)
	}

	return model.DependentContext{
		Ancestor:    ancestor,
		PrevSibling: prevSibling,
		Children:    children,
	}
}

// UpdateNodeContext 更新节点上下文
func (s *NodeService) UpdateNodeContext(ctx *gin.Context, nodeID string, req dto.UpdateNodeContextRequest) (*dto.NodeResponse, error) {
	node, err := s.nodeRepo.FindByID(ctx, nodeID)
	if err != nil {
		return nil, err
	}
	if len(req.Context.Ancestor) > 0 {
		node.Context.Ancestor = req.Context.Ancestor
	}
	if len(req.Context.PrevSibling) > 0 {
		node.Context.PrevSibling = req.Context.PrevSibling
	}
	if len(req.Context.Children) > 0 {
		node.Context.Children = req.Context.Children
	}
	if err := s.nodeRepo.Update(ctx, node); err != nil {
		return nil, err
	}
	resp := dto.ToNodeResponse(node)
	return &resp, nil
}

// ResetNodeContext 重置上下文
func (s *NodeService) ResetNodeContext(ctx *gin.Context, nodeID string) (*dto.NodeResponse, error) {
	node, err := s.nodeRepo.FindByID(ctx, nodeID)
	if err != nil {
		return nil, err
	}

	// Reset ancestor context
	node.Context.Ancestor = s.getAncestor(ctx, nodeID)

	// Reset previous sibling context
	node.Context.PrevSibling = s.getPreSibling(ctx, node)

	// Reset children context
	node.Context.Children = s.getChildren(ctx, nodeID)

	// Update node with new context
	if err := s.nodeRepo.Update(ctx, node); err != nil {
		return nil, err
	}

	resp := dto.ToNodeResponse(node)
	return &resp, nil
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
			X: req.Position.X,
			Y: req.Position.Y,
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
			X: req.Position.X,
			Y: req.Position.Y,
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
func (s *NodeService) getAncestor(ctx *gin.Context, nodeID string) []model.NodeContext {
	var nodeContexts []model.NodeContext

	// 获取当前节点
	node, err := s.nodeRepo.FindByID(ctx, nodeID)
	if err != nil || node.ParentID == "" {
		return nodeContexts
	}

	// 获取父节点
	parentNode, err := s.nodeRepo.FindByID(ctx, node.ParentID)
	if err != nil {
		return nodeContexts
	}

	// 添加父节点的问题和目标
	nodeContext := model.NodeContext{
		Question: parentNode.Question,
		Target:   parentNode.Target,
		Abstract: "", // 可以根据需要添加摘要逻辑
		Status:   parentNode.Status,
	}
	// 递归获取更上层的祖先节点
	ancestor := s.getAncestor(ctx, parentNode.ID)
	ancestor = append(ancestor, nodeContext)

	return ancestor
}

// getPreSibling 获取所有前一个兄弟节点的问题、目标和结论
func (s *NodeService) getPreSibling(ctx *gin.Context, node *model.ThinkingNode) []model.NodeContext {
	var nodeContexts []model.NodeContext

	// node.Dependencies 是当前节点依赖的节点id，通过这个查询依赖节点的问题、目标和结论
	if len(node.Dependencies) == 0 {
		return nodeContexts
	}

	// 查找所有依赖节点
	depNodes, err := s.nodeRepo.FindByIDs(ctx, node.Dependencies)
	if err != nil {
		return nodeContexts
	}

	for _, depNode := range depNodes {
		nodeContext := model.NodeContext{
			Question:   depNode.Question,
			Target:     depNode.Target,
			Conclusion: depNode.Conclusion,
			Abstract:   "", // 可以根据需要添加摘要逻辑
			Status:     depNode.Status,
		}
		nodeContexts = append(nodeContexts, nodeContext)
	}

	return nodeContexts
}

// getChildren 获取所有直接子节点的问题、目标和结论
func (s *NodeService) getChildren(ctx *gin.Context, nodeID string) []model.NodeContext {
	var nodeContexts []model.NodeContext

	// 获取所有直接子节点
	childNodes, err := s.nodeRepo.FindByParentID(ctx, nodeID)
	if err != nil {
		return nodeContexts
	}

	for _, childNode := range childNodes {
		nodeContext := model.NodeContext{
			Question:   childNode.Question,
			Target:     childNode.Target,
			Conclusion: childNode.Conclusion,
			Abstract:   "", // 可以根据需要添加摘要逻辑
			Status:     childNode.Status,
		}
		nodeContexts = append(nodeContexts, nodeContext)
	}

	return nodeContexts
}
