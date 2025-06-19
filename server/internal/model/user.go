/*
 * @Date: 2025-06-18 22:17:20
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-06-19 00:19:26
 * @FilePath: /thinking-map/server/internal/model/user.go
 */
package model

import (
	"time"

	"gorm.io/gorm"
)

// User represents the user model
type User struct {
	ID        string         `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	Username  string         `gorm:"type:varchar(32);uniqueIndex;not null" json:"username"`
	Email     string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Password  string         `gorm:"type:varchar(255);not null" json:"-"`
	FullName  string         `gorm:"type:varchar(100);not null" json:"full_name"`
	Status    int            `gorm:"type:smallint;default:1;not null" json:"status"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for User model
func (User) TableName() string {
	return "users"
}

// TokenInfo represents the token information stored in Redis
type TokenInfo struct {
	UserID      string    `json:"user_id"`
	Username    string    `json:"username"`
	AccessToken string    `json:"access_token"`
	ExpiresAt   time.Time `json:"expires_at"`
}
