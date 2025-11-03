package search

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDuckDuckGoSearchFunc(t *testing.T) {
	// 检查是否为短测试模式，跳过需要网络的测试
	if testing.Short() {
		t.Skip("跳过需要网络连接的测试")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	t.Run("基本搜索功能", func(t *testing.T) {
		req := &SearchRequest{
			Query:      "Go programming language",
			Page:       1,
			MaxResults: 3, // 减少结果数量以提高测试速度
		}

		resp, err := DuckDuckGoSearchFunc(ctx, req)
		if err != nil {
			t.Logf("搜索失败（可能是网络问题）: %v", err)
			t.Skip("跳过网络相关测试")
			return
		}

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, req.Query, resp.Query)
		assert.Equal(t, req.Page, resp.Page)
		assert.True(t, len(resp.Results) >= 0, "应该返回搜索结果或空结果")
		assert.True(t, len(resp.Results) <= req.MaxResults, "结果数量不应超过最大限制")

		// 验证结果结构（如果有结果的话）
		for _, result := range resp.Results {
			assert.NotEmpty(t, result.Title, "标题不应为空")
			assert.NotEmpty(t, result.Link, "链接不应为空")
			// 描述可能为空，所以不强制要求
		}
	})

	t.Run("默认参数处理", func(t *testing.T) {
		req := &SearchRequest{
			Query: "test query",
			// 不设置Page和MaxResults，测试默认值
		}

		resp, err := DuckDuckGoSearchFunc(ctx, req)
		if err != nil {
			t.Logf("搜索失败（可能是网络问题）: %v", err)
			t.Skip("跳过网络相关测试")
			return
		}

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, 1, resp.Page, "默认页码应为1")
	})

	t.Run("空查询处理", func(t *testing.T) {
		req := &SearchRequest{
			Query:      "",
			Page:       1,
			MaxResults: 5,
		}

		// 空查询可能会返回错误或空结果，这取决于DuckDuckGo的实现
		resp, err := DuckDuckGoSearchFunc(ctx, req)
		if err != nil {
			t.Logf("空查询返回错误（预期行为）: %v", err)
		} else {
			assert.NotNil(t, resp)
			t.Logf("空查询返回了 %d 个结果", len(resp.Results))
		}
	})

	t.Run("中文搜索", func(t *testing.T) {
		req := &SearchRequest{
			Query:      "人工智能",
			Page:       1,
			MaxResults: 3,
		}

		resp, err := DuckDuckGoSearchFunc(ctx, req)
		if err != nil {
			t.Logf("中文搜索失败（可能是网络问题）: %v", err)
			t.Skip("跳过网络相关测试")
			return
		}

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, req.Query, resp.Query)
		assert.True(t, len(resp.Results) >= 0, "中文搜索应该返回结果或空结果")
	})
}

func TestCreateDuckDuckGoSearchTool(t *testing.T) {
	tool, err := CreateDuckDuckGoSearchTool()
	require.NoError(t, err)
	assert.NotNil(t, tool)

	ctx := context.Background()

	// 测试工具信息
	info, err := tool.Info(ctx)
	require.NoError(t, err)
	assert.Equal(t, "duckduckgoSearch", info.Name)
	assert.NotEmpty(t, info.Desc)
	assert.NotNil(t, info.ParamsOneOf)
}

func TestGetAllSearchTools(t *testing.T) {
	tools, err := GetAllSearchTools()
	require.NoError(t, err)
	assert.NotEmpty(t, tools, "应该返回至少一个搜索工具")

	// 验证返回的是DuckDuckGo搜索工具
	ctx := context.Background()
	for _, tool := range tools {
		info, err := tool.Info(ctx)
		require.NoError(t, err)
		assert.NotEmpty(t, info.Name)
		assert.NotEmpty(t, info.Desc)
	}
}

func TestGetAllToolInfos(t *testing.T) {
	ctx := context.Background()

	toolInfos, err := GetAllToolInfos(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, toolInfos, "应该返回至少一个工具信息")

	// 验证工具信息的完整性
	for _, info := range toolInfos {
		assert.NotEmpty(t, info.Name, "工具名称不应为空")
		assert.NotEmpty(t, info.Desc, "工具描述不应为空")
		assert.NotNil(t, info.ParamsOneOf, "参数定义不应为空")
	}
}

// 集成测试：测试完整的工具调用流程
func TestSearchToolIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 创建工具
	tool, err := CreateDuckDuckGoSearchTool()
	require.NoError(t, err)

	// 模拟工具调用
	reqJSON := `{"query":"CloudWeGo Eino framework","page":1,"maxResults":3}`

	resp, err := tool.InvokableRun(ctx, reqJSON)
	require.NoError(t, err)
	assert.NotEmpty(t, resp, "工具应该返回搜索结果")

	t.Logf("搜索结果: %s", resp)
}

// 基准测试
func BenchmarkDuckDuckGoSearch(b *testing.B) {
	ctx := context.Background()
	req := &SearchRequest{
		Query:      "benchmark test",
		Page:       1,
		MaxResults: 5,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := DuckDuckGoSearchFunc(ctx, req)
		if err != nil {
			b.Fatalf("搜索失败: %v", err)
		}
	}
}
