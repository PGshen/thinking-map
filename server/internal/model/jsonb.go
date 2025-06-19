/*
 * @Date: 2025-06-18 23:05:02
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-06-19 00:29:32
 * @FilePath: /thinking-map/server/internal/model/jsonb.go
 */
package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// JSONB 自定义 JSONB 类型
type JSONB map[string]interface{}

// Value 实现 driver.Valuer 接口
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan 实现 sql.Scanner 接口
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSONB)
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("failed to unmarshal JSONB value: %v", value)
	}

	return json.Unmarshal(bytes, j)
}

// MarshalJSON 实现 json.Marshaler 接口
func (j JSONB) MarshalJSON() ([]byte, error) {
	if j == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(map[string]interface{}(j))
}

// UnmarshalJSON 实现 json.Unmarshaler 接口
func (j *JSONB) UnmarshalJSON(data []byte) error {
	if j == nil {
		return fmt.Errorf("cannot unmarshal into nil JSONB pointer")
	}
	*j = make(JSONB)
	return json.Unmarshal(data, j)
}
