/*
 * @Date: 2025-06-18 22:17:20
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-06-23 00:02:14
 * @FilePath: /thinking-map/server/internal/model/user.go
 */
package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents the user model
type User struct {
	ID        string         `gorm:"type:uuid;primary_key" json:"id"`
	Username  string         `gorm:"type:varchar(32);uniqueIndex;not null" json:"username"`
	Email     string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Password  string         `gorm:"type:varchar(255);not null" json:"-"`
	FullName  string         `gorm:"type:varchar(100);not null" json:"full_name"`
	Status    int            `gorm:"type:smallint;default:1;not null" json:"status"`
	CreatedAt time.Time      `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time      `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for User model
func (User) TableName() string {
	return "users"
}

// BeforeCreate 在创建记录前生成 UUID
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	return nil
}

// TokenInfo represents the token information stored in Redis
type TokenInfo struct {
	UserID      string    `json:"user_id"`
	Username    string    `json:"username"`
	AccessToken string    `json:"access_token"`
	ExpiresAt   time.Time `json:"expires_at"`
}
