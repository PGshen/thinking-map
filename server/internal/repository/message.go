/*
 * @Date: 2025-06-18 22:34:47
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-06-19 00:41:16
 * @FilePath: /thinking-map/server/internal/repository/message.go
 */
package repository

import (
	"context"

	"github.com/PGshen/thinking-map/server/internal/model"

	"gorm.io/gorm"
)

type messageRepository struct {
	db *gorm.DB
}

// NewMessageRepository 创建消息仓储实例
func NewMessageRepository(db *gorm.DB) Message {
	return &messageRepository{db: db}
}

func (r *messageRepository) Create(ctx context.Context, message *model.Message) error {
	return r.db.WithContext(ctx).Create(message).Error
}

func (r *messageRepository) Update(ctx context.Context, message *model.Message) error {
	return r.db.WithContext(ctx).Save(message).Error
}

func (r *messageRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.Message{}, id).Error
}

func (r *messageRepository) FindByID(ctx context.Context, id string) (*model.Message, error) {
	var message model.Message
	err := r.db.WithContext(ctx).First(&message, id).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}

func (r *messageRepository) FindByNodeID(ctx context.Context, nodeID string, offset, limit int) ([]*model.Message, int64, error) {
	var messages []*model.Message
	var total int64

	err := r.db.WithContext(ctx).Model(&model.Message{}).Where("node_id = ?", nodeID).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.WithContext(ctx).Where("node_id = ?", nodeID).Offset(offset).Limit(limit).Find(&messages).Error
	if err != nil {
		return nil, 0, err
	}

	return messages, total, nil
}

func (r *messageRepository) FindByNodeIDAndType(ctx context.Context, nodeID string, messageType string) ([]*model.Message, error) {
	var messages []*model.Message
	err := r.db.WithContext(ctx).Where("node_id = ? AND message_type = ?", nodeID, messageType).Find(&messages).Error
	if err != nil {
		return nil, err
	}
	return messages, nil
}

func (r *messageRepository) FindByParentID(ctx context.Context, parentID string) ([]*model.Message, error) {
	var messages []*model.Message
	err := r.db.WithContext(ctx).Where("parent_id = ?", parentID).Find(&messages).Error
	if err != nil {
		return nil, err
	}
	return messages, nil
}

func (r *messageRepository) UpdateVersion(ctx context.Context, id string, version int) error {
	return r.db.WithContext(ctx).Model(&model.Message{}).Where("id = ?", id).Update("version", version).Error
}

func (r *messageRepository) List(ctx context.Context, offset, limit int) ([]*model.Message, int64, error) {
	var messages []*model.Message
	var total int64

	err := r.db.WithContext(ctx).Model(&model.Message{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&messages).Error
	if err != nil {
		return nil, 0, err
	}

	return messages, total, nil
}
