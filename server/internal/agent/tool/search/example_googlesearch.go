package search

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// ExampleGoogleBasicUsage 演示Google搜索的基本使用
func ExampleGoogleBasicUsage() {
	// 检查环境变量
	if os.Getenv("GOOGLE_API_KEY") == "" || os.Getenv("GOOGLE_SEARCH_ENGINE_ID") == "" {
		fmt.Println("请设置GOOGLE_API_KEY和GOOGLE_SEARCH_ENGINE_ID环境变量")
		return
	}

	ctx := context.Background()

	// 创建Google搜索请求
	req := &GoogleSearchRequest{
		Query: "artificial intelligence machine learning",
		Num:   5,
		Lang:  "en",
	}

	// 执行搜索
	resp, err := GoogleSearchFunc(ctx, req)
	if err != nil {
		log.Printf("Google搜索失败: %v", err)
		return
	}

	// 打印搜索结果
	fmt.Printf("Google搜索关键词: %s\n", resp.Query)
	fmt.Printf("找到 %d 个结果:\n\n", resp.Total)

	for i, item := range resp.Items {
		fmt.Printf("%d. %s\n", i+1, item.Title)
		fmt.Printf("   链接: %s\n", item.Link)
		if item.Snippet != "" {
			fmt.Printf("   摘要: %s\n", item.Snippet)
		}
		if item.Desc != "" {
			fmt.Printf("   描述: %s\n", item.Desc)
		}
		fmt.Println()
	}
}

// ExampleGoogleToolUsage 演示Google搜索作为Eino工具的使用
func ExampleGoogleToolUsage() {
	// 检查环境变量
	if os.Getenv("GOOGLE_API_KEY") == "" || os.Getenv("GOOGLE_SEARCH_ENGINE_ID") == "" {
		fmt.Println("请设置GOOGLE_API_KEY和GOOGLE_SEARCH_ENGINE_ID环境变量")
		return
	}

	ctx := context.Background()

	// 创建Google搜索工具
	tool, err := CreateGoogleSearchTool()
	if err != nil {
		log.Printf("创建Google搜索工具失败: %v", err)
		return
	}

	// 获取工具信息
	info, err := tool.Info(ctx)
	if err != nil {
		log.Printf("获取工具信息失败: %v", err)
		return
	}

	fmt.Printf("工具名称: %s\n", info.Name)
	fmt.Printf("工具描述: %s\n", info.Desc)
	fmt.Println()

	// 构建搜索请求
	searchReq := map[string]interface{}{
		"query": "CloudWeGo Eino framework",
		"num":   3,
		"lang":  "zh-CN",
	}

	// 序列化请求
	jsonReq, err := json.Marshal(searchReq)
	if err != nil {
		log.Printf("序列化请求失败: %v", err)
		return
	}

	// 调用工具
	result, err := tool.InvokableRun(ctx, string(jsonReq))
	if err != nil {
		log.Printf("工具调用失败: %v", err)
		return
	}

	fmt.Printf("搜索结果:\n%s\n", result)
}

// ExampleGoogleChineseSearch 演示中文搜索
func ExampleGoogleChineseSearch() {
	// 检查环境变量
	if os.Getenv("GOOGLE_API_KEY") == "" || os.Getenv("GOOGLE_SEARCH_ENGINE_ID") == "" {
		fmt.Println("请设置GOOGLE_API_KEY和GOOGLE_SEARCH_ENGINE_ID环境变量")
		return
	}

	ctx := context.Background()

	// 创建中文搜索请求
	req := &GoogleSearchRequest{
		Query: "人工智能 深度学习",
		Num:   3,
		Lang:  "zh-CN",
	}

	// 执行搜索
	resp, err := GoogleSearchFunc(ctx, req)
	if err != nil {
		log.Printf("中文搜索失败: %v", err)
		return
	}

	// 打印搜索结果
	fmt.Printf("中文搜索关键词: %s\n", resp.Query)
	fmt.Printf("找到 %d 个结果:\n\n", resp.Total)

	for i, item := range resp.Items {
		fmt.Printf("%d. %s\n", i+1, item.Title)
		fmt.Printf("   链接: %s\n", item.Link)
		if item.Snippet != "" {
			fmt.Printf("   摘要: %s\n", item.Snippet)
		}
		fmt.Println()
	}
}

