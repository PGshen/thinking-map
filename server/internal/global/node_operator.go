/*
 * @Date: 2025-01-27 00:00:00
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-01-27 00:00:00
 * @FilePath: /thinking-map/server/internal/global/node_operator.go
 */
package global

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/PGshen/thinking-map/server/internal/model"
	"github.com/PGshen/thinking-map/server/internal/model/dto"
	"github.com/PGshen/thinking-map/server/internal/repository"
	"github.com/google/uuid"
)

var (
	// GlobalNodeOperator 全局节点操作器实例
	GlobalNodeOperator *NodeOperator
	nodeOperatorOnce   sync.Once
)

// NodeOperator 节点操作器
type NodeOperator struct {
	nodeRepo repository.ThinkingNode
	mapRepo  repository.ThinkingMap
}

// InitNodeOperator 初始化全局节点操作器
func InitNodeOperator(nodeRepo repository.ThinkingNode, mapRepo repository.ThinkingMap) {
	nodeOperatorOnce.Do(func() {
		GlobalNodeOperator = &NodeOperator{
			nodeRepo: nodeRepo,
			mapRepo:  mapRepo,
		}
	})
}

// GetNodeOperator 获取全局节点操作器实例
func GetNodeOperator() *NodeOperator {
	if GlobalNodeOperator == nil {
		panic("node operator not initialized, call InitNodeOperator first")
	}
	return GlobalNodeOperator
}

// CreateNode 创建节点
func (s *NodeOperator) CreateNode(ctx context.Context, req dto.CreateNodeRequest) (*dto.NodeResponse, error) {
	// 验证父节点是否存在（如果不是根节点）
	if req.ParentID != "" && req.ParentID != uuid.Nil.String() {
		if _, err := s.nodeRepo.FindByID(ctx, req.ParentID); err != nil {
			return nil, fmt.Errorf("parent node not found: %w", err)
		}
	}

	// 验证地图是否存在
	if _, err := s.mapRepo.FindByID(ctx, req.MapID); err != nil {
		return nil, fmt.Errorf("map not found: %w", err)
	}

	node := &model.ThinkingNode{
		ID:       uuid.NewString(),
		MapID:    req.MapID,
		ParentID: req.ParentID,
		NodeType: req.NodeType,
		Question: req.Question,
		Target:   req.Target,
		Status:   "initial",
		Position: req.Position,
		Context:  model.DependentContext{},
		Decomposition: model.Decomposition{
			ConversationID: "",
			LastMessageID:  "",
		},
		Conclusion: model.Conclusion{
			ConversationID: "",
			LastMessageID:  "",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.nodeRepo.Create(ctx, node); err != nil {
		return nil, fmt.Errorf("failed to create node: %w", err)
	}

	resp := dto.ToNodeResponse(node)
	return &resp, nil
}

// UpdateNode 更新节点
func (s *NodeOperator) UpdateNode(ctx context.Context, nodeID string, req dto.UpdateNodeRequest) (*dto.NodeResponse, error) {
	node, err := s.nodeRepo.FindByID(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("node not found: %w", err)
	}

	// 更新字段
	if req.Question != "" {
		node.Question = req.Question
	}
	if req.Target != "" {
		node.Target = req.Target
	}
	if req.Position.X != 0 || req.Position.Y != 0 {
		node.Position = req.Position
	}
	node.UpdatedAt = time.Now()

	if err = s.nodeRepo.Update(ctx, node); err != nil {
		return nil, fmt.Errorf("failed to update node: %w", err)
	}

	// 重新从数据库获取更新后的记录
	updatedNode, err := s.nodeRepo.FindByID(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated node: %w", err)
	}

	resp := dto.ToNodeResponse(updatedNode)
	return &resp, nil
}

// DeleteNode 删除节点
func (s *NodeOperator) DeleteNode(ctx context.Context, nodeID string) error {
	// 检查节点是否存在
	node, err := s.nodeRepo.FindByID(ctx, nodeID)
	if err != nil {
		return fmt.Errorf("node not found: %w", err)
	}

	// 检查是否有子节点
	children, err := s.nodeRepo.FindByParentID(ctx, nodeID)
	if err != nil {
		return fmt.Errorf("failed to check children: %w", err)
	}
	if len(children) > 0 {
		return errors.New("cannot delete node with children")
	}

	// 检查是否为根节点
	if node.NodeType == "root" {
		return errors.New("cannot delete root node")
	}

	return s.nodeRepo.Delete(ctx, nodeID)
}

// UpdateNodeDependencies 更新节点依赖关系
func (s *NodeOperator) UpdateNodeDependencies(ctx context.Context, nodeID string, dependencies []string) (*model.ThinkingNode, error) {
	// 获取节点
	node, err := s.nodeRepo.FindByID(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("node not found: %w", err)
	}

	// 更新依赖关系
	node.Dependencies = dependencies
	node.UpdatedAt = time.Now()

	// 保存到数据库
	if err := s.nodeRepo.Update(ctx, node); err != nil {
		return nil, fmt.Errorf("failed to update node dependencies: %w", err)
	}

	return node, nil
}

// GetNodesByIDs 获取多个节点
func (s *NodeOperator) GetNodesByIDs(ctx context.Context, nodeIDs []string) ([]*model.ThinkingNode, error) {
	nodes, err := s.nodeRepo.FindByIDs(ctx, nodeIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes: %w", err)
	}
	return nodes, nil
}
