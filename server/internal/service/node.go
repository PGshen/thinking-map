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
	nodeRepo repository.ThinkingNode
	mapRepo  repository.ThinkingMap
}

func NewNodeService(nodeRepo repository.ThinkingNode, mapRepo repository.ThinkingMap) *NodeService {
	return &NodeService{
		nodeRepo: nodeRepo,
		mapRepo:  mapRepo,
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

func (s *NodeService) GetNode(ctx context.Context, nodeID string) (*dto.NodeResponse, error) {
	node, err := s.nodeRepo.FindByID(ctx, nodeID)
	if err != nil {
		return nil, err
	}
	resp := dto.ToNodeResponse(node)
	return &resp, nil
}

// ExecutableNodes 获取下一个可执行节点
func (s *NodeService) ExecutableNodes(ctx context.Context, mapID, nodeID string) (*dto.ExecutableNodesResponse, error) {
	// 获取该map下的所有节点
	nodes, err := s.nodeRepo.FindByMapID(ctx, mapID)
	if err != nil {
		return nil, err
	}

	// 构建节点ID到节点的映射，避免重复查询
	nodeMap := make(map[string]*model.ThinkingNode)
	for _, node := range nodes {
		nodeMap[node.ID] = node
	}

	// 构建父节点ID到子节点列表的映射
	childrenMap := make(map[string][]*model.ThinkingNode)
	for _, node := range nodes {
		if node.ParentID != "" && node.ParentID != uuid.Nil.String() {
			childrenMap[node.ParentID] = append(childrenMap[node.ParentID], node)
		}
	}

	// 1. 找到所有可执行节点
	var executableNodeIDs []string
	var nodesToUpdate []*model.ThinkingNode

	// 处理节点状态流转
	for _, node := range nodes {
		// 3.2 当节点可执行时，变更为pending
		if node.Status == comm.NodeStatusInitial {
			// 检查节点的依赖是否都已完成
			canExecute := true
			for _, depID := range node.Dependencies {
				// 从映射中查找依赖节点
				depNode, exists := nodeMap[depID]
				if !exists {
					// 依赖节点不存在，跳过
					continue
				}
				// 如果依赖节点未完成，则当前节点不可执行
				if depNode.Status != comm.NodeStatusCompleted {
					canExecute = false
					break
				}
			}

			if canExecute {
				// 将节点状态更新为pending
				node.Status = comm.NodeStatusPending
				nodesToUpdate = append(nodesToUpdate, node)
			}
		}

		// 3.4 当节点开始总结时，变更为in_conclusion
		if node.Status == comm.NodeStatusInDecomposition {
			// 检查所有子节点是否都已完成
			allChildrenCompleted := true
			childNodes := childrenMap[node.ID]
			if len(childNodes) > 0 {
				for _, childNode := range childNodes {
					if childNode.Status != comm.NodeStatusCompleted {
						allChildrenCompleted = false
						break
					}
				}

				if allChildrenCompleted {
					// 将节点状态更新为in_conclusion
					node.Status = comm.NodeStatusInConclusion
					nodesToUpdate = append(nodesToUpdate, node)
				}
			}
		}
	}

	// 收集所有可执行的节点
	for _, node := range nodes {
		if node.Status == comm.NodeStatusPending {
			executableNodeIDs = append(executableNodeIDs, node.ID)
		}
	}

	// 批量更新节点状态
	for _, node := range nodesToUpdate {
		if err := s.nodeRepo.Update(ctx, node); err != nil {
			return nil, err
		}
	}

	// 2. 根据深度优先遍历算法，找到最建议执行的下一个节点
	var suggestedNodeID string
	if nodeID != "" {
		currentNode, exists := nodeMap[nodeID]
		if exists {
			// 当前节点状态
			switch currentNode.Status {
			case comm.NodeStatusCompleted:
				// 如果当前节点已完成，优先检查兄弟节点
				if currentNode.ParentID != "" && currentNode.ParentID != uuid.Nil.String() {
					parentNode, parentExists := nodeMap[currentNode.ParentID]
					if parentExists {
						siblings := childrenMap[parentNode.ID]
						// 查找未完成的兄弟节点
						for _, sibling := range siblings {
							if sibling.ID != currentNode.ID && sibling.Status == comm.NodeStatusPending {
								suggestedNodeID = sibling.ID
								break
							}
						}

						// 如果没有未完成的兄弟节点，回到父节点
						if suggestedNodeID == "" && parentNode.Status == comm.NodeStatusPending {
							suggestedNodeID = parentNode.ID
						}
					}
				}

				// 如果没有找到建议节点，从所有可执行节点中选择一个
				if suggestedNodeID == "" && len(executableNodeIDs) > 0 {
					suggestedNodeID = executableNodeIDs[0]
				}

			case comm.NodeStatusInConclusion:
				// 当前节点正在总结中，不建议执行其他节点
				suggestedNodeID = nodeID

			default:
				// 检查当前节点的子节点
				childNodes := childrenMap[currentNode.ID]
				for _, childNode := range childNodes {
					if childNode.Status == comm.NodeStatusPending {
						suggestedNodeID = childNode.ID
						break
					}
				}

				// 如果没有可执行的子节点，则建议执行当前节点
				if suggestedNodeID == "" && currentNode.Status == comm.NodeStatusPending {
					suggestedNodeID = currentNode.ID
				} else if suggestedNodeID == "" && len(executableNodeIDs) > 0 {
					// 如果当前节点不可执行，从所有可执行节点中选择一个
					suggestedNodeID = executableNodeIDs[0]
				}
			}
		}
	} else if len(executableNodeIDs) > 0 {
		// 如果没有指定当前节点，从所有可执行节点中选择第一个
		suggestedNodeID = executableNodeIDs[0]
	}

	return &dto.ExecutableNodesResponse{
		NodeIDs:         executableNodeIDs,
		SuggestedNodeID: suggestedNodeID,
	}, nil
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
		Decomposition: model.Decomposition{
			IsDecomposed:   false,
			LastMessageID:  "",
			ConversationID: "",
		},
		Conclusion: model.Conclusion{
			LastMessageID: "",
			Content:       "",
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
			Conclusion: depNode.Conclusion.Content,
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
			Conclusion: childNode.Conclusion.Content,
			Abstract:   "", // 可以根据需要添加摘要逻辑
			Status:     childNode.Status,
		}
		nodeContexts = append(nodeContexts, nodeContext)
	}

	return nodeContexts
}
