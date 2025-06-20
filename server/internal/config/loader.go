/*
 * @Date: 2025-06-18 22:17:04
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-06-20 23:28:12
 * @FilePath: /thinking-map/server/internal/config/loader.go
 */
package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

func LoadConfig(configPath string) (*Config, error) {
	// 先尝试加载.env文件（如果存在）
	_ = godotenv.Load()

	viper.SetConfigType("yaml")
	viper.SetConfigFile(configPath)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	// 处理配置文件中的环境变量引用
	for _, key := range viper.AllKeys() {
		value := viper.GetString(key)
		if strings.HasPrefix(value, "${") && strings.HasSuffix(value, "}") {
			// 提取环境变量名称
			envVar := strings.TrimSuffix(strings.TrimPrefix(value, "${"), "}")
			// 获取环境变量的值
			envValue := os.Getenv(envVar)
			if envValue != "" {
				// 设置实际值
				viper.Set(key, envValue)
			}
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}
