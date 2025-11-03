package search

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/cloudwego/eino-ext/components/tool/googlesearch"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
)

// GoogleSearchRequest Google搜索请求参数
type GoogleSearchRequest struct {
	Query  string `json:"query"`  // 搜索关键词
	Num    int    `json:"num"`    // 返回结果数量，默认为5
	Offset int    `json:"offset"` // 结果起始位置，默认为0
	Lang   string `json:"lang"`   // 搜索语言，默认为zh-CN
}

// GoogleSearchResult Google搜索单个结果
type GoogleSearchResult struct {
	Title   string `json:"title"`   // 标题
	Link    string `json:"link"`    // 链接
	Snippet string `json:"snippet"` // 摘要
	Desc    string `json:"desc"`    // 描述
}

// GoogleSearchResponse Google搜索响应
type GoogleSearchResponse struct {
	Query string               `json:"query"` // 搜索关键词
	Items []GoogleSearchResult `json:"items"` // 搜索结果列表
	Total int                  `json:"total"` // 结果总数
}

// GoogleSearchFunc 执行Google搜索
func GoogleSearchFunc(ctx context.Context, req *GoogleSearchRequest) (*GoogleSearchResponse, error) {
	// 设置默认值
	if req.Num <= 0 {
		req.Num = 5
	}
	if req.Offset < 0 {
		req.Offset = 0
	}
	if req.Lang == "" {
		req.Lang = "zh-CN"
	}

	// 从环境变量获取API配置
	googleAPIKey := os.Getenv("GOOGLE_API_KEY")
	googleSearchEngineID := os.Getenv("GOOGLE_SEARCH_ENGINE_ID")

	if googleAPIKey == "" || googleSearchEngineID == "" {
		return nil, fmt.Errorf("Google API配置缺失: 请设置GOOGLE_API_KEY和GOOGLE_SEARCH_ENGINE_ID环境变量")
	}

	// 创建Google搜索工具配置
	config := &googlesearch.Config{
		APIKey:         googleAPIKey,
		SearchEngineID: googleSearchEngineID,
		Lang:           req.Lang,
		Num:            req.Num,
		ToolName:       "google_search",
		ToolDesc:       "Google Custom Search API工具，用于进行高质量的网络搜索",
	}

	// 创建搜索工具实例
	searchTool, err := googlesearch.NewTool(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("创建Google搜索工具失败: %w", err)
	}

	// 构建搜索请求
	searchReq := googlesearch.SearchRequest{
		Query:  req.Query,
		Num:    req.Num,
		Offset: req.Offset,
		Lang:   req.Lang,
	}

	// 序列化请求
	jsonReq, err := json.Marshal(searchReq)
	if err != nil {
		return nil, fmt.Errorf("序列化搜索请求失败: %w", err)
	}

	// 执行搜索
	resp, err := searchTool.InvokableRun(ctx, string(jsonReq))
	if err != nil {
		return nil, fmt.Errorf("执行Google搜索失败: %w", err)
	}

	// 解析响应
	var googleResp googlesearch.SearchResult
	if err := json.Unmarshal([]byte(resp), &googleResp); err != nil {
		return nil, fmt.Errorf("解析搜索响应失败: %w", err)
	}

	// 转换为我们的响应格式
	results := make([]GoogleSearchResult, 0, len(googleResp.Items))
	for _, item := range googleResp.Items {
		results = append(results, GoogleSearchResult{
			Title:   item.Title,
			Link:    item.Link,
			Snippet: item.Snippet,
			Desc:    item.Desc,
		})
	}

	return &GoogleSearchResponse{
		Query: req.Query,
		Items: results,
		Total: len(results),
	}, nil
}

// CreateGoogleSearchTool 创建Google搜索工具
func CreateGoogleSearchTool() (tool.InvokableTool, error) {
	tool := utils.NewTool(&schema.ToolInfo{
		Name: "googleSearch",
		Desc: "使用Google Custom Search API进行网络搜索，获取高质量的搜索结果和相关信息",
		ParamsOneOf: schema.NewParamsOneOfByParams(
			map[string]*schema.ParameterInfo{
				"query": {
					Type:     schema.String,
					Desc:     "搜索关键词或问题",
					Required: true,
				},
				"num": {
					Type: schema.Number,
					Desc: "返回结果数量，默认为5，最大为10",
				},
				"offset": {
					Type: schema.Number,
					Desc: "结果起始位置，用于分页，默认为0",
				},
				"lang": {
					Type: schema.String,
					Desc: "搜索语言，如zh-CN（中文）、en（英文），默认为zh-CN",
				},
			},
		),
	}, GoogleSearchFunc)
	return tool, nil
}

// GetAllSearchToolsWithGoogle 获取所有搜索工具（包括Google搜索）
func GetAllSearchToolsWithGoogle() ([]tool.BaseTool, error) {
	var tools []tool.BaseTool

	// 添加DuckDuckGo搜索工具
	duckduckgoTool, err := CreateDuckDuckGoSearchTool()
	if err != nil {
		return nil, fmt.Errorf("创建DuckDuckGo搜索工具失败: %w", err)
	}
	tools = append(tools, duckduckgoTool)

	// 添加Google搜索工具
	googleTool, err := CreateGoogleSearchTool()
	if err != nil {
		return nil, fmt.Errorf("创建Google搜索工具失败: %w", err)
	}
	tools = append(tools, googleTool)

	return tools, nil
}

// GetAllToolInfosWithGoogle 获取所有搜索工具信息（包括Google搜索）
func GetAllToolInfosWithGoogle(ctx context.Context) ([]*schema.ToolInfo, error) {
	tools, err := GetAllSearchToolsWithGoogle()
	if err != nil {
		return nil, err
	}

	toolInfos := make([]*schema.ToolInfo, 0, len(tools))
	for _, tool := range tools {
		info, err := tool.Info(ctx)
		if err != nil {
			return nil, fmt.Errorf("获取工具信息失败: %w", err)
		}
		toolInfos = append(toolInfos, info)
	}
	return toolInfos, nil
}