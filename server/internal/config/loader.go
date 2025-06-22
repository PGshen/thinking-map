/*
 * @Date: 2025-06-18 22:17:04
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-06-22 10:41:26
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

// parseEnvVar 解析环境变量引用，支持默认值语法 ${VAR:-default}
func parseEnvVar(value string) string {
	if !strings.HasPrefix(value, "${") || !strings.HasSuffix(value, "}") {
		return value
	}

	// 提取 ${} 内的内容
	content := strings.TrimSuffix(strings.TrimPrefix(value, "${"), "}")

	// 检查是否包含默认值语法 :-
	if strings.Contains(content, ":-") {
		parts := strings.SplitN(content, ":-", 2)
		if len(parts) == 2 {
			envVar := strings.TrimSpace(parts[0])
			defaultValue := strings.TrimSpace(parts[1])

			// 获取环境变量的值
			envValue := os.Getenv(envVar)
			if envValue != "" {
				return envValue
			}
			// 如果环境变量为空，返回默认值
			return defaultValue
		}
	}

	// 没有默认值语法，直接获取环境变量
	envValue := os.Getenv(content)
	if envValue != "" {
		return envValue
	}

	// 环境变量不存在，返回原始值
	return value
}

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
			// 解析环境变量引用，支持默认值语法
			resolvedValue := parseEnvVar(value)
			viper.Set(key, resolvedValue)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}
