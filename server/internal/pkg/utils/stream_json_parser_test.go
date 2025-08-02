package utils

import (
	"testing"
	"fmt"
)

func TestSimplePathMatcher(t *testing.T) {
	matcher := NewSimplePathMatcher()
	
	// 测试路径解析
	tests := []struct {
		pattern  string
		expected []interface{}
	}{
		{"$", []interface{}{"$"}},
		{"users[*].name", []interface{}{"users", "*", "name"}},
		{"data.items[0].value", []interface{}{"data", "items", 0, "value"}},
		{"response.content", []interface{}{"response", "content"}},
	}
	
	for _, test := range tests {
		result := matcher.parsePath(test.pattern)
		if len(result) != len(test.expected) {
			t.Errorf("Pattern %s: expected length %d, got %d", test.pattern, len(test.expected), len(result))
			continue
		}
		
		for i, expected := range test.expected {
			if result[i] != expected {
				t.Errorf("Pattern %s: expected %v at index %d, got %v", test.pattern, expected, i, result[i])
			}
		}
	}
}

func TestStreamingJsonParser(t *testing.T) {
	matcher := NewSimplePathMatcher()
	results := make(map[string]interface{})
	
	// 注册路径匹配回调
	matcher.On("name", func(value interface{}, path []interface{}) {
		results["name"] = value
		fmt.Printf("Found name: %v at path: %v\n", value, path)
	})
	
	matcher.On("users[*].email", func(value interface{}, path []interface{}) {
		results["email"] = value
		fmt.Printf("Found email: %v at path: %v\n", value, path)
	})
	
	// 添加通用回调来查看所有路径
	matcher.On("*", func(value interface{}, path []interface{}) {
		fmt.Printf("DEBUG: All paths - value: %v, path: %v\n", value, path)
	})
	
	parser := NewStreamingJsonParser(matcher, false)
	
	// 测试简单JSON
	jsonData := `{"name": "John", "age": 30, "users": [{"email": "john@example.com"}]}`
	
	err := parser.Write(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}
	
	err = parser.End()
	if err != nil {
		t.Fatalf("Failed to end parsing: %v", err)
	}
	
	// 验证结果
	if results["name"] != "John" {
		t.Errorf("Expected name to be 'John', got %v", results["name"])
	}
	
	if results["email"] != "john@example.com" {
		t.Errorf("Expected email to be 'john@example.com', got %v", results["email"])
	}
}

func TestStreamingJsonParserRealtime(t *testing.T) {
	matcher := NewSimplePathMatcher()
	callbackCount := 0
	
	// 注册实时回调
	matcher.On("message", func(value interface{}, path []interface{}) {
		callbackCount++
		fmt.Printf("Realtime callback %d: %v\n", callbackCount, value)
	})
	
	parser := NewStreamingJsonParser(matcher, true)
	
	// 模拟流式输入
	chunks := []string{
		`{"message": "Hel`,
		`lo Wor`,
		`ld"}`}
	
	for _, chunk := range chunks {
		err := parser.Write(chunk)
		if err != nil {
			t.Fatalf("Failed to parse chunk '%s': %v", chunk, err)
		}
	}
	
	err := parser.End()
	if err != nil {
		t.Fatalf("Failed to end parsing: %v", err)
	}
	
	// 在实时模式下，应该有多次回调
	if callbackCount == 0 {
		t.Error("Expected at least one callback in realtime mode")
	}
	
	fmt.Printf("Total callbacks in realtime mode: %d\n", callbackCount)
}

func TestComplexJsonStructure(t *testing.T) {
	matcher := NewSimplePathMatcher()
	results := make([]interface{}, 0)
	
	// 匹配数组中的所有项目
	matcher.On("data.items[*].value", func(value interface{}, path []interface{}) {
		results = append(results, value)
		fmt.Printf("Found item value: %v at path: %v\n", value, path)
	})
	
	parser := NewStreamingJsonParser(matcher, false)
	
	// 复杂的JSON结构
	jsonData := `{
		"data": {
			"items": [
				{"value": "first", "id": 1},
				{"value": "second", "id": 2},
				{"value": "third", "id": 3}
			]
		},
		"meta": {"count": 3}
	}`
	
	err := parser.Write(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse complex JSON: %v", err)
	}
	
	err = parser.End()
	if err != nil {
		t.Fatalf("Failed to end parsing: %v", err)
	}
	
	// 验证结果
	expectedValues := []string{"first", "second", "third"}
	if len(results) != len(expectedValues) {
		t.Errorf("Expected %d results, got %d", len(expectedValues), len(results))
	}
	
	for i, expected := range expectedValues {
		if i < len(results) && results[i] != expected {
			t.Errorf("Expected result[%d] to be '%s', got %v", i, expected, results[i])
		}
	}
}

func TestErrorHandling(t *testing.T) {
	matcher := NewSimplePathMatcher()
	parser := NewStreamingJsonParser(matcher, false)
	
	// 测试无效JSON
	invalidJson := `{"name": "John", "age":}`
	
	err := parser.Write(invalidJson)
	if err == nil {
		t.Error("Expected error for invalid JSON, but got none")
	}
}