// ExampleGooglePaginatedSearch 演示分页搜索
func ExampleGooglePaginatedSearch() {
	// 检查环境变量
	if os.Getenv("GOOGLE_API_KEY") == "" || os.Getenv("GOOGLE_SEARCH_ENGINE_ID") == "" {
		fmt.Println("请设置GOOGLE_API_KEY和GOOGLE_SEARCH_ENGINE_ID环境变量")
		return
	}

	ctx := context.Background()

	// 第一页搜索
	fmt.Println("=== 第一页搜索结果 ===")
	req1 := &GoogleSearchRequest{
		Query:  "golang tutorial",
		Num:    3,
		Offset: 0, // 从第1个结果开始
		Lang:   "en",
	}

	resp1, err := GoogleSearchFunc(ctx, req1)
	if err != nil {
		log.Printf("第一页搜索失败: %v", err)
		return
	}

	for i, item := range resp1.Items {
		fmt.Printf("%d. %s\n", i+1, item.Title)
		fmt.Printf("   链接: %s\n", item.Link)
		fmt.Println()
	}

	// 第二页搜索
	fmt.Println("=== 第二页搜索结果 ===")
	req2 := &GoogleSearchRequest{
		Query:  "golang tutorial",
		Num:    3,
		Offset: 3, // 从第4个结果开始
		Lang:   "en",
	}

	resp2, err := GoogleSearchFunc(ctx, req2)
	if err != nil {
		log.Printf("第二页搜索失败: %v", err)
		return
	}

	for i, item := range resp2.Items {
		fmt.Printf("%d. %s\n", i+4, item.Title) // 继续编号
		fmt.Printf("   链接: %s\n", item.Link)
		fmt.Println()
	}
}

// ExampleAllSearchTools 演示使用所有搜索工具（DuckDuckGo + Google）
func ExampleAllSearchTools() {
	ctx := context.Background()

	// 获取所有搜索工具
	tools, err := GetAllSearchToolsWithGoogle()
	if err != nil {
		log.Printf("获取搜索工具失败: %v", err)
		return
	}

	fmt.Printf("可用的搜索工具数量: %d\n\n", len(tools))

	// 遍历所有工具并显示信息
	for i, tool := range tools {
		info, err := tool.Info(ctx)
		if err != nil {
			log.Printf("获取工具 %d 信息失败: %v", i, err)
			continue
		}

		fmt.Printf("工具 %d:\n", i+1)
		fmt.Printf("  名称: %s\n", info.Name)
		fmt.Printf("  描述: %s\n", info.Desc)
		fmt.Println()
	}
}

// ExampleGoogleErrorHandling 演示错误处理
func ExampleGoogleErrorHandling() {
	ctx := context.Background()

	// 测试缺少API配置的情况
	fmt.Println("=== 测试缺少API配置 ===")
	
	// 临时清除环境变量
	originalAPIKey := os.Getenv("GOOGLE_API_KEY")
	originalEngineID := os.Getenv("GOOGLE_SEARCH_ENGINE_ID")
	
	os.Unsetenv("GOOGLE_API_KEY")
	os.Unsetenv("GOOGLE_SEARCH_ENGINE_ID")

	req := &GoogleSearchRequest{
		Query: "test query",
	}

	_, err := GoogleSearchFunc(ctx, req)
	if err != nil {
		fmt.Printf("预期的错误: %v\n", err)
	}

	// 恢复环境变量
	if originalAPIKey != "" {
		os.Setenv("GOOGLE_API_KEY", originalAPIKey)
	}
	if originalEngineID != "" {
		os.Setenv("GOOGLE_SEARCH_ENGINE_ID", originalEngineID)
	}

	// 测试空查询
	fmt.Println("\n=== 测试空查询 ===")
	emptyReq := &GoogleSearchRequest{
		Query: "",
	}

	resp, err := GoogleSearchFunc(ctx, emptyReq)
	if err != nil {
		fmt.Printf("空查询错误: %v\n", err)
	} else if resp != nil {
		fmt.Printf("空查询返回了 %d 个结果\n", resp.Total)
	}
}

// ExampleCompareSearchEngines 演示比较不同搜索引擎的结果
func ExampleCompareSearchEngines() {
	// 检查Google API配置
	if os.Getenv("GOOGLE_API_KEY") == "" || os.Getenv("GOOGLE_SEARCH_ENGINE_ID") == "" {
		fmt.Println("请设置GOOGLE_API_KEY和GOOGLE_SEARCH_ENGINE_ID环境变量以进行比较")
		return
	}

	ctx := context.Background()
	query := "machine learning algorithms"

	fmt.Printf("搜索关键词: %s\n\n", query)

	// DuckDuckGo搜索
	fmt.Println("=== DuckDuckGo搜索结果 ===")
	ddgReq := &SearchRequest{
		Query:      query,
		MaxResults: 3,
	}

	ddgResp, err := DuckDuckGoSearchFunc(ctx, ddgReq)
	if err != nil {
		fmt.Printf("DuckDuckGo搜索失败: %v\n", err)
	} else {
		for i, result := range ddgResp.Results {
			fmt.Printf("%d. %s\n", i+1, result.Title)
			fmt.Printf("   %s\n", result.Link)
		}
	}

	fmt.Println()

	// Google搜索
	fmt.Println("=== Google搜索结果 ===")
	googleReq := &GoogleSearchRequest{
		Query: query,
		Num:   3,
		Lang:  "en",
	}

	googleResp, err := GoogleSearchFunc(ctx, googleReq)
	if err != nil {
		fmt.Printf("Google搜索失败: %v\n", err)
	} else {
		for i, item := range googleResp.Items {
			fmt.Printf("%d. %s\n", i+1, item.Title)
			fmt.Printf("   %s\n", item.Link)
		}
	}
}