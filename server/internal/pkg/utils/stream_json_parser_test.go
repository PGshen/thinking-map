package utils

import (
	"fmt"
	"testing"
	"time"
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

	parser := NewStreamingJsonParser(matcher, false, false)

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

	parser := NewStreamingJsonParser(matcher, true, false)

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

	parser := NewStreamingJsonParser(matcher, false, false)

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
	parser := NewStreamingJsonParser(matcher, false, false)

	// 测试无效JSON
	invalidJson := `{"name": "test", "invalid": }`
	err := parser.Write(invalidJson)
	if err == nil {
		t.Error("Expected error for invalid JSON, but got none")
	}
}

func TestStreamingJsonParserIncremental(t *testing.T) {
	tests := []struct {
		name        string
		input       []string // 分块输入
		incremental bool
		expected    []string // 期望的回调结果
	}{
		{
			name:        "incremental mode - complete values",
			input:       []string{`{"thought":"First part"}`, `{"thought":"Second part"}`, `{"thought":"Third part"}`},
			incremental: true,
			expected:    []string{"First part", "Second part", "Third part"},
		},
		{
			name:        "cumulative mode - complete values",
			input:       []string{`{"thought":"First part"}`, `{"thought":"Second part"}`, `{"thought":"Third part"}`},
			incremental: false,
			expected:    []string{"First part", "Second part", "Third part"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var results []string
			matcher := NewSimplePathMatcher()
			matcher.On("thought", func(value interface{}, path []interface{}) {
				if str, ok := value.(string); ok {
					results = append(results, str)
				}
			})

			parser := NewStreamingJsonParser(matcher, false, tt.incremental) // 非实时模式

			for _, chunk := range tt.input {
				parser.Reset() // 每个JSON重置解析器
				if err := parser.Write(chunk); err != nil {
					t.Fatalf("Failed to write chunk: %v", err)
				}
				if err := parser.End(); err != nil {
					t.Fatalf("Failed to end parsing: %v", err)
				}
			}

			if len(results) != len(tt.expected) {
				t.Fatalf("Expected %d results, got %d: %v", len(tt.expected), len(results), results)
			}

			for i, expected := range tt.expected {
				if results[i] != expected {
					t.Errorf("Result %d: expected %q, got %q", i, expected, results[i])
				}
			}
		})
	}
}

func TestStreamingJsonParserRealtimeIncremental(t *testing.T) {
	// 测试实时模式下的增量解析
	var thoughtResults []string
	matcher := NewSimplePathMatcher()
	matcher.On("thought", func(value interface{}, path []interface{}) {
		if str, ok := value.(string); ok {
			thoughtResults = append(thoughtResults, str)
		}
	})

	// 测试增量模式
	parser := NewStreamingJsonParser(matcher, true, true)

	// 模拟流式输入 - 在字符串内部分块
	chunks := []string{
		`{"thought":"Let me`,
		` think about`,
		` this problem"}`}

	for _, chunk := range chunks {
		if err := parser.Write(chunk); err != nil {
			t.Fatalf("Failed to write chunk: %v", err)
		}
		time.Sleep(1 * time.Second)
	}

	// 在增量模式下，应该收到每个字符的增量更新
	// 但最后一个结果应该是完整的字符串
	if len(thoughtResults) == 0 {
		t.Fatal("Expected at least one result")
	}

	// 最后一个结果应该包含完整内容
	lastResult := thoughtResults[len(thoughtResults)-1]
	if lastResult != "Let me think about this problem" {
		t.Errorf("Expected final result to be complete string, got: %q", lastResult)
	}

	// 测试累积模式
	thoughtResults = nil // 重置结果
	parser2 := NewStreamingJsonParser(matcher, true, false)

	for _, chunk := range chunks {
		if err := parser2.Write(chunk); err != nil {
			t.Fatalf("Failed to write chunk: %v", err)
		}
	}

	if len(thoughtResults) == 0 {
		t.Fatal("Expected at least one result in cumulative mode")
	}

	// 在累积模式下，最后一个结果也应该是完整的字符串
	lastResult2 := thoughtResults[len(thoughtResults)-1]
	if lastResult2 != "Let me think about this problem" {
		t.Errorf("Expected final result to be complete string in cumulative mode, got: %q", lastResult2)
	}

	// 验证增量模式产生的结果数量应该少于或等于累积模式
	// （因为增量模式避免了重复发送相同内容）
	t.Logf("Incremental mode results: %d, Cumulative mode results: %d", len(thoughtResults), len(thoughtResults))
}
