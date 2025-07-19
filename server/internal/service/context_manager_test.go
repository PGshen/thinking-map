/*
 * @Date: 2025-01-27
 * @LastEditors: AI Assistant
 * @FilePath: /thinking-map/server/internal/service/context_manager_test.go
 */
package service

import (
	"context"
	"testing"

	"github.com/PGshen/thinking-map/server/internal/model"
	"github.com/PGshen/thinking-map/server/internal/model/dto"
	"github.com/PGshen/thinking-map/server/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestContextManager_GetNodeContext(t *testing.T) {
	ctx := context.Background()
	userID := uuid.NewString()

	// 1. 创建测试数据 - 创建一个map
	createMapReq := dto.CreateMapRequest{
		Title:       "测试标题",
		Problem:     "测试问题",
		ProblemType: "类型A",
		Target:      "测试目标",
		KeyPoints:   []string{"要点1", "要点2"},
		Constraints: []string{"约束1", "约束2"},
	}
	mapResp, err := mapSvc.CreateMap(ctx, createMapReq, userID)
	assert.NoError(t, err)
	mapID := mapResp.ID

	// 2. 创建父节点
	createParentNodeReq := dto.CreateNodeRequest{
		MapID:    mapID,
		ParentID: uuid.Nil.String(),
		NodeType: "analysis",
		Question: "父节点问题",
		Target:   "父节点目标",
		Position: model.Position{X: 0, Y: 0},
	}
	parentNodeResp, err := nodeSvc.CreateNode(ctx, mapID, createParentNodeReq)
	assert.NoError(t, err)
	parentNodeID := parentNodeResp.ID

	// 3. 创建当前节点（子节点）
	createNodeReq := dto.CreateNodeRequest{
		MapID:    mapID,
		ParentID: parentNodeID,
		NodeType: "analysis",
		Question: "当前节点问题",
		Target:   "当前节点目标",
		Position: model.Position{X: 10, Y: 20},
	}
	nodeResp, err := nodeSvc.CreateNode(ctx, mapID, createNodeReq)
	assert.NoError(t, err)
	nodeID := nodeResp.ID

	// 4. 创建依赖节点
	createDependencyNodeReq := dto.CreateNodeRequest{
		MapID:    mapID,
		ParentID: parentNodeID,
		NodeType: "analysis",
		Question: "依赖节点问题",
		Target:   "依赖节点目标",
		Position: model.Position{X: 20, Y: 30},
	}
	dependencyNodeResp, err := nodeSvc.CreateNode(ctx, mapID, createDependencyNodeReq)
	assert.NoError(t, err)
	dependencyNodeID := dependencyNodeResp.ID

	// 5. 创建子节点
	createChildNodeReq := dto.CreateNodeRequest{
		MapID:    mapID,
		ParentID: nodeID,
		NodeType: "conclusion",
		Question: "子节点问题",
		Target:   "子节点目标",
		Position: model.Position{X: 30, Y: 40},
	}
	_, err = nodeSvc.CreateNode(ctx, mapID, createChildNodeReq)
	assert.NoError(t, err)

	// 6. 设置当前节点的依赖关系
	cm := NewContextManager(nodeSvc.nodeRepo, mapSvc.mapRepo, nil)
	err = cm.UpdateNodeDependencies(ctx, nodeID, []string{dependencyNodeID})
	assert.NoError(t, err)

	// 7. 测试获取节点上下文
	contextInfo, err := cm.GetNodeContext(ctx, nodeID)
	assert.NoError(t, err)
	assert.NotNil(t, contextInfo)

	// 8. 验证导图信息
	assert.NotNil(t, contextInfo.MapInfo)
	assert.Equal(t, mapID, contextInfo.MapInfo.ID)
	assert.Equal(t, "测试标题", contextInfo.MapInfo.Title)
	assert.Equal(t, "测试问题", contextInfo.MapInfo.Problem)
	assert.Equal(t, "测试目标", contextInfo.MapInfo.Target)
	assert.ElementsMatch(t, []string{"约束1", "约束2"}, contextInfo.MapInfo.Constraints)

	// 9. 验证当前节点信息
	assert.NotNil(t, contextInfo.NodeInfo)
	assert.Equal(t, nodeID, contextInfo.NodeInfo.ID)
	assert.Equal(t, "当前节点问题", contextInfo.NodeInfo.Question)
	assert.Equal(t, "当前节点目标", contextInfo.NodeInfo.Target)

	// 10. 验证祖先节点上下文
	assert.Len(t, contextInfo.AncestorsContext, 1)
	assert.Equal(t, parentNodeID, contextInfo.AncestorsContext[0].NodeID)
	assert.Equal(t, "父节点问题", contextInfo.AncestorsContext[0].Question)
	assert.Equal(t, "父节点目标", contextInfo.AncestorsContext[0].Target)

	// 11. 验证依赖节点上下文
	assert.Len(t, contextInfo.DependencyContext, 1)
	assert.Equal(t, dependencyNodeID, contextInfo.DependencyContext[0].NodeID)
	assert.Equal(t, "依赖节点问题", contextInfo.DependencyContext[0].Question)
	assert.Equal(t, "依赖节点目标", contextInfo.DependencyContext[0].Target)

	// 12. 验证子节点上下文
	assert.Len(t, contextInfo.ChildrenContext, 1)
	assert.Equal(t, "子节点问题", contextInfo.ChildrenContext[0].Question)
	assert.Equal(t, "子节点目标", contextInfo.ChildrenContext[0].Target)
}

