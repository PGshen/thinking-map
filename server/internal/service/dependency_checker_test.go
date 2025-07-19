/*
 * @Date: 2025-01-27
 * @LastEditors: AI Assistant
 * @FilePath: /thinking-map/server/internal/service/dependency_checker_test.go
 */
package service

import (
	"context"
	"testing"

	"github.com/PGshen/thinking-map/server/internal/model"
	"github.com/PGshen/thinking-map/server/internal/model/dto"
	"github.com/PGshen/thinking-map/server/internal/pkg/comm"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestDependencyChecker_CheckExecutionReadiness(t *testing.T) {
	ctx := context.Background()
	userID := uuid.NewString()

	// 1. 创建测试数据 - 创建一个map
	createMapReq := dto.CreateMapRequest{
		Title:       "依赖检查测试",
		Problem:     "测试依赖检查功能",
		ProblemType: "测试类型",
		Target:      "测试目标",
		KeyPoints:   []string{"要点1"},
		Constraints: []string{"约束1"},
	}
	mapResp, err := mapSvc.CreateMap(ctx, createMapReq, userID)
	assert.NoError(t, err)
	mapID := mapResp.ID

	// 2. 创建依赖节点A
	createNodeAReq := dto.CreateNodeRequest{
		MapID:    mapID,
		ParentID: uuid.Nil.String(),
		NodeType: "analysis",
		Question: "节点A问题",
		Target:   "节点A目标",
		Position: model.Position{X: 0, Y: 0},
	}
	nodeAResp, err := nodeSvc.CreateNode(ctx, mapID, createNodeAReq)
	assert.NoError(t, err)
	nodeAID := nodeAResp.ID

	// 3. 创建依赖节点B
	createNodeBReq := dto.CreateNodeRequest{
		MapID:    mapID,
		ParentID: uuid.Nil.String(),
		NodeType: "analysis",
		Question: "节点B问题",
		Target:   "节点B目标",
		Position: model.Position{X: 10, Y: 10},
	}
	nodeBResp, err := nodeSvc.CreateNode(ctx, mapID, createNodeBReq)
	assert.NoError(t, err)
	nodeBID := nodeBResp.ID

	// 4. 创建目标节点C，依赖于A和B
	createNodeCReq := dto.CreateNodeRequest{
		MapID:    mapID,
		ParentID: uuid.Nil.String(),
		NodeType: "conclusion",
		Question: "节点C问题",
		Target:   "节点C目标",
		Position: model.Position{X: 20, Y: 20},
	}
	nodeCResp, err := nodeSvc.CreateNode(ctx, mapID, createNodeCReq)
	assert.NoError(t, err)
	nodeCID := nodeCResp.ID

	// 5. 设置节点C的依赖关系
	cm := NewContextManager(nodeSvc.nodeRepo, mapSvc.mapRepo, nil)
	err = cm.UpdateNodeDependencies(ctx, nodeCID, []string{nodeAID, nodeBID})
	assert.NoError(t, err)

	// 6. 创建依赖检查器
	dc := NewDependencyChecker(nodeSvc.nodeRepo)

	// 7. 测试节点C的执行就绪性（依赖节点未完成）
	result, err := dc.CheckExecutionReadiness(ctx, nodeCID)
	assert.NoError(t, err)
	assert.False(t, result.Ready)
	assert.Contains(t, result.Message, "尚未完成")

	// 8. 完成节点A
	nodeA, err := nodeSvc.nodeRepo.FindByID(ctx, nodeAID)
	assert.NoError(t, err)
	nodeA.Status = comm.NodeStatusCompleted
	err = nodeSvc.nodeRepo.Update(ctx, nodeA)
	assert.NoError(t, err)

	// 9. 再次检查节点C（还有节点B未完成）
	result, err = dc.CheckExecutionReadiness(ctx, nodeCID)
	assert.NoError(t, err)
	assert.False(t, result.Ready)
	assert.Contains(t, result.Message, "节点B问题")

	// 10. 完成节点B
	nodeB, err := nodeSvc.nodeRepo.FindByID(ctx, nodeBID)
	assert.NoError(t, err)
	nodeB.Status = comm.NodeStatusCompleted
	err = nodeSvc.nodeRepo.Update(ctx, nodeB)
	assert.NoError(t, err)

	// 11. 再次检查节点C（所有依赖已完成）
	result, err = dc.CheckExecutionReadiness(ctx, nodeCID)
	assert.NoError(t, err)
	assert.True(t, result.Ready)
	assert.Contains(t, result.Message, "所有依赖和子节点已满足")

	// 12. 测试无依赖节点
	result, err = dc.CheckExecutionReadiness(ctx, nodeAID)
	assert.NoError(t, err)
	assert.True(t, result.Ready)
	assert.Contains(t, result.Message, "无依赖节点且子节点已完成")
}

func TestDependencyChecker_CheckExecutionReadinessWithChildren(t *testing.T) {
	ctx := context.Background()
	userID := uuid.NewString()

	// 1. 创建测试数据 - 创建一个map
	createMapReq := dto.CreateMapRequest{
		Title:       "子节点状态检查测试",
		Problem:     "测试子节点状态检查功能",
		ProblemType: "测试类型",
		Target:      "测试目标",
		KeyPoints:   []string{"要点1"},
		Constraints: []string{"约束1"},
	}
	mapResp, err := mapSvc.CreateMap(ctx, createMapReq, userID)
	assert.NoError(t, err)
	mapID := mapResp.ID

	// 2. 创建父节点
	parentNodeResp, err := nodeSvc.CreateNode(ctx, mapID, dto.CreateNodeRequest{
		MapID:    mapID,
		ParentID: uuid.Nil.String(),
		NodeType: "analysis",
		Question: "父节点",
		Target:   "父节点目标",
		Position: model.Position{X: 0, Y: 0},
	})
	assert.NoError(t, err)
	parentNodeID := parentNodeResp.ID

	// 3. 创建子节点1
	childNode1Resp, err := nodeSvc.CreateNode(ctx, mapID, dto.CreateNodeRequest{
		MapID:    mapID,
		ParentID: parentNodeID,
		NodeType: "analysis",
		Question: "子节点1",
		Target:   "子节点1目标",
		Position: model.Position{X: 10, Y: 10},
	})
	assert.NoError(t, err)
	childNode1ID := childNode1Resp.ID

	// 4. 创建子节点2
	childNode2Resp, err := nodeSvc.CreateNode(ctx, mapID, dto.CreateNodeRequest{
		MapID:    mapID,
		ParentID: parentNodeID,
		NodeType: "analysis",
		Question: "子节点2",
		Target:   "子节点2目标",
		Position: model.Position{X: 20, Y: 20},
	})
	assert.NoError(t, err)
	childNode2ID := childNode2Resp.ID

	// 5. 创建依赖检查器
	dc := NewDependencyChecker(nodeSvc.nodeRepo)

	// 6. 测试父节点的执行就绪性（子节点未完成）
	result, err := dc.CheckExecutionReadiness(ctx, parentNodeID)
	assert.NoError(t, err)
	assert.False(t, result.Ready)
	assert.Contains(t, result.Message, "直接子节点")
	assert.Contains(t, result.Message, "尚未完成")

	// 7. 完成子节点1
	childNode1, err := nodeSvc.nodeRepo.FindByID(ctx, childNode1ID)
	assert.NoError(t, err)
	childNode1.Status = comm.NodeStatusCompleted
	err = nodeSvc.nodeRepo.Update(ctx, childNode1)
	assert.NoError(t, err)

	// 8. 再次检查父节点（还有子节点2未完成）
	result, err = dc.CheckExecutionReadiness(ctx, parentNodeID)
	assert.NoError(t, err)
	assert.False(t, result.Ready)
	assert.Contains(t, result.Message, "子节点2")
	assert.Contains(t, result.Message, "尚未完成")

	// 9. 完成子节点2
	childNode2, err := nodeSvc.nodeRepo.FindByID(ctx, childNode2ID)
	assert.NoError(t, err)
	childNode2.Status = comm.NodeStatusCompleted
	err = nodeSvc.nodeRepo.Update(ctx, childNode2)
	assert.NoError(t, err)

	// 10. 再次检查父节点（所有子节点已完成）
	result, err = dc.CheckExecutionReadiness(ctx, parentNodeID)
	assert.NoError(t, err)
	assert.True(t, result.Ready)
	assert.Contains(t, result.Message, "无依赖节点且子节点已完成")
}

func TestDependencyChecker_GetExecutionOrder(t *testing.T) {
	ctx := context.Background()
	userID := uuid.NewString()

	// 1. 创建测试数据
	createMapReq := dto.CreateMapRequest{
		Title:       "执行顺序测试",
		Problem:     "测试执行顺序功能",
		ProblemType: "测试类型",
		Target:      "测试目标",
		KeyPoints:   []string{"要点1"},
		Constraints: []string{"约束1"},
	}
	mapResp, err := mapSvc.CreateMap(ctx, createMapReq, userID)
	assert.NoError(t, err)
	mapID := mapResp.ID

	// 2. 创建节点A（无依赖）
	nodeAResp, err := nodeSvc.CreateNode(ctx, mapID, dto.CreateNodeRequest{
		MapID:    mapID,
		ParentID: uuid.Nil.String(),
		NodeType: "analysis",
		Question: "节点A",
		Target:   "目标A",
		Position: model.Position{X: 0, Y: 0},
	})
	assert.NoError(t, err)
	nodeAID := nodeAResp.ID

	// 3. 创建节点B（无依赖）
	nodeBResp, err := nodeSvc.CreateNode(ctx, mapID, dto.CreateNodeRequest{
		MapID:    mapID,
		ParentID: uuid.Nil.String(),
		NodeType: "analysis",
		Question: "节点B",
		Target:   "目标B",
		Position: model.Position{X: 10, Y: 10},
	})
	assert.NoError(t, err)
	nodeBID := nodeBResp.ID

	// 4. 创建节点C（依赖A）
	nodeCResp, err := nodeSvc.CreateNode(ctx, mapID, dto.CreateNodeRequest{
		MapID:    mapID,
		ParentID: uuid.Nil.String(),
		NodeType: "analysis",
		Question: "节点C",
		Target:   "目标C",
		Position: model.Position{X: 20, Y: 20},
	})
	assert.NoError(t, err)
	nodeCID := nodeCResp.ID

	// 5. 创建节点D（依赖B和C）
	nodeDResp, err := nodeSvc.CreateNode(ctx, mapID, dto.CreateNodeRequest{
		MapID:    mapID,
		ParentID: uuid.Nil.String(),
		NodeType: "conclusion",
		Question: "节点D",
		Target:   "目标D",
		Position: model.Position{X: 30, Y: 30},
	})
	assert.NoError(t, err)
	nodeDID := nodeDResp.ID

	// 6. 设置依赖关系
	cm := NewContextManager(nodeSvc.nodeRepo, mapSvc.mapRepo, nil)
	err = cm.UpdateNodeDependencies(ctx, nodeCID, []string{nodeAID})
	assert.NoError(t, err)
	err = cm.UpdateNodeDependencies(ctx, nodeDID, []string{nodeBID, nodeCID})
	assert.NoError(t, err)

	// 7. 测试执行顺序
	dc := NewDependencyChecker(nodeSvc.nodeRepo)
	nodeIDs := []string{nodeAID, nodeBID, nodeCID, nodeDID}
	executionOrder, err := dc.GetExecutionOrder(ctx, nodeIDs)
	assert.NoError(t, err)
	assert.Len(t, executionOrder, 4)

	// 8. 验证执行顺序的正确性
	// A和B应该在C之前，C应该在D之前
	posA := findPosition(executionOrder, nodeAID)
	posB := findPosition(executionOrder, nodeBID)
	posC := findPosition(executionOrder, nodeCID)
	posD := findPosition(executionOrder, nodeDID)

	assert.True(t, posA < posC, "A应该在C之前")
	assert.True(t, posB < posD, "B应该在D之前")
	assert.True(t, posC < posD, "C应该在D之前")
}

func TestDependencyChecker_ValidateDependencyCycle(t *testing.T) {
	ctx := context.Background()
	userID := uuid.NewString()

	// 1. 创建测试数据
	createMapReq := dto.CreateMapRequest{
		Title:       "循环依赖测试",
		Problem:     "测试循环依赖检查",
		ProblemType: "测试类型",
		Target:      "测试目标",
		KeyPoints:   []string{"要点1"},
		Constraints: []string{"约束1"},
	}
	mapResp, err := mapSvc.CreateMap(ctx, createMapReq, userID)
	assert.NoError(t, err)
	mapID := mapResp.ID

	// 2. 创建节点A和B
	nodeAResp, err := nodeSvc.CreateNode(ctx, mapID, dto.CreateNodeRequest{
		MapID:    mapID,
		ParentID: uuid.Nil.String(),
		NodeType: "analysis",
		Question: "节点A",
		Target:   "目标A",
		Position: model.Position{X: 0, Y: 0},
	})
	assert.NoError(t, err)
	nodeAID := nodeAResp.ID

	nodeBResp, err := nodeSvc.CreateNode(ctx, mapID, dto.CreateNodeRequest{
		MapID:    mapID,
		ParentID: uuid.Nil.String(),
		NodeType: "analysis",
		Question: "节点B",
		Target:   "目标B",
		Position: model.Position{X: 10, Y: 10},
	})
	assert.NoError(t, err)
	nodeBID := nodeBResp.ID

	// 3. 设置A依赖B
	cm := NewContextManager(nodeSvc.nodeRepo, mapSvc.mapRepo, nil)
	err = cm.UpdateNodeDependencies(ctx, nodeAID, []string{nodeBID})
	assert.NoError(t, err)

	// 4. 测试循环依赖检查
	dc := NewDependencyChecker(nodeSvc.nodeRepo)

	// 5. 检查B依赖A是否会产生循环（应该会）
	valid, err := dc.ValidateDependencyCycle(ctx, nodeBID, nodeAID)
	assert.NoError(t, err)
	assert.False(t, valid, "B依赖A会产生循环依赖")

	// 6. 检查A依赖B是否会产生循环（不会，因为已经存在）
	valid, err = dc.ValidateDependencyCycle(ctx, nodeAID, nodeBID)
	assert.NoError(t, err)
	assert.True(t, valid, "A依赖B不会产生新的循环")
}

func TestDependencyChecker_GetBlockedNodes(t *testing.T) {
	ctx := context.Background()
	userID := uuid.NewString()

	// 1. 创建测试数据
	createMapReq := dto.CreateMapRequest{
		Title:       "阻塞节点测试",
		Problem:     "测试阻塞节点功能",
		ProblemType: "测试类型",
		Target:      "测试目标",
		KeyPoints:   []string{"要点1"},
		Constraints: []string{"约束1"},
	}
	mapResp, err := mapSvc.CreateMap(ctx, createMapReq, userID)
	assert.NoError(t, err)
	mapID := mapResp.ID

	// 2. 创建节点A（被依赖的节点）
	nodeAResp, err := nodeSvc.CreateNode(ctx, mapID, dto.CreateNodeRequest{
		MapID:    mapID,
		ParentID: uuid.Nil.String(),
		NodeType: "analysis",
		Question: "节点A",
		Target:   "目标A",
		Position: model.Position{X: 0, Y: 0},
	})
	assert.NoError(t, err)
	nodeAID := nodeAResp.ID

	// 3. 创建节点B和C，都依赖A
	nodeBResp, err := nodeSvc.CreateNode(ctx, mapID, dto.CreateNodeRequest{
		MapID:    mapID,
		ParentID: uuid.Nil.String(),
		NodeType: "analysis",
		Question: "节点B",
		Target:   "目标B",
		Position: model.Position{X: 10, Y: 10},
	})
	assert.NoError(t, err)
	nodeBID := nodeBResp.ID

	nodeCResp, err := nodeSvc.CreateNode(ctx, mapID, dto.CreateNodeRequest{
		MapID:    mapID,
		ParentID: uuid.Nil.String(),
		NodeType: "analysis",
		Question: "节点C",
		Target:   "目标C",
		Position: model.Position{X: 20, Y: 20},
	})
	assert.NoError(t, err)
	nodeCID := nodeCResp.ID

	// 4. 设置依赖关系
	cm := NewContextManager(nodeSvc.nodeRepo, mapSvc.mapRepo, nil)
	err = cm.UpdateNodeDependencies(ctx, nodeBID, []string{nodeAID})
	assert.NoError(t, err)
	err = cm.UpdateNodeDependencies(ctx, nodeCID, []string{nodeAID})
	assert.NoError(t, err)

	// 5. 测试获取被A阻塞的节点
	dc := NewDependencyChecker(nodeSvc.nodeRepo)
	blockedNodes, err := dc.GetBlockedNodes(ctx, nodeAID)
	assert.NoError(t, err)
	assert.Len(t, blockedNodes, 2)
	assert.Contains(t, blockedNodes, nodeBID)
	assert.Contains(t, blockedNodes, nodeCID)
}

// findPosition 在切片中查找元素的位置
func findPosition(slice []string, element string) int {
	for i, v := range slice {
		if v == element {
			return i
		}
	}
	return -1
}