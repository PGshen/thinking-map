package repository

import (
	"context"

	"github.com/PGshen/thinking-map/server/internal/model"

	"gorm.io/gorm"
)

type ragRecordRepository struct {
	db *gorm.DB
}

// NewRAGRecordRepository 创建RAG记录仓储实例
func NewRAGRecordRepository(db *gorm.DB) RAGRecord {
	return &ragRecordRepository{db: db}
}

func (r *ragRecordRepository) Create(ctx context.Context, record *model.RAGRecord) error {
	return r.db.WithContext(ctx).Create(record).Error
}

func (r *ragRecordRepository) Update(ctx context.Context, record *model.RAGRecord) error {
	return r.db.WithContext(ctx).Save(record).Error
}

func (r *ragRecordRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.RAGRecord{}, id).Error
}

func (r *ragRecordRepository) FindByID(ctx context.Context, id string) (*model.RAGRecord, error) {
	var record model.RAGRecord
	err := r.db.WithContext(ctx).First(&record, id).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *ragRecordRepository) FindByNodeID(ctx context.Context, nodeID string, offset, limit int) ([]*model.RAGRecord, int64, error) {
	var records []*model.RAGRecord
	var total int64

	err := r.db.WithContext(ctx).Model(&model.RAGRecord{}).Where("node_id = ?", nodeID).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.WithContext(ctx).Where("node_id = ?", nodeID).Offset(offset).Limit(limit).Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

func (r *ragRecordRepository) List(ctx context.Context, offset, limit int) ([]*model.RAGRecord, int64, error) {
	var records []*model.RAGRecord
	var total int64

	err := r.db.WithContext(ctx).Model(&model.RAGRecord{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, nil
}
