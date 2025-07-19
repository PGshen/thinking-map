/*
 * @Date: 2025-01-27
 * @LastEditors: AI Assistant
 * @FilePath: /thinking-map/server/internal/service/dependency_checker.go
 */
package service

import (
	"context"
	"fmt"
	"sort"

	"github.com/PGshen/thinking-map/server/internal/model"
	"github.com/PGshen/thinking-map/server/internal/pkg/comm"
	"github.com/PGshen/thinking-map/server/internal/repository"
)

// DependencyChecker 依赖检查器
// 职责：通过工程化方式检查节点的依赖关系，确保执行顺序的正确性
type DependencyChecker struct {
	nodeRepo repository.ThinkingNode
}

// NewDependencyChecker 创建新的依赖检查器实例
func NewDependencyChecker(nodeRepo repository.ThinkingNode) *DependencyChecker {
	return &DependencyChecker{
		nodeRepo: nodeRepo,
	}
}

// ExecutionReadinessResult 执行就绪检查结果
type ExecutionReadinessResult struct {
	Ready   bool   `json:"ready"`
	Message string `json:"message"`
}

// CheckExecutionReadiness 检查节点是否可以执行
func (dc *DependencyChecker) CheckExecutionReadiness(ctx context.Context, nodeID string) (*ExecutionReadinessResult, error) {
	// 获取当前节点
	node, err := dc.nodeRepo.FindByID(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to find node %s: %w", nodeID, err)
	}

	// 检查直接子节点状态
	childNodes, err := dc.nodeRepo.FindByParentID(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to find child nodes: %w", err)
	}

	// 如果有直接子节点，检查它们是否都已完成
	for _, childNode := range childNodes {
		if childNode.Status != comm.NodeStatusCompleted {
			return &ExecutionReadinessResult{
				Ready:   false,
				Message: fmt.Sprintf("直接子节点 %s (%s) 尚未完成，当前状态: %s", childNode.Question, childNode.ID, childNode.Status),
			}, nil
		}
	}

	// 如果没有依赖，且子节点都已完成，可以执行
	if len(node.Dependencies) == 0 {
		return &ExecutionReadinessResult{
			Ready:   true,
			Message: "无依赖节点且子节点已完成，可以执行",
		}, nil
	}

	// 检查所有依赖节点的状态
	dependencyNodes, err := dc.nodeRepo.FindByIDs(ctx, node.Dependencies)
	if err != nil {
		return nil, fmt.Errorf("failed to find dependency nodes: %w", err)
	}

	// 检查每个依赖节点是否已完成
	for _, depNode := range dependencyNodes {
		if depNode.Status != comm.NodeStatusCompleted {
			return &ExecutionReadinessResult{
				Ready:   false,
				Message: fmt.Sprintf("依赖节点 %s (%s) 尚未完成，当前状态: %s", depNode.Question, depNode.ID, depNode.Status),
			}, nil
		}
	}

	return &ExecutionReadinessResult{
		Ready:   true,
		Message: "所有依赖和子节点已满足，可以执行",
	}, nil
}

// GetExecutionOrder 获取节点的执行顺序（拓扑排序）
func (dc *DependencyChecker) GetExecutionOrder(ctx context.Context, nodeIDs []string) ([]string, error) {
	if len(nodeIDs) == 0 {
		return []string{}, nil
	}

	// 获取所有节点
	nodes, err := dc.nodeRepo.FindByIDs(ctx, nodeIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to find nodes: %w", err)
	}

	// 构建节点映射
	nodeMap := make(map[string]*model.ThinkingNode)
	for _, node := range nodes {
		nodeMap[node.ID] = node
	}

	// 执行拓扑排序
	return dc.topologicalSort(nodeMap, nodeIDs)
}

// ValidateDependencyCycle 检查是否会产生循环依赖
func (dc *DependencyChecker) ValidateDependencyCycle(ctx context.Context, fromNodeID, toNodeID string) (bool, error) {
	// 检查从toNode到fromNode是否存在路径
	hasPath, err := dc.hasPath(ctx, toNodeID, fromNodeID)
	if err != nil {
		return false, fmt.Errorf("failed to check path: %w", err)
	}

	// 如果存在路径，则会产生循环依赖
	return !hasPath, nil
}

// GetDependencyChain 获取节点的完整依赖链
func (dc *DependencyChecker) GetDependencyChain(ctx context.Context, nodeID string) ([]string, error) {
	visited := make(map[string]bool)
	var chain []string

	err := dc.buildDependencyChain(ctx, nodeID, visited, &chain)
	if err != nil {
		return nil, err
	}

	return chain, nil
}

