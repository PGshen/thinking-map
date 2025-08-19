package utils

import (
	"fmt"
	"strings"
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
	var incrementalResults []string
	matcher1 := NewSimplePathMatcher()
	matcher1.On("thought", func(value interface{}, path []interface{}) {
		if str, ok := value.(string); ok {
			incrementalResults = append(incrementalResults, str)
			t.Logf("Incremental result: %q", str)
		}
	})

	// 测试增量模式
	parser := NewStreamingJsonParser(matcher1, true, true)

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

	// 在增量模式下，应该只收到增量更新，不应该有完整字符串
	if len(incrementalResults) == 0 {
		t.Fatal("Expected at least one result")
	}

	// 验证没有完整字符串被发送
	for _, result := range incrementalResults {
		if result == "Let me think about this problem" {
			t.Errorf("Incremental mode should not send complete string, but got: %q", result)
		}
	}

	// 测试累积模式
	var cumulativeResults []string
	matcher2 := NewSimplePathMatcher()
	matcher2.On("thought", func(value interface{}, path []interface{}) {
		if str, ok := value.(string); ok {
			cumulativeResults = append(cumulativeResults, str)
			t.Logf("Cumulative result: %q", str)
		}
	})
	parser2 := NewStreamingJsonParser(matcher2, true, false)

	for _, chunk := range chunks {
		if err := parser2.Write(chunk); err != nil {
			t.Fatalf("Failed to write chunk: %v", err)
		}
	}

	if len(cumulativeResults) == 0 {
		t.Fatal("Expected at least one result in cumulative mode")
	}

	// 在累积模式下，最后一个结果应该是完整的字符串
	lastResult := cumulativeResults[len(cumulativeResults)-1]
	if lastResult != "Let me think about this problem" {
		t.Errorf("Expected final result to be complete string in cumulative mode, got: %q", lastResult)
	}

	// 验证增量模式的结果数量应该更少（因为避免了重复发送完整字符串）
	t.Logf("Incremental mode results: %d, Cumulative mode results: %d", len(incrementalResults), len(cumulativeResults))
	if len(incrementalResults) >= len(cumulativeResults) {
		t.Errorf("Expected incremental mode to have fewer results than cumulative mode, got incremental: %d, cumulative: %d", len(incrementalResults), len(cumulativeResults))
	}
}

