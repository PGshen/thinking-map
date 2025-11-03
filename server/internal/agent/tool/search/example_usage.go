package search

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
)

// ExampleBasicUsage 演示基本搜索功能的使用
func ExampleBasicUsage() {
	ctx := context.Background()

	// 创建搜索请求
	req := &SearchRequest{
		Query:      "Go programming language tutorial",
		Page:       1,
		MaxResults: 5,
	}

	// 执行搜索
	resp, err := DuckDuckGoSearchFunc(ctx, req)
	if err != nil {
		log.Printf("搜索失败: %v", err)
		return
	}

	// 打印搜索结果
	fmt.Printf("搜索关键词: %s\n", resp.Query)
	fmt.Printf("找到 %d 个结果:\n\n", resp.Total)

	for i, result := range resp.Results {
		fmt.Printf("%d. %s\n", i+1, result.Title)
		fmt.Printf("   链接: %s\n", result.Link)
		if result.Description != "" {
			fmt.Printf("   描述: %s\n", result.Description)
		}
		fmt.Println()
	}
}

// ExampleToolUsage 演示作为Eino工具的使用
func ExampleToolUsage() {
	ctx := context.Background()

	// 创建搜索工具
	tool, err := CreateDuckDuckGoSearchTool()
	if err != nil {
		log.Printf("创建搜索工具失败: %v", err)
		return
	}

	// 获取工具信息
	info, err := tool.Info(ctx)
	if err != nil {
		log.Printf("获取工具信息失败: %v", err)
		return
	}

	fmt.Printf("工具名称: %s\n", info.Name)
	fmt.Printf("工具描述: %s\n\n", info.Desc)

	// 准备搜索请求JSON
	searchReq := map[string]interface{}{
		"query":      "CloudWeGo Eino framework",
		"page":       1,
		"maxResults": 3,
	}

	reqJSON, err := json.Marshal(searchReq)
	if err != nil {
		log.Printf("序列化请求失败: %v", err)
		return
	}

	// 执行搜索
	respJSON, err := tool.InvokableRun(ctx, string(reqJSON))
	if err != nil {
		log.Printf("执行搜索失败: %v", err)
		return
	}

	// 解析响应
	var resp SearchResponse
	if err := json.Unmarshal([]byte(respJSON), &resp); err != nil {
		log.Printf("解析响应失败: %v", err)
		return
	}

	fmt.Printf("搜索结果:\n")
	for i, result := range resp.Results {
		fmt.Printf("%d. %s\n", i+1, result.Title)
		fmt.Printf("   %s\n\n", result.Link)
	}
}

// ExampleMultipleTools 演示获取所有搜索工具
func ExampleMultipleTools() {
	ctx := context.Background()

	// 获取所有搜索工具
	tools, err := GetAllSearchTools()
	if err != nil {
		log.Printf("获取搜索工具失败: %v", err)
		return
	}

	fmt.Printf("可用的搜索工具数量: %d\n\n", len(tools))

	// 获取工具信息
	toolInfos, err := GetAllToolInfos(ctx)
	if err != nil {
		log.Printf("获取工具信息失败: %v", err)
		return
	}

	for i, info := range toolInfos {
		fmt.Printf("工具 %d:\n", i+1)
		fmt.Printf("  名称: %s\n", info.Name)
		fmt.Printf("  描述: %s\n", info.Desc)
		
		// 打印参数信息
		fmt.Printf("  参数: 支持query(搜索关键词), page(页码), maxResults(最大结果数)\n")
		fmt.Println()
	}
}

// ExampleChineseSearch 演示中文搜索
func ExampleChineseSearch() {
	ctx := context.Background()

	// 创建中文搜索请求
	req := &SearchRequest{
		Query:      "人工智能发展趋势",
		Page:       1,
		MaxResults: 3,
	}

	// 执行搜索
	resp, err := DuckDuckGoSearchFunc(ctx, req)
	if err != nil {
		log.Printf("中文搜索失败: %v", err)
		return
	}

	fmt.Printf("中文搜索结果 - 关键词: %s\n", resp.Query)
	fmt.Printf("找到 %d 个结果:\n\n", resp.Total)

	for i, result := range resp.Results {
		fmt.Printf("%d. %s\n", i+1, result.Title)
		fmt.Printf("   %s\n", result.Link)
		if result.Description != "" {
			fmt.Printf("   %s\n", result.Description)
		}
		fmt.Println()
	}
}

// ExampleErrorHandling 演示错误处理
func ExampleErrorHandling() {
	ctx := context.Background()

	// 测试各种错误情况
	testCases := []struct {
		name string
		req  *SearchRequest
	}{
		{
			name: "空查询",
			req: &SearchRequest{
				Query:      "",
				Page:       1,
				MaxResults: 5,
			},
		},
		{
			name: "负数页码",
			req: &SearchRequest{
				Query:      "test",
				Page:       -1,
				MaxResults: 5,
			},
		},
		{
			name: "过大的结果数量",
			req: &SearchRequest{
				Query:      "test",
				Page:       1,
				MaxResults: 100,
			},
		},
	}

	for _, tc := range testCases {
		fmt.Printf("测试用例: %s\n", tc.name)
		
		resp, err := DuckDuckGoSearchFunc(ctx, tc.req)
		if err != nil {
			fmt.Printf("  错误: %v\n", err)
		} else {
			fmt.Printf("  成功: 返回 %d 个结果\n", len(resp.Results))
		}
		fmt.Println()
	}
}