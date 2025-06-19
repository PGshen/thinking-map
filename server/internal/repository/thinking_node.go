/*
 * @Date: 2025-06-18 22:34:15
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-06-18 22:52:54
 * @FilePath: /thinking-map/server/internal/repository/thinking_node.go
 */
package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/thinking-map/server/internal/model"
	"gorm.io/gorm"
)

type thinkingNodeRepository struct {
	db *gorm.DB
}

// NewThinkingNodeRepository 创建思维节点仓储实例
func NewThinkingNodeRepository(db *gorm.DB) ThinkingNode {
	return &thinkingNodeRepository{db: db}
}

func (r *thinkingNodeRepository) Create(ctx context.Context, node *model.ThinkingNode) error {
	return r.db.WithContext(ctx).Create(node).Error
}

func (r *thinkingNodeRepository) Update(ctx context.Context, node *model.ThinkingNode) error {
	return r.db.WithContext(ctx).Save(node).Error
}

func (r *thinkingNodeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.ThinkingNode{}, id).Error
}

func (r *thinkingNodeRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.ThinkingNode, error) {
	var node model.ThinkingNode
	err := r.db.WithContext(ctx).First(&node, id).Error
	if err != nil {
		return nil, err
	}
	return &node, nil
}

func (r *thinkingNodeRepository) FindByMapID(ctx context.Context, mapID uuid.UUID) ([]*model.ThinkingNode, error) {
	var nodes []*model.ThinkingNode
	err := r.db.WithContext(ctx).Where("map_id = ?", mapID).Find(&nodes).Error
	if err != nil {
		return nil, err
	}
	return nodes, nil
}

func (r *thinkingNodeRepository) FindByParentID(ctx context.Context, parentID uuid.UUID) ([]*model.ThinkingNode, error) {
	var nodes []*model.ThinkingNode
	err := r.db.WithContext(ctx).Where("parent_id = ?", parentID).Find(&nodes).Error
	if err != nil {
		return nil, err
	}
	return nodes, nil
}

func (r *thinkingNodeRepository) List(ctx context.Context, offset, limit int) ([]*model.ThinkingNode, int64, error) {
	var nodes []*model.ThinkingNode
	var total int64

	err := r.db.WithContext(ctx).Model(&model.ThinkingNode{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&nodes).Error
	if err != nil {
		return nil, 0, err
	}

	return nodes, total, nil
}

// UpdatePosition 更新节点位置
func (r *thinkingNodeRepository) UpdatePosition(ctx context.Context, id uuid.UUID, position model.JSONB) error {
	return r.db.WithContext(ctx).Model(&model.ThinkingNode{}).
		Where("id = ?", id).
		Update("position", position).Error
}

// UpdateStatus 更新节点状态
func (r *thinkingNodeRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status int) error {
	return r.db.WithContext(ctx).Model(&model.ThinkingNode{}).
		Where("id = ?", id).
		Update("status", status).Error
}
