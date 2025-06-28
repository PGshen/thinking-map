/*
 * @Date: 2025-06-26 23:24:56
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-06-26 23:43:13
 * @FilePath: /thinking-map/server/internal/service/map_test.go
 */
package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/PGshen/thinking-map/server/internal/model/dto"
)

func TestMapService_CRUD(t *testing.T) {
	ctx := context.Background()
	userID := uuid.NewString()

	// 1. CreateMap
	createReq := dto.CreateMapRequest{
		Problem:     "测试问题",
		ProblemType: "类型A",
		Target:      "测试目标",
		KeyPoints:   []string{"要点1", "要点2"},
		Constraints: []string{"约束1"},
	}
	mapResp, err := mapSvc.CreateMap(ctx, createReq, userID)
	assert.NoError(t, err)
	assert.Equal(t, createReq.Problem, mapResp.Problem)
	assert.Equal(t, createReq.ProblemType, mapResp.ProblemType)
	assert.Equal(t, createReq.Target, mapResp.Target)
	assert.ElementsMatch(t, createReq.KeyPoints, mapResp.KeyPoints)
	assert.ElementsMatch(t, createReq.Constraints, mapResp.Constraints)
	assert.NotEmpty(t, mapResp.ID)
	assert.NotEmpty(t, mapResp.RootNodeID)

	mapID := mapResp.ID

	// 2. GetMap
	gotMap, err := mapSvc.GetMap(ctx, mapID)
	assert.NoError(t, err)
	assert.Equal(t, mapID, gotMap.ID)
	assert.Equal(t, createReq.Problem, gotMap.Problem)

	// 3. ListMaps
	listQuery := dto.MapListQuery{Page: 1, Limit: 10, Status: 0}
	listResp, err := mapSvc.ListMaps(ctx, listQuery, userID)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, listResp.Total, 1)
	found := false
	for _, item := range listResp.Items {
		if item.ID == mapID {
			found = true
			break
		}
	}
	assert.True(t, found, "created map should be in list")

	// 4. UpdateMap
	updateReq := dto.UpdateMapRequest{
		Status:      2,
		Problem:     "新问题",
		ProblemType: "类型B",
		Target:      "新目标",
		KeyPoints:   []string{"新要点1"},
		Constraints: []string{"新约束1"},
		Conclusion:  "结论内容",
	}
	updatedMap, err := mapSvc.UpdateMap(ctx, mapID, updateReq, userID)
	assert.NoError(t, err)
	assert.Equal(t, updateReq.Problem, updatedMap.Problem)
	assert.Equal(t, updateReq.ProblemType, updatedMap.ProblemType)
	assert.Equal(t, updateReq.Target, updatedMap.Target)
	assert.ElementsMatch(t, updateReq.KeyPoints, updatedMap.KeyPoints)
	assert.ElementsMatch(t, updateReq.Constraints, updatedMap.Constraints)
	assert.Equal(t, updateReq.Conclusion, updatedMap.Conclusion)
	assert.Equal(t, updateReq.Status, updatedMap.Status)

	// 5. DeleteMap
	err = mapSvc.DeleteMap(ctx, mapID, userID)
	assert.NoError(t, err)

	// 6. GetMap (should not exist)
	_, err = mapSvc.GetMap(ctx, mapID)
	assert.Error(t, err)
}
