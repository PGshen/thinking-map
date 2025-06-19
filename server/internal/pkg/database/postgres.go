/*
 * @Date: 2025-06-18 22:17:28
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-06-18 22:47:08
 * @FilePath: /thinking-map/server/internal/pkg/database/postgres.go
 */
package database

import (
	"fmt"

	"github.com/thinking-map/server/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewPostgresDB(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.DBName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, nil
}
