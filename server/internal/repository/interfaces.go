/*
 * @Date: 2025-06-18 22:33:07
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-06-18 22:48:49
 * @FilePath: /thinking-map/server/internal/repository/interfaces.go
 */
package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/thinking-map/server/internal/model"
)

// User 用户仓储接口
type User interface {
	Create(ctx context.Context, user *model.User) error
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	List(ctx context.Context, offset, limit int) ([]*model.User, int64, error)
}

// ThinkingMap 思维导图仓储接口
type ThinkingMap interface {
	Create(ctx context.Context, map_ *model.ThinkingMap) error
	Update(ctx context.Context, map_ *model.ThinkingMap) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.ThinkingMap, error)
	FindByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*model.ThinkingMap, int64, error)
	List(ctx context.Context, offset, limit int) ([]*model.ThinkingMap, int64, error)
}

// ThinkingNode 节点仓储接口
type ThinkingNode interface {
	Create(ctx context.Context, node *model.ThinkingNode) error
	Update(ctx context.Context, node *model.ThinkingNode) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.ThinkingNode, error)
	FindByMapID(ctx context.Context, mapID uuid.UUID) ([]*model.ThinkingNode, error)
	FindByParentID(ctx context.Context, parentID uuid.UUID) ([]*model.ThinkingNode, error)
	UpdatePosition(ctx context.Context, id uuid.UUID, position model.JSONB) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status int) error
}

// NodeDetail 节点详情仓储接口
type NodeDetail interface {
	Create(ctx context.Context, detail *model.NodeDetail) error
	Update(ctx context.Context, detail *model.NodeDetail) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.NodeDetail, error)
	FindByNodeID(ctx context.Context, nodeID uuid.UUID) ([]*model.NodeDetail, error)
	FindByNodeIDAndType(ctx context.Context, nodeID uuid.UUID, tabType string) (*model.NodeDetail, error)
	UpdateContent(ctx context.Context, id uuid.UUID, content model.JSONB) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status int) error
}

// Message 消息仓储接口
type Message interface {
	Create(ctx context.Context, message *model.Message) error
	Update(ctx context.Context, message *model.Message) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Message, error)
	FindByNodeID(ctx context.Context, nodeID uuid.UUID, offset, limit int) ([]*model.Message, int64, error)
	FindByParentID(ctx context.Context, parentID uuid.UUID) ([]*model.Message, error)
	UpdateVersion(ctx context.Context, id uuid.UUID, version int) error
}

// RAGRecord RAG检索记录仓储接口
type RAGRecord interface {
	Create(ctx context.Context, record *model.RAGRecord) error
	Update(ctx context.Context, record *model.RAGRecord) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.RAGRecord, error)
	FindByNodeID(ctx context.Context, nodeID uuid.UUID, offset, limit int) ([]*model.RAGRecord, int64, error)
}
