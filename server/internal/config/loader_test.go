package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseEnvVar(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		envVars  map[string]string
		expected string
	}{
		{
			name:     "普通字符串，不包含环境变量引用",
			input:    "localhost",
			expected: "localhost",
		},
		{
			name:     "环境变量存在",
			input:    "${DB_HOST}",
			envVars:  map[string]string{"DB_HOST": "postgres.example.com"},
			expected: "postgres.example.com",
		},
		{
			name:     "环境变量不存在，返回原始值",
			input:    "${DB_HOST}",
			expected: "${DB_HOST}",
		},
		{
			name:     "带默认值，环境变量存在",
			input:    "${DB_HOST:-localhost}",
			envVars:  map[string]string{"DB_HOST": "postgres.example.com"},
			expected: "postgres.example.com",
		},
		{
			name:     "带默认值，环境变量不存在",
			input:    "${DB_HOST:-localhost}",
			expected: "localhost",
		},
		{
			name:     "带默认值，环境变量为空",
			input:    "${DB_HOST:-localhost}",
			envVars:  map[string]string{"DB_HOST": ""},
			expected: "localhost",
		},
		{
			name:     "带默认值，默认值包含空格",
			input:    "${DB_HOST:-localhost:5432}",
			expected: "localhost:5432",
		},
		{
			name:     "带默认值，环境变量名包含空格",
			input:    "${DB_HOST:-localhost}",
			envVars:  map[string]string{"DB_HOST": "postgres.example.com"},
			expected: "postgres.example.com",
		},
		{
			name:     "多个冒号，只分割第一个",
			input:    "${DB_HOST:-localhost:5432}",
			envVars:  map[string]string{"DB_HOST": "postgres.example.com:5432"},
			expected: "postgres.example.com:5432",
		},
		{
			name:     "默认值包含特殊字符",
			input:    "${DB_PASSWORD:-my@password}",
			expected: "my@password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 设置环境变量
			for key, value := range tt.envVars {
				os.Setenv(key, value)
				defer os.Unsetenv(key)
			}

			result := parseEnvVar(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLoadConfigWithEnvVars(t *testing.T) {
	// 创建临时配置文件
	tempConfig := `
server:
  port: 8080
  mode: debug

database:
  driver: postgres
  host: ${DB_HOST:-localhost}
  port: ${DB_PORT:-5432}
  username: ${DB_USER:-postgres}
  password: ${DB_PASSWORD:-postgres}
  dbname: ${DB_NAME:-thinking_map}
  sslmode: disable

redis:
  addr: ${REDIS_ADDR:-localhost:6379}
  password: ${REDIS_PASSWORD:-}

jwt:
  secret: ${JWT_SECRET:-your-secret-key}
  expire: ${JWT_EXPIRE:-24h}
`

	// 写入临时文件
	tmpFile, err := os.CreateTemp("", "test_config_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(tempConfig); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	// 设置环境变量
	os.Setenv("DB_HOST", "test.example.com")
	os.Setenv("DB_PASSWORD", "testpassword")
	os.Setenv("JWT_SECRET", "test-secret")
	defer func() {
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("JWT_SECRET")
	}()

	// 加载配置
	cfg, err := LoadConfig(tmpFile.Name())
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	// 验证环境变量被正确解析
	assert.Equal(t, "test.example.com", cfg.Database.Host)
	assert.Equal(t, "testpassword", cfg.Database.Password)
	assert.Equal(t, "test-secret", cfg.JWT.Secret)

	// 验证默认值被正确使用
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, "postgres", cfg.Database.Username)
	assert.Equal(t, "thinking_map", cfg.Database.DBName)
	assert.Equal(t, "localhost:6379", cfg.Redis.Addr)
	assert.Equal(t, "24h", cfg.JWT.Expire)
}
