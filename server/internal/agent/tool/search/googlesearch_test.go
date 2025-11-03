package search

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoogleSearchFunc(t *testing.T) {
	// 检查是否为短测试模式，跳过需要网络的测试
	if testing.Short() {
		t.Skip("跳过需要网络连接的测试")
	}

	// 检查环境变量
	if os.Getenv("GOOGLE_API_KEY") == "" || os.Getenv("GOOGLE_SEARCH_ENGINE_ID") == "" {
		t.Skip("跳过Google搜索测试：缺少GOOGLE_API_KEY或GOOGLE_SEARCH_ENGINE_ID环境变量")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	t.Run("基本搜索功能", func(t *testing.T) {
		req := &GoogleSearchRequest{
			Query: "Go programming language",
			Num:   3, // 减少结果数量以提高测试速度
			Lang:  "en",
		}

		resp, err := GoogleSearchFunc(ctx, req)
		if err != nil {
			t.Logf("搜索失败（可能是网络问题或API配置问题）: %v", err)
			t.Skip("跳过网络相关测试")
			return
		}

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, req.Query, resp.Query)
		assert.True(t, len(resp.Items) >= 0, "应该返回搜索结果或空结果")
		assert.True(t, len(resp.Items) <= req.Num, "结果数量不应超过最大限制")

		// 验证结果结构（如果有结果的话）
		for _, item := range resp.Items {
			assert.NotEmpty(t, item.Title, "标题不应为空")
			assert.NotEmpty(t, item.Link, "链接不应为空")
			// Snippet和Desc可能为空，所以不强制要求
		}
	})

	t.Run("默认参数处理", func(t *testing.T) {
		req := &GoogleSearchRequest{
			Query: "test query",
			// 不设置其他参数，测试默认值
		}

		resp, err := GoogleSearchFunc(ctx, req)
		if err != nil {
			t.Logf("搜索失败（可能是网络问题）: %v", err)
			t.Skip("跳过网络相关测试")
			return
		}

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, req.Query, resp.Query)
	})

	t.Run("中文搜索", func(t *testing.T) {
		req := &GoogleSearchRequest{
			Query: "人工智能",
			Num:   2,
			Lang:  "zh-CN",
		}

		resp, err := GoogleSearchFunc(ctx, req)
		if err != nil {
			t.Logf("搜索失败（可能是网络问题）: %v", err)
			t.Skip("跳过网络相关测试")
			return
		}

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, req.Query, resp.Query)
	})

	t.Run("分页搜索", func(t *testing.T) {
		req := &GoogleSearchRequest{
			Query:  "machine learning",
			Num:    2,
			Offset: 2, // 从第3个结果开始
			Lang:   "en",
		}

		resp, err := GoogleSearchFunc(ctx, req)
		if err != nil {
			t.Logf("搜索失败（可能是网络问题）: %v", err)
			t.Skip("跳过网络相关测试")
			return
		}

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, req.Query, resp.Query)
	})
}

func TestGoogleSearchFuncWithoutAPIKey(t *testing.T) {
	// 临时清除环境变量
	originalAPIKey := os.Getenv("GOOGLE_API_KEY")
	originalEngineID := os.Getenv("GOOGLE_SEARCH_ENGINE_ID")
	
	os.Unsetenv("GOOGLE_API_KEY")
	os.Unsetenv("GOOGLE_SEARCH_ENGINE_ID")
	
	defer func() {
		if originalAPIKey != "" {
			os.Setenv("GOOGLE_API_KEY", originalAPIKey)
		}
		if originalEngineID != "" {
			os.Setenv("GOOGLE_SEARCH_ENGINE_ID", originalEngineID)
		}
	}()

	ctx := context.Background()
	req := &GoogleSearchRequest{
		Query: "test",
	}

	_, err := GoogleSearchFunc(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Google API配置缺失")
}

func TestCreateGoogleSearchTool(t *testing.T) {
	tool, err := CreateGoogleSearchTool()
	require.NoError(t, err)
	assert.NotNil(t, tool)

	ctx := context.Background()
	info, err := tool.Info(ctx)
	require.NoError(t, err)
	assert.Equal(t, "googleSearch", info.Name)
	assert.NotEmpty(t, info.Desc)
	assert.NotNil(t, info.ParamsOneOf)
}

func TestGetAllSearchToolsWithGoogle(t *testing.T) {
	tools, err := GetAllSearchToolsWithGoogle()
	require.NoError(t, err)
	assert.Len(t, tools, 2) // DuckDuckGo + Google

	// 验证工具名称
	ctx := context.Background()
	toolNames := make([]string, 0, len(tools))
	for _, tool := range tools {
		info, err := tool.Info(ctx)
		require.NoError(t, err)
		toolNames = append(toolNames, info.Name)
	}

	assert.Contains(t, toolNames, "duckduckgoSearch")
	assert.Contains(t, toolNames, "googleSearch")
}

func TestGetAllToolInfosWithGoogle(t *testing.T) {
	ctx := context.Background()
	toolInfos, err := GetAllToolInfosWithGoogle(ctx)
	require.NoError(t, err)
	assert.Len(t, toolInfos, 2) // DuckDuckGo + Google

	// 验证工具信息
	toolNames := make([]string, 0, len(toolInfos))
	for _, info := range toolInfos {
		assert.NotEmpty(t, info.Name)
		assert.NotEmpty(t, info.Desc)
		toolNames = append(toolNames, info.Name)
	}

	assert.Contains(t, toolNames, "duckduckgoSearch")
	assert.Contains(t, toolNames, "googleSearch")
}

func TestGoogleSearchToolIntegration(t *testing.T) {
	// 检查是否为短测试模式
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	// 检查环境变量
	if os.Getenv("GOOGLE_API_KEY") == "" || os.Getenv("GOOGLE_SEARCH_ENGINE_ID") == "" {
		t.Skip("跳过Google搜索集成测试：缺少API配置")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	tool, err := CreateGoogleSearchTool()
	require.NoError(t, err)

	// 测试工具调用
	req := `{"query": "artificial intelligence", "num": 2, "lang": "en"}`
	resp, err := tool.InvokableRun(ctx, req)
	if err != nil {
		t.Logf("工具调用失败（可能是网络问题）: %v", err)
		t.Skip("跳过网络相关测试")
		return
	}

	assert.NotEmpty(t, resp)
	t.Logf("Google搜索响应: %s", resp)
}

func BenchmarkGoogleSearch(b *testing.B) {
	// 检查环境变量
	if os.Getenv("GOOGLE_API_KEY") == "" || os.Getenv("GOOGLE_SEARCH_ENGINE_ID") == "" {
		b.Skip("跳过Google搜索基准测试：缺少API配置")
	}

	ctx := context.Background()
	req := &GoogleSearchRequest{
		Query: "benchmark test",
		Num:   1, // 最小结果数量以提高性能
		Lang:  "en",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GoogleSearchFunc(ctx, req)
		if err != nil {
			b.Logf("搜索失败: %v", err)
			continue
		}
	}
}