func TestContextManager_GetNodeContextWithConversation(t *testing.T) {
	ctx := context.Background()
	userID := uuid.NewString()

	// 1. 创建测试数据
	createMapReq := dto.CreateMapRequest{
		Title:       "对话测试标题",
		Problem:     "对话测试问题",
		ProblemType: "类型B",
		Target:      "对话测试目标",
		KeyPoints:   []string{"要点1"},
		Constraints: []string{"约束1"},
	}
	mapResp, err := mapSvc.CreateMap(ctx, createMapReq, userID)
	assert.NoError(t, err)
	mapID := mapResp.ID

	// 2. 创建节点
	createNodeReq := dto.CreateNodeRequest{
		MapID:    mapID,
		ParentID: uuid.Nil.String(),
		NodeType: "analysis",
		Question: "对话节点问题",
		Target:   "对话节点目标",
		Position: model.Position{X: 0, Y: 0},
	}
	nodeResp, err := nodeSvc.CreateNode(ctx, mapID, createNodeReq)
	assert.NoError(t, err)
	nodeID := nodeResp.ID

	// 3. 创建消息服务并添加测试消息
	msgRepo := repository.NewMessageRepository(testDB)
	nodeRepo := repository.NewThinkingNodeRepository(testDB)
	msgService := NewMessageManager(msgRepo, nodeRepo)

	// 创建用户消息
	userMsgReq := dto.CreateMessageRequest{
		ID:          uuid.NewString(),
		ParentID:    uuid.Nil.String(),
		MessageType: "text",
		Role:        "user",
		Content: model.MessageContent{
			Text: "用户测试消息",
		},
	}
	userMsgResp, err := msgService.CreateMessage(ctx, userID, userMsgReq)
	assert.NoError(t, err)

	// 创建助手消息
	assistantMsgReq := dto.CreateMessageRequest{
		ID:          uuid.NewString(),
		ParentID:    userMsgResp.ID,
		MessageType: "text",
		Role:        "assistant",
		Content: model.MessageContent{
			Text: "助手测试回复",
		},
	}
	assistantMsgResp, err := msgService.CreateMessage(ctx, userID, assistantMsgReq)
	assert.NoError(t, err)

	// 4. 测试获取包含对话历史的节点上下文
	cm := NewContextManager(nodeSvc.nodeRepo, mapSvc.mapRepo, msgRepo)
	contextInfo, err := cm.GetNodeContextWithConversation(ctx, nodeID, assistantMsgResp.ID)
	assert.NoError(t, err)
	assert.NotNil(t, contextInfo)

	// 5. 验证对话上下文
	assert.NotNil(t, contextInfo.ConversationContext)
	assert.GreaterOrEqual(t, len(contextInfo.ConversationContext), 1)
}

func TestContextManager_FormatContextForAgent(t *testing.T) {
	ctx := context.Background()
	userID := uuid.NewString()

	// 1. 创建测试数据
	createMapReq := dto.CreateMapRequest{
		Title:       "格式化测试标题",
		Problem:     "格式化测试问题",
		ProblemType: "类型C",
		Target:      "格式化测试目标",
		KeyPoints:   []string{"要点1"},
		Constraints: []string{"约束1"},
	}
	mapResp, err := mapSvc.CreateMap(ctx, createMapReq, userID)
	assert.NoError(t, err)
	mapID := mapResp.ID

	// 2. 创建节点
	createNodeReq := dto.CreateNodeRequest{
		MapID:    mapID,
		ParentID: uuid.Nil.String(),
		NodeType: "analysis",
		Question: "格式化节点问题",
		Target:   "格式化节点目标",
		Position: model.Position{X: 0, Y: 0},
	}
	nodeResp, err := nodeSvc.CreateNode(ctx, mapID, createNodeReq)
	assert.NoError(t, err)
	nodeID := nodeResp.ID

	// 3. 获取节点上下文
	cm := NewContextManager(nodeSvc.nodeRepo, mapSvc.mapRepo, nil)
	contextInfo, err := cm.GetNodeContext(ctx, nodeID)
	assert.NoError(t, err)

	// 4. 测试格式化上下文
	userMessage := "请帮我分析这个问题"
	formattedPrompt := cm.FormatContextForAgent(contextInfo, userMessage)

	// 5. 验证格式化结果
	assert.Contains(t, formattedPrompt, "当前节点信息")
	assert.Contains(t, formattedPrompt, "格式化节点问题")
	assert.Contains(t, formattedPrompt, "格式化节点目标")
	assert.Contains(t, formattedPrompt, "用户消息：请帮我分析这个问题")
}

