package repository

import (
	"github.com/google/uuid"
	"github.com/thinking-map/server/internal/model"
	"gorm.io/gorm"
)

type MapRepository struct {
	db *gorm.DB
}

func NewMapRepository(db *gorm.DB) *MapRepository {
	return &MapRepository{
		db: db,
	}
}

// CreateMap creates a new thinking map and its root node
func (r *MapRepository) CreateMap(thinkingMap *model.ThinkingMap, rootNode *model.ThinkingNode) error {
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
func (r *MapRepository) ListMaps(userID uuid.UUID, status int, page, limit int) ([]model.ThinkingMap, int64, error) {
	var maps []model.ThinkingMap
	var total int64

	dbQuery := r.db.Model(&model.ThinkingMap{}).Where("created_by = ?", userID)
	if status > 0 {
		dbQuery = dbQuery.Where("status = ?", status)
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
func (r *MapRepository) GetMap(mapID, userID uuid.UUID) (*model.ThinkingMap, error) {
	var thinkingMap model.ThinkingMap
	if err := r.db.Where("id = ? AND created_by = ?", mapID, userID).First(&thinkingMap).Error; err != nil {
		return nil, err
	}
	return &thinkingMap, nil
}

// GetRootNode retrieves the root node of a thinking map
func (r *MapRepository) GetRootNode(mapID uuid.UUID) (*model.ThinkingNode, error) {
	var rootNode model.ThinkingNode
	if err := r.db.Where("map_id = ? AND parent_id IS NULL", mapID).First(&rootNode).Error; err != nil {
		return nil, err
	}
	return &rootNode, nil
}

// GetNodeCount retrieves the number of nodes in a thinking map
func (r *MapRepository) GetNodeCount(mapID uuid.UUID) (int64, error) {
	var count int64
	if err := r.db.Model(&model.ThinkingNode{}).Where("map_id = ?", mapID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// UpdateMap updates a thinking map
func (r *MapRepository) UpdateMap(mapID, userID uuid.UUID, updates map[string]interface{}) error {
	return r.db.Model(&model.ThinkingMap{}).
		Where("id = ? AND created_by = ?", mapID, userID).
		Updates(updates).Error
}

// DeleteMap deletes a thinking map and all its nodes
func (r *MapRepository) DeleteMap(mapID, userID uuid.UUID) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	if err := tx.Where("map_id = ?", mapID).Delete(&model.ThinkingNode{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Where("id = ? AND created_by = ?", mapID, userID).Delete(&model.ThinkingMap{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
