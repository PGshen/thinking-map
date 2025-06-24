/*
 * @Date: 2025-06-18 22:16:57
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-06-20 23:36:41
 * @FilePath: /thinking-map/server/internal/config/config.go
 */
package config

import (
	"github.com/PGshen/thinking-map/server/internal/pkg/logger"
)

// Config 配置结构体
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	JWT      JWTConfig      `yaml:"jwt"`
	Log      logger.Config  `yaml:"log"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port int    `yaml:"port"`
	Mode string `yaml:"mode"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver   string `yaml:"driver"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret string `yaml:"secret"`
	Expire string `yaml:"expire"`
}

// RedisConfig Redis配置
// 可根据需要扩展更多配置项
// 例如：PoolSize、MinIdleConns等
type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
}
