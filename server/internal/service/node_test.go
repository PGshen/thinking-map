/*
 * @Date: 2025-06-27
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @FilePath: /thinking-map/server/internal/service/node_test.go
 */
package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/PGshen/thinking-map/server/internal/model"
	"github.com/PGshen/thinking-map/server/internal/model/dto"
)

func TestNodeService_CRUD(t *testing.T) {

	ctx := context.Background()
	userID := uuid.NewString()

	// 1. 先创建一个 map 以便挂载节点
	createMapReq := dto.CreateMapRequest{
		Problem:     "测试问题",
		ProblemType: "类型A",
		Target:      "测试目标",
		KeyPoints:   []string{"要点1"},
		Constraints: []string{"约束1"},
	}
	mapResp, err := mapSvc.CreateMap(ctx, createMapReq, userID)
	assert.NoError(t, err)
	assert.NotEmpty(t, mapResp.ID)
	mapID := mapResp.ID

	// 2. CreateNode
	createNodeReq := dto.CreateNodeRequest{
		MapID:    mapResp.ID,
		NodeType: "analysis",
		Question: "分析问题?",
		Target:   "分析目标",
		Position: model.Position{X: 10, Y: 20, Width: 100, Height: 50},
	}
	nodeResp, err := nodeSvc.CreateNode(ctx, mapID, createNodeReq)
	assert.NoError(t, err)
	assert.Equal(t, createNodeReq.Question, nodeResp.Question)
	assert.Equal(t, createNodeReq.Target, nodeResp.Target)
	assert.Equal(t, createNodeReq.NodeType, nodeResp.NodeType)
	assert.Equal(t, createNodeReq.ParentID, nodeResp.ParentID)
	assert.Equal(t, mapID, nodeResp.MapID)
	assert.NotEmpty(t, nodeResp.ID)

	nodeID := nodeResp.ID

	// 3. ListNodes
	nodes, err := nodeSvc.ListNodes(ctx, mapID)
	assert.NoError(t, err)
	found := false
	for _, n := range nodes {
		if n.ID == nodeID {
			found = true
			break
		}
	}
	assert.True(t, found, "created node should be in list")

	// 4. UpdateNode
	updateReq := dto.UpdateNodeRequest{
		Question: "新问题?",
		Target:   "新目标",
		Position: model.Position{X: 30, Y: 40, Width: 120, Height: 60},
	}
	updatedNode, err := nodeSvc.UpdateNode(ctx, nodeID, updateReq)
	assert.NoError(t, err)
	assert.Equal(t, updateReq.Question, updatedNode.Question)
	assert.Equal(t, updateReq.Target, updatedNode.Target)
	assert.Equal(t, updateReq.Position, updatedNode.Position)

	// 5. DeleteNode
	err = nodeSvc.DeleteNode(ctx, nodeID)
	assert.NoError(t, err)

	// 6. ListNodes (should not find deleted node)
	nodes, err = nodeSvc.ListNodes(ctx, mapID)
	assert.NoError(t, err)
	for _, n := range nodes {
		assert.NotEqual(t, nodeID, n.ID, "deleted node should not be in list")
	}
}
