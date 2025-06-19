package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/thinking-map/server/internal/model"
	"gorm.io/gorm"
)

type thinkingMapRepository struct {
	db *gorm.DB
}

// NewThinkingMapRepository 创建思维导图仓储实例
func NewThinkingMapRepository(db *gorm.DB) ThinkingMap {
	return &thinkingMapRepository{db: db}
}

func (r *thinkingMapRepository) Create(ctx context.Context, map_ *model.ThinkingMap) error {
	return r.db.WithContext(ctx).Create(map_).Error
}

func (r *thinkingMapRepository) Update(ctx context.Context, map_ *model.ThinkingMap) error {
	return r.db.WithContext(ctx).Save(map_).Error
}

func (r *thinkingMapRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.ThinkingMap{}, id).Error
}

func (r *thinkingMapRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.ThinkingMap, error) {
	var map_ model.ThinkingMap
	err := r.db.WithContext(ctx).First(&map_, id).Error
	if err != nil {
		return nil, err
	}
	return &map_, nil
}

func (r *thinkingMapRepository) FindByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*model.ThinkingMap, int64, error) {
	var maps []*model.ThinkingMap
	var total int64

	err := r.db.WithContext(ctx).Model(&model.ThinkingMap{}).Where("user_id = ?", userID).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.WithContext(ctx).Where("user_id = ?", userID).Offset(offset).Limit(limit).Find(&maps).Error
	if err != nil {
		return nil, 0, err
	}

	return maps, total, nil
}

func (r *thinkingMapRepository) List(ctx context.Context, offset, limit int) ([]*model.ThinkingMap, int64, error) {
	var maps []*model.ThinkingMap
	var total int64

	err := r.db.WithContext(ctx).Model(&model.ThinkingMap{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&maps).Error
	if err != nil {
		return nil, 0, err
	}

	return maps, total, nil
}