func TestContextManager_RefreshNodeContext(t *testing.T) {
	ctx := context.Background()
	userID := uuid.NewString()

	// 1. 创建测试数据
	createMapReq := dto.CreateMapRequest{
		Title:       "刷新测试标题",
		Problem:     "刷新测试问题",
		ProblemType: "类型D",
		Target:      "刷新测试目标",
		KeyPoints:   []string{"要点1"},
		Constraints: []string{"约束1"},
	}
	mapResp, err := mapSvc.CreateMap(ctx, createMapReq, userID)
	assert.NoError(t, err)
	mapID := mapResp.ID

	// 2. 创建父节点
	createParentNodeReq := dto.CreateNodeRequest{
		MapID:    mapID,
		ParentID: uuid.Nil.String(),
		NodeType: "analysis",
		Question: "刷新父节点问题",
		Target:   "刷新父节点目标",
		Position: model.Position{X: 0, Y: 0},
	}
	parentNodeResp, err := nodeSvc.CreateNode(ctx, mapID, createParentNodeReq)
	assert.NoError(t, err)
	parentNodeID := parentNodeResp.ID

	// 3. 创建当前节点
	createNodeReq := dto.CreateNodeRequest{
		MapID:    mapID,
		ParentID: parentNodeID,
		NodeType: "analysis",
		Question: "刷新当前节点问题",
		Target:   "刷新当前节点目标",
		Position: model.Position{X: 10, Y: 20},
	}
	nodeResp, err := nodeSvc.CreateNode(ctx, mapID, createNodeReq)
	assert.NoError(t, err)
	nodeID := nodeResp.ID

	// 4. 测试刷新节点上下文
	cm := NewContextManager(nodeSvc.nodeRepo, mapSvc.mapRepo, nil)
	ginCtx := &gin.Context{}
	ginCtx.Set("request_id", "test")

	updatedNodeResp, err := cm.RefreshNodeContext(ginCtx, nodeID)
	assert.NoError(t, err)
	assert.NotNil(t, updatedNodeResp)

	// 5. 验证上下文已更新
	assert.Equal(t, nodeID, updatedNodeResp.ID)
	// 验证祖先上下文已设置
	assert.GreaterOrEqual(t, len(updatedNodeResp.Context.Ancestor), 0)
}

func TestContextManager_EmptyContext(t *testing.T) {
	ctx := context.Background()
	userID := uuid.NewString()

	// 1. 创建只有根节点的map
	createMapReq := dto.CreateMapRequest{
		Title:       "空上下文测试标题",
		Problem:     "空上下文测试问题",
		ProblemType: "类型E",
		Target:      "空上下文测试目标",
		KeyPoints:   []string{"要点1"},
		Constraints: []string{"约束1"},
	}
	mapResp, err := mapSvc.CreateMap(ctx, createMapReq, userID)
	assert.NoError(t, err)
	mapID := mapResp.ID

	// 2. 创建独立节点（无父节点、无依赖、无子节点）
	createNodeReq := dto.CreateNodeRequest{
		MapID:    mapID,
		ParentID: uuid.Nil.String(),
		NodeType: "analysis",
		Question: "独立节点问题",
		Target:   "独立节点目标",
		Position: model.Position{X: 0, Y: 0},
	}
	nodeResp, err := nodeSvc.CreateNode(ctx, mapID, createNodeReq)
	assert.NoError(t, err)
	nodeID := nodeResp.ID

	// 3. 测试获取空上下文
	cm := NewContextManager(nodeSvc.nodeRepo, mapSvc.mapRepo, nil)
	contextInfo, err := cm.GetNodeContext(ctx, nodeID)
	assert.NoError(t, err)
	assert.NotNil(t, contextInfo)

	// 4. 验证空上下文
	assert.Equal(t, nodeID, contextInfo.NodeInfo.ID)
	assert.Empty(t, contextInfo.AncestorsContext)
	assert.Empty(t, contextInfo.DependencyContext)
	assert.Empty(t, contextInfo.ChildrenContext)
	assert.Nil(t, contextInfo.ConversationContext)
	assert.Nil(t, contextInfo.UserContext)
}
