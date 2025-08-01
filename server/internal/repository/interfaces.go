/*
 * @Date: 2025-06-18 22:33:07
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-06-23 22:55:05
 * @FilePath: /thinking-map/server/internal/repository/interfaces.go
 */
package repository

import (
	"context"
	"time"

	"github.com/PGshen/thinking-map/server/internal/model"
)

const whereID = "id = ?"

// User 用户仓储接口
type User interface {
	Create(ctx context.Context, user *model.User) error
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	List(ctx context.Context, offset, limit int) ([]*model.User, int64, error)
}

// ThinkingMap 思维导图仓储接口
type ThinkingMap interface {
	Create(ctx context.Context, map_ *model.ThinkingMap, rootNode *model.ThinkingNode) error
	Update(ctx context.Context, mapID string, updates map[string]interface{}) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*model.ThinkingMap, error)
	List(ctx context.Context, userID string, status string, problemType, search string, startTime, endTime time.Time, page, limit int) ([]*model.ThinkingMap, int64, error)
}

// ThinkingNode 节点仓储接口
type ThinkingNode interface {
	Create(ctx context.Context, node *model.ThinkingNode) error
	Update(ctx context.Context, node *model.ThinkingNode) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*model.ThinkingNode, error)
	FindByIDs(ctx context.Context, ids []string) ([]*model.ThinkingNode, error)
	FindByMapID(ctx context.Context, mapID string) ([]*model.ThinkingNode, error)
	FindByParentID(ctx context.Context, parentID string) ([]*model.ThinkingNode, error)
	UpdatePosition(ctx context.Context, id string, position model.JSONB) error
	UpdateStatus(ctx context.Context, id string, status int) error
}

// Message 消息仓储接口
type Message interface {
	Create(ctx context.Context, message *model.Message) error
	Update(ctx context.Context, message *model.Message) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*model.Message, error)
	FindByParentID(ctx context.Context, parentID string) ([]*model.Message, error)
	FindByConversationID(ctx context.Context, conversationID string) ([]*model.Message, error)
	List(ctx context.Context, offset, limit int) ([]*model.Message, int64, error)
}

// RAGRecord RAG检索记录仓储接口
type RAGRecord interface {
	Create(ctx context.Context, record *model.RAGRecord) error
	Update(ctx context.Context, record *model.RAGRecord) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*model.RAGRecord, error)
	FindByNodeID(ctx context.Context, nodeID string, offset, limit int) ([]*model.RAGRecord, int64, error)
}
