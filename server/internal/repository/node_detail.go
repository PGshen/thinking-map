/*
 * @Date: 2025-06-18 22:34:33
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-06-18 22:54:08
 * @FilePath: /thinking-map/server/internal/repository/node_detail.go
 */
package repository

import (
	"context"

	"github.com/PGshen/thinking-map/server/internal/model"

	"gorm.io/gorm"
)

type nodeDetailRepository struct {
	db *gorm.DB
}

// NewNodeDetailRepository 创建节点详情仓储实例
func NewNodeDetailRepository(db *gorm.DB) NodeDetail {
	return &nodeDetailRepository{db: db}
}

func (r *nodeDetailRepository) Create(ctx context.Context, detail *model.NodeDetail) error {
	return r.db.WithContext(ctx).Create(detail).Error
}

func (r *nodeDetailRepository) Update(ctx context.Context, detail *model.NodeDetail) error {
	return r.db.WithContext(ctx).Save(detail).Error
}

func (r *nodeDetailRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where(whereID, id).Delete(&model.NodeDetail{}).Error
}

func (r *nodeDetailRepository) FindByID(ctx context.Context, id string) (*model.NodeDetail, error) {
	var detail model.NodeDetail
	err := r.db.WithContext(ctx).Where(whereID, id).First(&detail).Error
	if err != nil {
		return nil, err
	}
	return &detail, nil
}

func (r *nodeDetailRepository) FindByNodeID(ctx context.Context, nodeID string) ([]*model.NodeDetail, error) {
	var details []*model.NodeDetail
	err := r.db.WithContext(ctx).Where("node_id = ?", nodeID).Find(&details).Error
	if err != nil {
		return nil, err
	}
	return details, nil
}

func (r *nodeDetailRepository) FindByNodeIDAndTabType(ctx context.Context, nodeID string, tabType string) (*model.NodeDetail, error) {
	var detail model.NodeDetail
	err := r.db.WithContext(ctx).Where("node_id = ? AND tab_type = ?", nodeID, tabType).First(&detail).Error
	if err != nil {
		return nil, err
	}
	return &detail, nil
}

func (r *nodeDetailRepository) FindByNodeIDAndType(ctx context.Context, nodeID string, detailType string) (*model.NodeDetail, error) {
	var detail model.NodeDetail
	err := r.db.WithContext(ctx).Where("node_id = ? AND detail_type = ?", nodeID, detailType).First(&detail).Error
	if err != nil {
		return nil, err
	}
	return &detail, nil
}

func (r *nodeDetailRepository) List(ctx context.Context, offset, limit int) ([]*model.NodeDetail, int64, error) {
	var details []*model.NodeDetail
	var total int64

	err := r.db.WithContext(ctx).Model(&model.NodeDetail{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&details).Error
	if err != nil {
		return nil, 0, err
	}

	return details, total, nil
}

// UpdateContent 更新节点详情内容
func (r *nodeDetailRepository) UpdateContent(ctx context.Context, id string, content model.JSONB) error {
	return r.db.WithContext(ctx).Model(&model.NodeDetail{}).
		Where("id = ?", id).
		Update("content", content).Error
}

// UpdateStatus 更新节点详情状态
func (r *nodeDetailRepository) UpdateStatus(ctx context.Context, id string, status int) error {
	return r.db.WithContext(ctx).Model(&model.NodeDetail{}).
		Where("id = ?", id).
		Update("status", status).Error
}
