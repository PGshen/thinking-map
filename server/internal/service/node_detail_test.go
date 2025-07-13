/*
 * @Date: 2025-06-27
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @FilePath: /thinking-map/server/internal/service/node_detail_test.go
 */
package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/PGshen/thinking-map/server/internal/model"
	"github.com/PGshen/thinking-map/server/internal/model/dto"
)

func TestNodeDetailService_CRUD(t *testing.T) {
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

	// 2. 创建一个节点以便挂载节点详情
	createNodeReq := dto.CreateNodeRequest{
		MapID:    mapResp.ID,
		ParentID: "",
		NodeType: "analysis",
		Question: "分析问题?",
		Target:   "分析目标",
		Position: model.Position{X: 10, Y: 20},
	}
	nodeResp, err := nodeSvc.CreateNode(ctx, mapID, createNodeReq)
	assert.NoError(t, err)
	assert.NotEmpty(t, nodeResp.ID)
	nodeID := nodeResp.ID

	// 3. CreateNodeDetail
	createDetailReq := dto.CreateNodeDetailRequest{
		DetailType: "custom",
		Content: model.DetailContent{
			Question: "自定义问题",
			Target:   "自定义目标",
			Context: []model.ContextInfo{
				{
					NodeID:     nodeID,
					Type:       "analysis",
					Question:   "上下文问题",
					Target:     "上下文目标",
					Conclusion: "上下文结论",
				},
			},
			DecomposeResult: []model.DecomposeResult{
				{
					Question: "分解问题1",
					Target:   "分解目标1",
				},
			},
			Conclusion: "测试结论",
		},
		Status: 1,
		Metadata: map[string]interface{}{
			"custom_field": "custom_value",
			"priority":     5,
		},
	}
	detailResp, err := nodeDetailSvc.CreateNodeDetail(ctx, nodeID, createDetailReq)
	assert.NoError(t, err)
	assert.Equal(t, createDetailReq.DetailType, detailResp.DetailType)
	assert.Equal(t, createDetailReq.Content.Question, detailResp.Content.Question)
	assert.Equal(t, createDetailReq.Content.Target, detailResp.Content.Target)
	assert.Equal(t, createDetailReq.Status, detailResp.Status)
	assert.Equal(t, createDetailReq.Metadata["custom_field"], detailResp.Metadata["custom_field"])
	assert.NotEmpty(t, detailResp.ID)
	assert.Equal(t, nodeID, detailResp.NodeID)

	detailID := detailResp.ID

	// 4. GetNodeDetails
	details, err := nodeDetailSvc.GetNodeDetails(ctx, nodeID)
	assert.NoError(t, err)
	found := false
	for _, d := range details {
		if d.ID == detailID {
			found = true
			assert.Equal(t, createDetailReq.DetailType, d.DetailType)
			assert.Equal(t, createDetailReq.Content.Question, d.Content.Question)
			break
		}
	}
	assert.True(t, found, "created detail should be in list")

	// 5. UpdateNodeDetail
	updateReq := dto.UpdateNodeDetailRequest{
		Content: model.DetailContent{
			Question: "更新的问题",
			Target:   "更新的目标",
			Context: []model.ContextInfo{
				{
					NodeID:     nodeID,
					Type:       "updated",
					Question:   "更新的上下文问题",
					Target:     "更新的上下文目标",
					Conclusion: "更新的上下文结论",
				},
			},
			DecomposeResult: []model.DecomposeResult{
				{
					Question: "更新的分解问题",
					Target:   "更新的分解目标",
				},
			},
			Conclusion: "更新的结论",
		},
		Status: 2,
		Metadata: map[string]interface{}{
			"updated_field": "updated_value",
			"priority":      10,
		},
	}
	updatedDetail, err := nodeDetailSvc.UpdateNodeDetail(ctx, detailID, updateReq)
	assert.NoError(t, err)
	assert.Equal(t, updateReq.Content.Question, updatedDetail.Content.Question)
	assert.Equal(t, updateReq.Content.Target, updatedDetail.Content.Target)
	assert.Equal(t, updateReq.Status, updatedDetail.Status)
	assert.Equal(t, updateReq.Metadata["updated_field"], updatedDetail.Metadata["updated_field"])

	// 6. DeleteNodeDetail
	err = nodeDetailSvc.DeleteNodeDetail(ctx, detailID)
	assert.NoError(t, err)

	// 7. GetNodeDetails (should not find deleted detail)
	details, err = nodeDetailSvc.GetNodeDetails(ctx, nodeID)
	assert.NoError(t, err)
	for _, d := range details {
		assert.NotEqual(t, detailID, d.ID, "deleted detail should not be in list")
	}
}

