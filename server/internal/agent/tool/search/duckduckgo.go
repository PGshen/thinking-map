package search

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cloudwego/eino-ext/components/tool/duckduckgo"
	"github.com/cloudwego/eino-ext/components/tool/duckduckgo/ddgsearch"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
)

// SearchRequest 搜索请求参数
type SearchRequest struct {
	Query      string `json:"query"`      // 搜索关键词
	Page       int    `json:"page"`       // 页码，默认为1
	MaxResults int    `json:"maxResults"` // 最大结果数量，默认为10
}

// SearchResult 单个搜索结果
type SearchResult struct {
	Title       string `json:"title"`       // 标题
	Description string `json:"description"` // 描述
	Link        string `json:"link"`        // 链接
}

// SearchResponse 搜索响应
type SearchResponse struct {
	Results []SearchResult `json:"results"` // 搜索结果列表
	Query   string         `json:"query"`   // 搜索关键词
	Page    int            `json:"page"`    // 当前页码
	Total   int            `json:"total"`   // 结果总数
}

// DuckDuckGoSearchFunc 执行DuckDuckGo搜索
func DuckDuckGoSearchFunc(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.MaxResults <= 0 {
		req.MaxResults = 10
	}

	// 创建DuckDuckGo搜索工具配置
	config := &duckduckgo.Config{
		ToolName:   "duckduckgo_search",
		ToolDesc:   "search web for information by duckduckgo",
		Region:     ddgsearch.RegionWT, // 全球搜索
		MaxResults: req.MaxResults,
		SafeSearch: ddgsearch.SafeSearchOff,
		TimeRange:  ddgsearch.TimeRangeAll,
		DDGConfig: &ddgsearch.Config{
			Timeout:    30 * time.Second, // 30秒超时
			Cache:      true,             // 启用缓存
			MaxRetries: 3,                // 最大重试3次
		},
	}

	// 创建搜索工具实例
	searchTool, err := duckduckgo.NewTool(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create duckduckgo search tool: %w", err)
	}

	// 构建搜索请求
	searchReq := &duckduckgo.SearchRequest{
		Query: req.Query,
		Page:  req.Page,
	}

	// 序列化请求
	jsonReq, err := json.Marshal(searchReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search request: %w", err)
	}

	// 执行搜索
	resp, err := searchTool.InvokableRun(ctx, string(jsonReq))
	if err != nil {
		return nil, fmt.Errorf("failed to execute duckduckgo search: %w", err)
	}

	// 解析响应
	var ddgResp duckduckgo.SearchResponse
	if err := json.Unmarshal([]byte(resp), &ddgResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal search response: %w", err)
	}

	// 转换为我们的响应格式
	results := make([]SearchResult, 0, len(ddgResp.Results))
	for _, result := range ddgResp.Results {
		results = append(results, SearchResult{
			Title:       result.Title,
			Description: result.Description,
			Link:        result.Link,
		})
	}

	return &SearchResponse{
		Results: results,
		Query:   req.Query,
		Page:    req.Page,
		Total:   len(results),
	}, nil
}

// CreateDuckDuckGoSearchTool 创建DuckDuckGo搜索工具
func CreateDuckDuckGoSearchTool() (tool.InvokableTool, error) {
	tool := utils.NewTool(&schema.ToolInfo{
		Name: "duckduckgoSearch",
		Desc: "使用DuckDuckGo搜索引擎进行网络搜索，获取相关信息和资料",
		ParamsOneOf: schema.NewParamsOneOfByParams(
			map[string]*schema.ParameterInfo{
				"query": {
					Type:     schema.String,
					Desc:     "搜索关键词或问题",
					Required: true,
				},
				"page": {
					Type: schema.Number,
					Desc: "页码，默认为1",
				},
				"maxResults": {
					Type: schema.Number,
					Desc: "最大结果数量，默认为10，建议不超过20",
				},
			},
		),
	}, DuckDuckGoSearchFunc)
	return tool, nil
}

// GetAllSearchTools 获取所有搜索工具
func GetAllSearchTools() ([]tool.BaseTool, error) {
	duckduckgoTool, err := CreateDuckDuckGoSearchTool()
	if err != nil {
		return nil, err
	}

	return []tool.BaseTool{duckduckgoTool}, nil
}

// GetAllToolInfos 获取所有搜索工具信息
func GetAllToolInfos(ctx context.Context) ([]*schema.ToolInfo, error) {
	tools, err := GetAllSearchTools()
	if err != nil {
		return nil, err
	}

	toolInfos := make([]*schema.ToolInfo, 0, len(tools))
	for _, tool := range tools {
		info, err := tool.Info(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get tool info: %w", err)
		}
		toolInfos = append(toolInfos, info)
	}
	return toolInfos, nil
}
