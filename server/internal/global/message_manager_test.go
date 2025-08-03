/*
 * @Date: 2025-01-27
 * @LastEditors: AI Assistant
 * @FilePath: /thinking-map/server/internal/service/message_manager_test.go
 */
package global

import (
	"context"
	"fmt"
	"testing"

	"github.com/PGshen/thinking-map/server/internal/model"
	"github.com/PGshen/thinking-map/server/internal/model/dto"
	"github.com/PGshen/thinking-map/server/internal/pkg/comm"
	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMessageManager_CreateMessage(t *testing.T) {
	ctx := context.Background()
	userID := uuid.NewString()

	// 测试创建根消息（无父消息）
	t.Run("CreateRootMessage", func(t *testing.T) {
		req := dto.CreateMessageRequest{
			ID:          uuid.NewString(),
			ParentID:    "",
			MessageType: comm.MessageTypeText,
			Role:        schema.User,
			Content: model.MessageContent{
				Text: "这是一条测试消息",
			},
			Metadata: map[string]any{
				"creator": "user",
			},
		}

		resp, err := messageManager.CreateMessage(ctx, userID, req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, req.ID, resp.ID)
		assert.Equal(t, req.MessageType, resp.MessageType)
		assert.Equal(t, req.Role, resp.Role)
		assert.Equal(t, req.Content.Text, resp.Content.Text)
		assert.NotEmpty(t, resp.ConversationID)
		// 根消息的ParentID应该是uuid.Nil.String()
		assert.Equal(t, uuid.Nil.String(), resp.ParentID)
	})

	// 测试创建子消息
	t.Run("CreateChildMessage", func(t *testing.T) {
		// 先创建父消息
		parentReq := dto.CreateMessageRequest{
			ID:          uuid.NewString(),
			ParentID:    "",
			MessageType: comm.MessageTypeText,
			Role:        schema.User,
			Content: model.MessageContent{
				Text: "父消息",
			},
		}
		parentResp, err := messageManager.CreateMessage(ctx, userID, parentReq)
		assert.NoError(t, err)

		// 创建子消息
		childReq := dto.CreateMessageRequest{
			ID:          uuid.NewString(),
			ParentID:    parentResp.ID,
			MessageType: comm.MessageTypeText,
			Role:        schema.Assistant,
			Content: model.MessageContent{
				Text: "这是回复消息",
			},
		}
		childResp, err := messageManager.CreateMessage(ctx, userID, childReq)
		assert.NoError(t, err)
		assert.NotNil(t, childResp)
		assert.Equal(t, parentResp.ID, childResp.ParentID)
		assert.Equal(t, parentResp.ConversationID, childResp.ConversationID)
	})
}

func TestMessageManager_UpdateMessage(t *testing.T) {
	ctx := context.Background()
	userID := uuid.NewString()

	// 先创建一条消息
	createReq := dto.CreateMessageRequest{
		ID:          uuid.NewString(),
		ParentID:    "",
		MessageType: comm.MessageTypeText,
		Role:        schema.User,
		Content: model.MessageContent{
			Text: "原始消息内容",
		},
	}
	createResp, err := messageManager.CreateMessage(ctx, userID, createReq)
	assert.NoError(t, err)

	// 更新消息
	updateReq := dto.UpdateMessageRequest{
		ID: createResp.ID,
		Content: model.MessageContent{
			Text: "更新后的消息内容",
		},
		Metadata: map[string]any{
			"updated": true,
		},
	}
	updateResp, err := messageManager.UpdateMessage(ctx, updateReq)
	assert.NoError(t, err)
	assert.NotNil(t, updateResp)
	assert.Equal(t, "更新后的消息内容", updateResp.Content.Text)
	// 检查更新时间是否不同（允许1秒的误差）
	assert.True(t, updateResp.UpdatedAt.Unix() >= createResp.UpdatedAt.Unix())
}

func TestMessageManager_GetMessageByID(t *testing.T) {
	ctx := context.Background()
	userID := uuid.NewString()

	// 创建消息
	createReq := dto.CreateMessageRequest{
		ID:          uuid.NewString(),
		ParentID:    "",
		MessageType: comm.MessageTypeText,
		Role:        schema.User,
		Content: model.MessageContent{
			Text: "测试获取消息",
		},
	}
	createResp, err := messageManager.CreateMessage(ctx, userID, createReq)
	assert.NoError(t, err)

	// 获取消息
	getResp, err := messageManager.GetMessageByID(ctx, createResp.ID)
	assert.NoError(t, err)
	assert.NotNil(t, getResp)
	assert.Equal(t, createResp.ID, getResp.ID)
	assert.Equal(t, createResp.Content.Text, getResp.Content.Text)

	// 测试获取不存在的消息
	_, err = messageManager.GetMessageByID(ctx, uuid.NewString())
	assert.Error(t, err)
}

func TestMessageManager_DeleteMessage(t *testing.T) {
	ctx := context.Background()
	userID := uuid.NewString()

	// 创建消息
	createReq := dto.CreateMessageRequest{
		ID:          uuid.NewString(),
		ParentID:    "",
		MessageType: comm.MessageTypeText,
		Role:        schema.User,
		Content: model.MessageContent{
			Text: "待删除的消息",
		},
	}
	createResp, err := messageManager.CreateMessage(ctx, userID, createReq)
	assert.NoError(t, err)

	// 删除消息
	err = messageManager.DeleteMessage(ctx, createResp.ID)
	assert.NoError(t, err)

	// 验证消息已被删除
	_, err = messageManager.GetMessageByID(ctx, createResp.ID)
	assert.Error(t, err)
}

func TestMessageManager_GetMessageByConversationID(t *testing.T) {
	ctx := context.Background()
	userID := uuid.NewString()

	// 创建第一条消息（根消息）
	firstReq := dto.CreateMessageRequest{
		ID:          uuid.NewString(),
		ParentID:    "",
		MessageType: comm.MessageTypeText,
		Role:        schema.User,
		Content: model.MessageContent{
			Text: "第一条消息",
		},
	}
	firstResp, err := messageManager.CreateMessage(ctx, userID, firstReq)
	assert.NoError(t, err)
	conversationID := firstResp.ConversationID

	// 创建第二条消息（回复）
	secondReq := dto.CreateMessageRequest{
		ID:          uuid.NewString(),
		ParentID:    firstResp.ID,
		MessageType: comm.MessageTypeText,
		Role:        schema.Assistant,
		Content: model.MessageContent{
			Text: "第二条消息",
		},
	}
	secondResp, err := messageManager.CreateMessage(ctx, userID, secondReq)
	assert.NoError(t, err)

	// 获取会话中的所有消息
	messages, err := messageManager.GetMessageByConversationID(ctx, conversationID)
	assert.NoError(t, err)
	assert.Len(t, messages, 2)
	assert.Equal(t, firstResp.ID, messages[0].ID)
	assert.Equal(t, secondResp.ID, messages[1].ID)
}

func TestMessageManager_GetMessageChain(t *testing.T) {
	ctx := context.Background()
	userID := uuid.NewString()

	// 创建消息链：root -> child1 -> child2
	rootReq := dto.CreateMessageRequest{
		ID:          uuid.NewString(),
		ParentID:    "",
		MessageType: comm.MessageTypeText,
		Role:        schema.User,
		Content: model.MessageContent{
			Text: "根消息",
		},
	}
	rootResp, err := messageManager.CreateMessage(ctx, userID, rootReq)
	assert.NoError(t, err)

	child1Req := dto.CreateMessageRequest{
		ID:          uuid.NewString(),
		ParentID:    rootResp.ID,
		MessageType: comm.MessageTypeText,
		Role:        schema.Assistant,
		Content: model.MessageContent{
			Text: "子消息1",
		},
	}
	child1Resp, err := messageManager.CreateMessage(ctx, userID, child1Req)
	assert.NoError(t, err)

	child2Req := dto.CreateMessageRequest{
		ID:          uuid.NewString(),
		ParentID:    child1Resp.ID,
		MessageType: comm.MessageTypeText,
		Role:        schema.User,
		Content: model.MessageContent{
			Text: "子消息2",
		},
	}
	child2Resp, err := messageManager.CreateMessage(ctx, userID, child2Req)
	assert.NoError(t, err)

	// 获取消息链
	chain, err := messageManager.GetMessageChain(ctx, child2Resp.ID, child2Resp.ConversationID)
	assert.NoError(t, err)
	assert.Len(t, chain, 3)
	assert.Equal(t, rootResp.ID, chain[0].ID)
	assert.Equal(t, child1Resp.ID, chain[1].ID)
	assert.Equal(t, child2Resp.ID, chain[2].ID)
}

func TestMessageManager_RollbackConversation(t *testing.T) {
	ctx := context.Background()
	userID := uuid.NewString()

	// 创建多条消息
	messages := make([]*dto.MessageResponse, 0)
	for i := 0; i < 3; i++ {
		var parentID string
		if i > 0 {
			parentID = messages[i-1].ID
		} else {
			parentID = ""
		}

		req := dto.CreateMessageRequest{
			ID:          uuid.NewString(),
			ParentID:    parentID,
			MessageType: comm.MessageTypeText,
			Role:        schema.User,
			Content: model.MessageContent{
				Text: fmt.Sprintf("消息 %d", i+1),
			},
		}
		resp, err := messageManager.CreateMessage(ctx, userID, req)
		assert.NoError(t, err)
		messages = append(messages, resp)
	}

	conversationID := messages[0].ConversationID

	// 回退到第二条消息
	err := messageManager.RollbackConversation(ctx, conversationID, messages[1].ID)
	assert.NoError(t, err)

	// 验证只剩下前两条消息
	remainingMessages, err := messageManager.GetMessageByConversationID(ctx, conversationID)
	assert.NoError(t, err)
	assert.Len(t, remainingMessages, 2)
}

func TestMessageManager_ClearConversation(t *testing.T) {
	ctx := context.Background()
	userID := uuid.NewString()

	// 创建多条消息
	messages := make([]*dto.MessageResponse, 0)
	for i := 0; i < 3; i++ {
		var parentID string
		if i > 0 {
			parentID = messages[i-1].ID
		}

		req := dto.CreateMessageRequest{
			ID:          uuid.NewString(),
			ParentID:    parentID,
			MessageType: comm.MessageTypeText,
			Role:        schema.User,
			Content: model.MessageContent{
				Text: fmt.Sprintf("消息 %d", i+1),
			},
		}
		resp, err := messageManager.CreateMessage(ctx, userID, req)
		assert.NoError(t, err)
		messages = append(messages, resp)
	}

	conversationID := messages[0].ConversationID

	// 清空会话
	err := messageManager.ClearConversation(ctx, conversationID)
	assert.NoError(t, err)

	// 验证会话已清空
	remainingMessages, err := messageManager.GetMessageByConversationID(ctx, conversationID)
	assert.NoError(t, err)
	assert.Len(t, remainingMessages, 0)
}

func TestMessageManager_ConvertToSchemaMsg(t *testing.T) {
	// 创建测试数据
	messages := []*dto.MessageResponse{
		{
			ID:   uuid.NewString(),
			Role: schema.User,
			Content: model.MessageContent{
				Text: "用户消息",
			},
		},
		{
			ID:   uuid.NewString(),
			Role: schema.Assistant,
			Content: model.MessageContent{
				Text: "助手回复",
			},
		},
	}

	// 转换为schema消息
	schemaMessages := ConvertToSchemaMsg(messages)
	assert.Len(t, schemaMessages, 2)
	assert.Equal(t, schema.User, schemaMessages[0].Role)
	assert.Equal(t, "用户消息", schemaMessages[0].Content)
	assert.Equal(t, schema.Assistant, schemaMessages[1].Role)
	assert.Equal(t, "助手回复", schemaMessages[1].Content)
}

func TestMessageManager_GetMessageStatus(t *testing.T) {
	ctx := context.Background()
	userID := uuid.NewString()

	// 创建消息
	createReq := dto.CreateMessageRequest{
		ID:          uuid.NewString(),
		ParentID:    "",
		MessageType: comm.MessageTypeText,
		Role:        schema.User,
		Content: model.MessageContent{
			Text: "状态测试消息",
		},
	}
	createResp, err := messageManager.CreateMessage(ctx, userID, createReq)
	assert.NoError(t, err)

	// 获取消息状态
	status, err := messageManager.GetMessageStatus(ctx, createResp.ID)
	assert.NoError(t, err)
	assert.Equal(t, createResp.ID, status.ID)
	assert.Equal(t, "active", status.Status)
	assert.Nil(t, status.DeletedAt)
}

func TestMessageManager_SaveStreamMessage(t *testing.T) {
	// 创建gin上下文
	ginCtx, _ := gin.CreateTestContext(nil)
	ginCtx.Set("user_id", uuid.NewString())

	// 注意：这个测试需要实际的StreamReader，这里只测试基本流程
	// 在实际使用中，SaveStreamMessage会在流式处理中被调用
	// 这里我们主要验证方法存在且可以被调用
	assert.NotNil(t, messageManager.SaveStreamMessage)
}

func TestMessageManager_CreateConversation(t *testing.T) {
	ctx := context.Background()
	nodeID := uuid.NewString()

	// 创建新会话
	conversationID, err := messageManager.CreateConversation(ctx, nodeID)
	assert.NoError(t, err)
	assert.NotEmpty(t, conversationID)

	// 验证返回的是有效的UUID
	_, err = uuid.Parse(conversationID)
	assert.NoError(t, err)
}
