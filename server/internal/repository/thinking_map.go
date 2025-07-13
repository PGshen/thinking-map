package repository

import (
	"context"
	"time"

	"github.com/PGshen/thinking-map/server/internal/model"

	"gorm.io/gorm"
)

type thinkingMapRepository struct {
	db *gorm.DB
}

// NewThinkingMapRepository 创建思维导图仓储实例
func NewThinkingMapRepository(db *gorm.DB) ThinkingMap {
	return &thinkingMapRepository{db: db}
}

// CreateMap creates a new thinking map and its root node
func (r *thinkingMapRepository) Create(ctx context.Context, thinkingMap *model.ThinkingMap, rootNode *model.ThinkingNode) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	if err := tx.Create(thinkingMap).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Create(rootNode).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// ListMaps retrieves a list of thinking maps with pagination
func (r *thinkingMapRepository) List(ctx context.Context, userID string, status string, problemType, search string, startTime, endTime time.Time, page, limit int) ([]*model.ThinkingMap, int64, error) {
	var maps []*model.ThinkingMap
	var total int64

	dbQuery := r.db.Model(&model.ThinkingMap{}).Where("user_id = ?", userID)
	if status != "" {
		dbQuery = dbQuery.Where("status = ?", status)
	}
	if problemType != "" {
		dbQuery = dbQuery.Where("problem_type = ?", problemType)
	}
	if search != "" {
		like := "%" + search + "%"
		dbQuery = dbQuery.Where("problem LIKE ? OR target LIKE ?", like, like)
	}

	if !startTime.IsZero() && !endTime.IsZero() {
		dbQuery = dbQuery.Where("created_at BETWEEN ? AND ?", startTime, endTime)
	}

	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * limit
	if err := dbQuery.Offset(offset).Limit(limit).Find(&maps).Error; err != nil {
		return nil, 0, err
	}

	return maps, total, nil
}

// GetMap retrieves a specific thinking map
func (r *thinkingMapRepository) FindByID(ctx context.Context, mapID string) (*model.ThinkingMap, error) {
	var thinkingMap model.ThinkingMap
	if err := r.db.Where(whereID, mapID).First(&thinkingMap).Error; err != nil {
		return nil, err
	}
	return &thinkingMap, nil
}

// GetRootNode retrieves the root node of a thinking map
func (r *thinkingMapRepository) GetRootNode(ctx context.Context, mapID string) (*model.ThinkingNode, error) {
	var rootNode model.ThinkingNode
	if err := r.db.Where("map_id = ? AND parent_id IS NULL", mapID).First(&rootNode).Error; err != nil {
		return nil, err
	}
	return &rootNode, nil
}

// GetNodeCount retrieves the number of nodes in a thinking map
func (r *thinkingMapRepository) GetNodeCount(ctx context.Context, mapID string) (int64, error) {
	var count int64
	if err := r.db.Model(&model.ThinkingNode{}).Where("map_id = ?", mapID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// UpdateMap updates a thinking map
func (r *thinkingMapRepository) Update(ctx context.Context, mapID string, updates map[string]interface{}) error {
	return r.db.Model(&model.ThinkingMap{}).
		Where("id = ?", mapID).
		Updates(updates).Error
}

// DeleteMap deletes a thinking map and all its nodes
func (r *thinkingMapRepository) Delete(ctx context.Context, mapID string) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	if err := tx.Where("map_id = ?", mapID).Delete(&model.ThinkingNode{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Where("id = ?", mapID).Delete(&model.ThinkingMap{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
