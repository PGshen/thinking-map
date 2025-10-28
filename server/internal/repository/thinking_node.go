/*
 * @Date: 2025-06-18 22:34:15
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-06-18 22:52:54
 * @FilePath: /thinking-map/server/internal/repository/thinking_node.go
 */
package repository

import (
	"context"

	"github.com/PGshen/thinking-map/server/internal/model"
	"github.com/PGshen/thinking-map/server/internal/pkg/comm"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

// UpdateInTx 在事务中更新节点
func (r *thinkingNodeRepository) UpdateInTx(ctx context.Context, tx *gorm.DB, node *model.ThinkingNode) error {
	return tx.WithContext(ctx).Save(node).Error
}

func (r *thinkingNodeRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where(whereID, id).Delete(&model.ThinkingNode{}).Error
}

func (r *thinkingNodeRepository) FindByID(ctx context.Context, id string) (*model.ThinkingNode, error) {
	var node model.ThinkingNode
	err := r.db.WithContext(ctx).Where(whereID, id).First(&node).Error
	if err != nil {
		return nil, err
	}
	return &node, nil
}

// FindByIDForUpdate 使用行级锁查找节点
func (r *thinkingNodeRepository) FindByIDForUpdate(ctx context.Context, tx *gorm.DB, id string) (*model.ThinkingNode, error) {
	var node model.ThinkingNode
	err := tx.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where(whereID, id).First(&node).Error
	if err != nil {
		return nil, err
	}
	return &node, nil
}

// FindByIDs retrieves multiple ThinkingNode records by their IDs
func (r *thinkingNodeRepository) FindByIDs(ctx context.Context, ids []string) ([]*model.ThinkingNode, error) {
	var nodes []*model.ThinkingNode
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&nodes).Error
	if err != nil {
		return nil, err
	}
	return nodes, nil
}

func (r *thinkingNodeRepository) FindByMapID(ctx context.Context, mapID string) ([]*model.ThinkingNode, error) {
	var nodes []*model.ThinkingNode
	err := r.db.WithContext(ctx).Where("map_id = ?", mapID).Find(&nodes).Error
	if err != nil {
		return nil, err
	}
	return nodes, nil
}

func (r *thinkingNodeRepository) FindByParentID(ctx context.Context, parentID string) ([]*model.ThinkingNode, error) {
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
func (r *thinkingNodeRepository) UpdatePosition(ctx context.Context, id string, position model.JSONB) error {
	return r.db.WithContext(ctx).Model(&model.ThinkingNode{}).
		Where("id = ?", id).
		Update("position", position).Error
}

// UpdateStatus 更新节点状态
func (r *thinkingNodeRepository) UpdateStatus(ctx context.Context, id string, status int) error {
	return r.db.WithContext(ctx).Model(&model.ThinkingNode{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *thinkingNodeRepository) UpdateIsDecomposed(ctx context.Context, id string, isDecomposed bool) error {
	// Update isDecomposed field in decomposition JSONB column
	return r.db.WithContext(ctx).Model(&model.ThinkingNode{}).
		Where("id = ?", id).
		UpdateColumn("status", comm.NodeStatusInDecomposition).
		UpdateColumn("decomposition", gorm.Expr("jsonb_set(decomposition, '{isDecomposed}', ?)", isDecomposed)).Error
}