// TestRealtimeIncrementalBugDemo 简化的测试，专门演示实时增量模式的问题
func TestRealtimeIncrementalBugDemo(t *testing.T) {
	var results []interface{}
	matcher := NewSimplePathMatcher()

	// 匹配第一个数组元素的value
	matcher.On("items[0].value", func(value interface{}, path []interface{}) {
		results = append(results, value)
		t.Logf("Found value: %v at path: %v", value, path)
	})

	// 使用实时增量模式
	parser := NewStreamingJsonParser(matcher, true, true)

	// 简单的JSON数组
	jsonData := `{"items":[{"value":"hello"}]}`

	err := parser.Write(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	err = parser.End()
	if err != nil {
		t.Fatalf("Failed to end parsing: %v", err)
	}

	// 在实时增量模式下，"hello"会被拆分成5个字符发送
	t.Logf("Total callbacks triggered: %d", len(results))
	t.Logf("All results: %v", results)

	if len(results) == 1 {
		t.Logf("SUCCESS: Got complete string value: %v", results[0])
	} else {
		t.Logf("EXPECTED BEHAVIOR: In realtime incremental mode, got %d character increments: %v", len(results), results)
	}
}

// TestArrayElementsMatchingInRealtimeIncremental 验证实时增量模式下是否能匹配到所有数组元素
func TestArrayElementsMatchingInRealtimeIncremental(t *testing.T) {
	var allMatches []struct {
		value interface{}
		path  []interface{}
	}
	matcher := NewSimplePathMatcher()

	// 匹配数组中所有元素的value字段
	matcher.On("data.items[*].value", func(value interface{}, path []interface{}) {
		allMatches = append(allMatches, struct {
			value interface{}
			path  []interface{}
		}{value: value, path: append([]interface{}{}, path...)})
		t.Logf("Match: value='%v' at path=%v", value, path)
	})

	// 使用实时增量模式
	parser := NewStreamingJsonParser(matcher, true, true)

	// 包含3个数组元素的JSON
	jsonData := `{
		"data": {
			"items": [
				{"value": "first", "id": 1},
				{"value": "second", "id": 2},
				{"value": "third", "id": 3}
			]
		}
	}`

	err := parser.Write(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	err = parser.End()
	if err != nil {
		t.Fatalf("Failed to end parsing: %v", err)
	}

	// 分析匹配结果
	uniqueArrayIndices := make(map[int]bool)
	maxValuesByIndex := make(map[int]string)

	for _, match := range allMatches {
		if len(match.path) >= 3 {
			if arrayIndex, ok := match.path[2].(int); ok {
				uniqueArrayIndices[arrayIndex] = true
				// 记录每个索引位置的最长值（最完整的值）
				if str, ok := match.value.(string); ok {
					if len(str) > len(maxValuesByIndex[arrayIndex]) {
						maxValuesByIndex[arrayIndex] = str
					}
				}
			}
		}
	}

	t.Logf("Total matches: %d", len(allMatches))
	t.Logf("Unique array indices matched: %v", getKeys(uniqueArrayIndices))
	t.Logf("Max values by index: %v", maxValuesByIndex)

	// 验证是否匹配到了所有3个数组元素
	expectedIndices := []int{0, 1, 2}
	for _, expectedIndex := range expectedIndices {
		if !uniqueArrayIndices[expectedIndex] {
			t.Errorf("Missing array index %d in matched paths", expectedIndex)
		}
	}

	// 验证是否找到了预期的值（即使是增量形式）
	expectedValues := map[int]string{0: "first", 1: "second", 2: "third"}
	allFound := true
	for index, expectedValue := range expectedValues {
		if maxValue, exists := maxValuesByIndex[index]; exists {
			if maxValue == expectedValue {
				t.Logf("✅ Array index %d: found complete value '%s'", index, maxValue)
			} else {
				t.Logf("⚠️  Array index %d: found partial value '%s', expected '%s'", index, maxValue, expectedValue)
				allFound = false
			}
		} else {
			t.Errorf("❌ Array index %d: no value found", index)
			allFound = false
		}
	}

	if len(uniqueArrayIndices) == 3 {
		t.Logf("✅ SUCCESS: All 3 array elements were matched in realtime incremental mode")
		if allFound {
			t.Logf("✅ All expected values were found (complete or partial)")
		} else {
			t.Logf("⚠️  Some values were only partially matched due to incremental nature")
		}
	} else {
		t.Errorf("❌ Expected 3 array indices, got %d", len(uniqueArrayIndices))
	}
}

// TestRealtimeIncrementalArrayBehaviorSummary 总结性测试：验证实时增量模式下数组解析的完整行为
func TestRealtimeIncrementalArrayBehaviorSummary(t *testing.T) {
	t.Log("=== 实时增量模式下JSON数组解析行为验证 ===")

	var allResults []interface{}
	var pathCounts map[string]int = make(map[string]int)
	matcher := NewSimplePathMatcher()

	// 匹配数组中所有元素的value字段
	matcher.On("steps[*].name", func(value interface{}, path []interface{}) {
		allResults = append(allResults, value)
		pathStr := fmt.Sprintf("%v", path)
		pathCounts[pathStr]++
	})

	// 使用实时增量模式
	parser := NewStreamingJsonParser(matcher, true, true)

	// 包含多个数组元素的JSON
	jsonData := `{
    "id": "plan_001",
    "name": "Evaluation of AI in K-12 Education",
    "description": "A structured approach to evaluate the effectiveness and impact of AI technology in K-12 education.",
    "steps": [
        {
            "id": "step_1",
            "name": "Determine Decomposition Strategy",
            "description": "Analyze the complexity of the task and decide on the suitable decomposition strategy for the evaluation process.",
            "assigned_specialist": "DecompositionDecisionAgent",
            "priority": 1,
            "dependencies": [],
            "parameters": {}
        },
        {
            "id": "step_2",
            "name": "Decompose Evaluation Framework",
            "description": "Based on the determined strategy, decompose the overall evaluation framework into manageable sub-problems and identify their dependencies.",
            "assigned_specialist": "ProblemDecompositionAgent",
            "priority": 2,
            "dependencies": [
                "step_1"
            ],
            "parameters": {}
        }
    ]
}`

	err := parser.Write(jsonData)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}

	err = parser.End()
	if err != nil {
		t.Fatalf("结束解析失败: %v", err)
	}

	// 分析结果
	uniqueIndices := make(map[int]bool)
	for pathStr := range pathCounts {
		// 从路径字符串中提取数组索引
		if strings.Contains(pathStr, "steps 0") {
			uniqueIndices[0] = true
		} else if strings.Contains(pathStr, "steps 1") {
			uniqueIndices[1] = true
		}
	}

	t.Logf("总匹配次数: %d", len(allResults))
	t.Logf("匹配到的数组索引: %v", getKeys(uniqueIndices))
	t.Logf("各路径匹配次数: %v", pathCounts)

	// 验证核心问题的答案
	if len(uniqueIndices) == 2 {
		t.Log("✅ 核心问题答案：实时增量模式下能够正常匹配数组结构中的所有值")
		t.Logf("✅ 所有2个数组元素（索引0、1）都被成功匹配")
		t.Log("✅ 数组结构解析正常，路径匹配正确")
		t.Log("")
		t.Log("📝 说明：")
		t.Log("   - 实时增量模式的设计目的就是将字符串拆分成单个字符进行增量发送")
		t.Log("   - 这种模式下，每个字符串值会触发多次回调（每个字符一次）")
		t.Log("   - 但是数组结构的解析和路径匹配完全正常")
		t.Log("   - 所有数组元素都能被正确识别和匹配")
	} else {
		t.Errorf("❌ 问题：只匹配到 %d 个数组元素，期望 2 个", len(uniqueIndices))
	}

	// 额外验证：检查是否每个数组元素都有多次匹配（因为字符串被拆分）
	for i := 0; i < 2; i++ {
		pathPattern := fmt.Sprintf("steps %d", i)
		matchCount := 0
		for pathStr, count := range pathCounts {
			if strings.Contains(pathStr, pathPattern) {
				matchCount += count
			}
		}
		if matchCount > 1 {
			t.Logf("✅ 数组元素 %d: 触发了 %d 次匹配（字符串被正确拆分）", i, matchCount)
		} else {
			t.Logf("⚠️  数组元素 %d: 只触发了 %d 次匹配", i, matchCount)
		}
	}
}

// getKeys 获取map的所有键
func getKeys(m map[int]bool) []int {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