func TestNodeDetailService_CreateMultipleDetails(t *testing.T) {
	ctx := context.Background()
	userID := uuid.NewString()

	// 1. 创建 map 和节点
	createMapReq := dto.CreateMapRequest{
		Problem:     "多详情测试问题",
		ProblemType: "类型B",
		Target:      "多详情测试目标",
		KeyPoints:   []string{"要点1", "要点2"},
		Constraints: []string{"约束1", "约束2"},
	}
	mapResp, err := mapSvc.CreateMap(ctx, createMapReq, userID)
	assert.NoError(t, err)
	mapID := mapResp.ID

	createNodeReq := dto.CreateNodeRequest{
		MapID:    mapResp.ID,
		NodeType: "analysis",
		Question: "多详情分析问题?",
		Target:   "多详情分析目标",
		Position: model.Position{X: 50, Y: 100},
	}
	nodeResp, err := nodeSvc.CreateNode(ctx, mapID, createNodeReq)
	assert.NoError(t, err)
	nodeID := nodeResp.ID

	// 2. 创建多个不同类型的节点详情
	detailTypes := []string{"info", "decompose", "conclusion", "custom"}
	var detailIDs []string

	for i, detailType := range detailTypes {
		createDetailReq := dto.CreateNodeDetailRequest{
			DetailType: detailType,
			Content: model.DetailContent{
				Question:   fmt.Sprintf("%s类型问题%d", detailType, i+1),
				Target:     fmt.Sprintf("%s类型目标%d", detailType, i+1),
				Conclusion: fmt.Sprintf("%s类型结论%d", detailType, i+1),
			},
			Status: i + 1,
			Metadata: map[string]interface{}{
				"type_index": i,
				"type_name":  detailType,
			},
		}
		var detailResp *dto.NodeDetailResponse
		detailResp, err = nodeDetailSvc.CreateNodeDetail(ctx, nodeID, createDetailReq)
		assert.NoError(t, err)
		assert.Equal(t, detailType, detailResp.DetailType)
		assert.Equal(t, createDetailReq.Content.Question, detailResp.Content.Question)
		detailIDs = append(detailIDs, detailResp.ID)
	}

	// 3. 验证所有详情都被创建
	details, err := nodeDetailSvc.GetNodeDetails(ctx, nodeID)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(details), len(detailTypes))

	// 验证每种类型都存在
	typeCount := make(map[string]int)
	for _, detail := range details {
		typeCount[detail.DetailType]++
	}
	for _, detailType := range detailTypes {
		assert.GreaterOrEqual(t, typeCount[detailType], 1, "should have at least one detail of type %s", detailType)
	}

	// 4. 清理测试数据
	for _, detailID := range detailIDs {
		err := nodeDetailSvc.DeleteNodeDetail(ctx, detailID)
		assert.NoError(t, err)
	}
}

func TestNodeDetailService_UpdateContentOnly(t *testing.T) {
	ctx := context.Background()
	userID := uuid.NewString()

	// 1. 创建 map 和节点
	createMapReq := dto.CreateMapRequest{
		Problem:     "内容更新测试问题",
		ProblemType: "类型C",
		Target:      "内容更新测试目标",
		KeyPoints:   []string{"要点1"},
		Constraints: []string{"约束1"},
	}
	mapResp, err := mapSvc.CreateMap(ctx, createMapReq, userID)
	assert.NoError(t, err)
	mapID := mapResp.ID

	createNodeReq := dto.CreateNodeRequest{
		MapID:    mapResp.ID,
		NodeType: "analysis",
		Question: "内容更新分析问题?",
		Target:   "内容更新分析目标",
		Position: model.Position{X: 100, Y: 150},
	}
	nodeResp, err := nodeSvc.CreateNode(ctx, mapID, createNodeReq)
	assert.NoError(t, err)
	nodeID := nodeResp.ID

	// 2. 创建节点详情
	createDetailReq := dto.CreateNodeDetailRequest{
		DetailType: "content_test",
		Content: model.DetailContent{
			Question:   "原始问题",
			Target:     "原始目标",
			Conclusion: "原始结论",
		},
		Status: 1,
		Metadata: map[string]interface{}{
			"original": true,
		},
	}
	detailResp, err := nodeDetailSvc.CreateNodeDetail(ctx, nodeID, createDetailReq)
	assert.NoError(t, err)
	detailID := detailResp.ID

	// 3. 只更新内容，保持其他字段不变
	updateReq := dto.UpdateNodeDetailRequest{
		Content: model.DetailContent{
			Question:   "更新的问题",
			Target:     "更新的目标",
			Conclusion: "更新的结论",
			Context: []model.ContextInfo{
				{
					NodeID:     nodeID,
					Type:       "context",
					Question:   "上下文问题",
					Target:     "上下文目标",
					Conclusion: "上下文结论",
				},
			},
		},
		Status: 1, // 保持原状态
		Metadata: map[string]interface{}{
			"original": true, // 保持原元数据
		},
	}
	updatedDetail, err := nodeDetailSvc.UpdateNodeDetail(ctx, detailID, updateReq)
	assert.NoError(t, err)
	assert.Equal(t, updateReq.Content.Question, updatedDetail.Content.Question)
	assert.Equal(t, updateReq.Content.Target, updatedDetail.Content.Target)
	assert.Equal(t, updateReq.Content.Conclusion, updatedDetail.Content.Conclusion)
	assert.Equal(t, 1, len(updatedDetail.Content.Context))
	assert.Equal(t, createDetailReq.Status, updatedDetail.Status)                             // 状态应该保持不变
	assert.Equal(t, createDetailReq.Metadata["original"], updatedDetail.Metadata["original"]) // 元数据应该保持不变

	// 4. 清理测试数据
	err = nodeDetailSvc.DeleteNodeDetail(ctx, detailID)
	assert.NoError(t, err)
}