// GetBlockedNodes 获取被指定节点阻塞的所有节点
func (dc *DependencyChecker) GetBlockedNodes(ctx context.Context, nodeID string) ([]string, error) {
	// 获取所有节点（在同一个map中）
	node, err := dc.nodeRepo.FindByID(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to find node: %w", err)
	}

	allNodes, err := dc.nodeRepo.FindByMapID(ctx, node.MapID)
	if err != nil {
		return nil, fmt.Errorf("failed to find all nodes in map: %w", err)
	}

	var blockedNodes []string
	for _, n := range allNodes {
		// 检查该节点是否依赖于指定节点
		for _, depID := range n.Dependencies {
			if depID == nodeID {
				blockedNodes = append(blockedNodes, n.ID)
				break
			}
		}
	}

	return blockedNodes, nil
}

// topologicalSort 拓扑排序实现
func (dc *DependencyChecker) topologicalSort(nodeMap map[string]*model.ThinkingNode, nodeIDs []string) ([]string, error) {
	// 计算入度
	inDegree := make(map[string]int)
	adjList := make(map[string][]string)

	// 初始化入度和邻接表
	for _, nodeID := range nodeIDs {
		inDegree[nodeID] = 0
		adjList[nodeID] = []string{}
	}

	// 构建图和计算入度
	for _, nodeID := range nodeIDs {
		node := nodeMap[nodeID]
		if node == nil {
			continue
		}

		for _, depID := range node.Dependencies {
			// 只考虑在给定节点列表中的依赖
			if _, exists := inDegree[depID]; exists {
				adjList[depID] = append(adjList[depID], nodeID)
				inDegree[nodeID]++
			}
		}
	}

	// 找到所有入度为0的节点
	var queue []string
	for nodeID, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, nodeID)
		}
	}

	// 对队列进行排序以确保结果的确定性
	sort.Strings(queue)

	var result []string

	// 拓扑排序
	for len(queue) > 0 {
		// 取出队列中的第一个节点
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		// 更新相邻节点的入度
		var newZeroDegreeNodes []string
		for _, neighbor := range adjList[current] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				newZeroDegreeNodes = append(newZeroDegreeNodes, neighbor)
			}
		}

		// 对新的零入度节点排序后加入队列
		sort.Strings(newZeroDegreeNodes)
		queue = append(queue, newZeroDegreeNodes...)
	}

	// 检查是否存在循环依赖
	if len(result) != len(nodeIDs) {
		return nil, fmt.Errorf("检测到循环依赖，无法完成拓扑排序")
	}

	return result, nil
}

// hasPath 检查从源节点到目标节点是否存在路径（DFS）
func (dc *DependencyChecker) hasPath(ctx context.Context, sourceID, targetID string) (bool, error) {
	if sourceID == targetID {
		return true, nil
	}

	visited := make(map[string]bool)
	return dc.dfsPath(ctx, sourceID, targetID, visited)
}

// dfsPath 深度优先搜索路径
func (dc *DependencyChecker) dfsPath(ctx context.Context, currentID, targetID string, visited map[string]bool) (bool, error) {
	if currentID == targetID {
		return true, nil
	}

	if visited[currentID] {
		return false, nil
	}

	visited[currentID] = true

	// 获取当前节点
	node, err := dc.nodeRepo.FindByID(ctx, currentID)
	if err != nil {
		return false, err
	}

	// 遍历所有依赖节点
	for _, depID := range node.Dependencies {
		found, err := dc.dfsPath(ctx, depID, targetID, visited)
		if err != nil {
			return false, err
		}
		if found {
			return true, nil
		}
	}

	return false, nil
}

// buildDependencyChain 构建依赖链
func (dc *DependencyChecker) buildDependencyChain(ctx context.Context, nodeID string, visited map[string]bool, chain *[]string) error {
	if visited[nodeID] {
		return fmt.Errorf("检测到循环依赖，节点ID: %s", nodeID)
	}

	visited[nodeID] = true
	*chain = append(*chain, nodeID)

	// 获取当前节点
	node, err := dc.nodeRepo.FindByID(ctx, nodeID)
	if err != nil {
		return err
	}

	// 递归处理依赖节点
	for _, depID := range node.Dependencies {
		err := dc.buildDependencyChain(ctx, depID, visited, chain)
		if err != nil {
			return err
		}
	}

	visited[nodeID] = false // 回溯
	return nil
}