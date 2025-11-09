package search

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/PGshen/thinking-map/server/internal/global"
	"github.com/PGshen/thinking-map/server/internal/model"
	"github.com/PGshen/thinking-map/server/internal/model/dto"
	"github.com/PGshen/thinking-map/server/internal/pkg/logger"
	"github.com/PGshen/thinking-map/server/internal/pkg/sse"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	DefaultSearchDepth = "basic"
	DefaultMaxResults  = 5
	TavilyAPIURL       = "https://api.tavily.com"
)

type TavilyClient struct {
	APIKey     string
	HttpClient *http.Client
}

func NewTavilyClient() *TavilyClient {
	apiKey := viper.GetString("service.tavily.api_key")
	timeout := viper.GetDuration("service.tavily.timeout")
	if timeout == 0 {
		timeout = 120 * time.Second
	}
	return &TavilyClient{
		APIKey: apiKey,
		HttpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

type SearchRequest struct {
	Query             string   `json:"query"`
	SearchDepth       string   `json:"search_depth,omitempty"`
	IncludeAnswer     bool     `json:"include_answer,omitempty"`
	IncludeRawContent bool     `json:"include_raw_content,omitempty"`
	MaxResults        int      `json:"max_results,omitempty"`
	IncludeDomains    []string `json:"include_domains,omitempty"`
	ExcludeDomains    []string `json:"exclude_domains,omitempty"`
	IncludeFavicon    bool     `json:"include_favicon,omitempty"`
}

type SearchResult struct {
	Title      string  `json:"title"`
	URL        string  `json:"url"`
	Content    string  `json:"content"`
	Score      float64 `json:"score"`
	RawContent string  `json:"raw_content,omitempty"`
	Favicon    string  `json:"favicon,omitempty"`
}

type SearchResponse struct {
	Answer       string         `json:"answer,omitempty"`
	Query        string         `json:"query"`
	ResponseTime float64        `json:"response_time"`
	Results      []SearchResult `json:"results"`
}

func (c *TavilyClient) Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", TavilyAPIURL+"/search", strings.NewReader(string(reqBytes)))
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HttpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send search request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("search request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var searchResp SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	return &searchResp, nil
}

type ExtractRequest struct {
	Query     string `json:"query"`
	URL       string `json:"url"`
	MaxTokens int    `json:"max_tokens,omitempty"`
}

type ExtractResponse struct {
	Content string `json:"content"`
}

func (c *TavilyClient) Extract(ctx context.Context, req *ExtractRequest) (*ExtractResponse, error) {
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal extract request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", TavilyAPIURL+"/extract", strings.NewReader(string(reqBytes)))
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HttpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send extract request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("extract request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var extractResp ExtractResponse
	if err := json.NewDecoder(resp.Body).Decode(&extractResp); err != nil {
		return nil, fmt.Errorf("failed to decode extract response: %w", err)
	}

	return &extractResp, nil
}

type SearchFuncRequest struct {
	Query             string   `json:"query"`
	SearchDepth       string   `json:"search_depth,omitempty"`
	IncludeAnswer     bool     `json:"include_answer,omitempty"`
	IncludeRawContent bool     `json:"include_raw_content,omitempty"`
	MaxResults        int      `json:"max_results,omitempty"`
	IncludeDomains    []string `json:"include_domains,omitempty"`
	ExcludeDomains    []string `json:"exclude_domains,omitempty"`
	IncludeFavicon    bool     `json:"include_favicon,omitempty"`
}

func SearchFunc(ctx context.Context, req *SearchFuncRequest) (*SearchResponse, error) {
	mapID := ctx.Value("mapID").(string)
	nodeID := ctx.Value("nodeID").(string)
	client := NewTavilyClient()

	searchReq := &SearchRequest{
		Query:             req.Query,
		SearchDepth:       req.SearchDepth,
		IncludeAnswer:     req.IncludeAnswer,
		IncludeRawContent: req.IncludeRawContent,
		MaxResults:        req.MaxResults,
		IncludeDomains:    req.IncludeDomains,
		ExcludeDomains:    req.ExcludeDomains,
		IncludeFavicon:    true,
	}
	if searchReq.SearchDepth == "" {
		searchReq.SearchDepth = DefaultSearchDepth
	}
	if searchReq.MaxResults == 0 {
		searchReq.MaxResults = DefaultMaxResults
	}

	notice := model.Notice{
		Type:    model.NoticeTypeInfo,
		Name:    "网络检索",
		Content: fmt.Sprintf("关键词: %s", req.Query),
	}

	// 发送开始检索的消息给前端
	global.GetBroker().PublishToSession(mapID, sse.Event{
		ID:   uuid.NewString(),
		Type: dto.MessageNoticeEventType,
		Data: dto.MessageNoticeEvent{
			NodeID:    nodeID,
			MessageID: uuid.NewString(),
			Notice:    notice,
		},
	})

	global.GetMessageManager().SaveDecompositionMessage(ctx, nodeID, dto.CreateMessageRequest{
		ID:          uuid.NewString(),
		MessageType: model.MsgTypeNotice,
		Role:        schema.Assistant,
		Content:     model.MessageContent{Notice: &notice},
	})

	resp, err := client.Search(ctx, searchReq)
	if err != nil {
		return nil, err
	}

	// 将搜索结果转换为ragRecord
	ragRecord := model.RAGRecord{
		ID:     uuid.NewString(),
		Query:  resp.Query,
		Answer: resp.Answer,
	}

	var results model.Results
	for _, result := range resp.Results {
		results = append(results, model.Result{
			Title:      result.Title,
			URL:        result.URL,
			Content:    result.Content,
			Score:      result.Score,
			RawContent: result.RawContent,
			Favicon:    result.Favicon,
		})
	}
	ragRecord.Sources = model.RagTavily
	ragRecord.Results = results

	// 4. 保存 RAG 记录到数据库
	if err := global.GetRAGRecordRepository().Create(ctx, &ragRecord); err != nil {
		logger.Error("save rag record failed", zap.Error(err))
		return nil, err
	}

	// 5. 发送 SSE 到前端
	messageID := uuid.NewString()
	global.GetBroker().PublishToSession(mapID, sse.Event{
		ID:   uuid.NewString(),
		Type: dto.MessageRagEventType,
		Data: dto.MessageRagEvent{
			NodeID:    nodeID,
			MessageID: messageID,
			RagRecord: ragRecord,
		},
	})

	// 保存消息
	global.GetMessageManager().SaveDecompositionMessage(ctx, nodeID, dto.CreateMessageRequest{
		ID:          messageID,
		MessageType: model.MsgTypeRAG,
		Role:        schema.Tool,
		Content: model.MessageContent{
			RagID: ragRecord.ID,
		},
	})

	return resp, nil
}

type ExtractFuncRequest struct {
	Query     string `json:"query"`
	URL       string `json:"url"`
	MaxTokens int    `json:"max_tokens,omitempty"`
}

func ExtractFunc(ctx context.Context, req *ExtractFuncRequest) (*ExtractResponse, error) {
	mapID := ctx.Value("mapID").(string)
	nodeID := ctx.Value("nodeID").(string)
	client := NewTavilyClient()

	extractReq := &ExtractRequest{
		Query:     req.Query,
		URL:       req.URL,
		MaxTokens: req.MaxTokens,
	}

	resp, err := client.Extract(ctx, extractReq)
	if err != nil {
		return nil, err
	}

	notice := model.Notice{
		Type:    model.NoticeTypeSuccess,
		Name:    "内容提取",
		Content: fmt.Sprintf("从 %s 提取内容成功", req.URL),
	}

	global.GetBroker().PublishToSession(mapID, sse.Event{
		ID:   uuid.NewString(),
		Type: dto.MessageNoticeEventType,
		Data: dto.MessageNoticeEvent{
			NodeID:    nodeID,
			MessageID: uuid.NewString(),
			Notice:    notice,
		},
	})

	global.GetMessageManager().SaveDecompositionMessage(ctx, nodeID, dto.CreateMessageRequest{
		ID:          uuid.NewString(),
		MessageType: model.MsgTypeNotice,
		Role:        schema.Tool,
		Content:     model.MessageContent{Notice: &notice},
	})

	return resp, nil
}

func SearchTool() (tool.InvokableTool, error) {
	t := utils.NewTool(&schema.ToolInfo{
		Name: "search",
		Desc: "使用Tavily搜索引擎进行网络检索",
		ParamsOneOf: schema.NewParamsOneOfByParams(
			map[string]*schema.ParameterInfo{
				"query": {
					Type:     schema.String,
					Desc:     "检索关键词",
					Required: true,
				},
				"search_depth": {
					Type: schema.String,
					Desc: "检索深度，可选值为 'basic' 或 'advanced'",
					Enum: []string{"basic", "advanced"},
				},
				"include_answer": {
					Type: schema.Boolean,
					Desc: "是否在结果中包含直接答案",
				},
				"include_raw_content": {
					Type: schema.Boolean,
					Desc: "是否在结果中包含原始网页内容",
				},
				"max_results": {
					Type: schema.Integer,
					Desc: "最大返回结果数量",
				},
				"include_domains": {
					Type:     schema.Array,
					ElemInfo: &schema.ParameterInfo{Type: schema.String},
					Desc:     "限定在这些域名内检索",
				},
				"exclude_domains": {
					Type:     schema.Array,
					ElemInfo: &schema.ParameterInfo{Type: schema.String},
					Desc:     "排除这些域名",
				},
				"include_favicon": {
					Type: schema.Boolean,
					Desc: "是否在结果中包含网站图标",
				},
			},
		),
	}, SearchFunc)
	return t, nil
}

func ExtractTool() (tool.InvokableTool, error) {
	t := utils.NewTool(&schema.ToolInfo{
		Name: "extract",
		Desc: "从指定URL提取与查询相关的内容",
		ParamsOneOf: schema.NewParamsOneOfByParams(
			map[string]*schema.ParameterInfo{
				"query": {
					Type:     schema.String,
					Desc:     "用于内容提取的查询或问题",
					Required: true,
				},
				"url": {
					Type:     schema.String,
					Desc:     "要提取内容的URL",
					Required: true,
				},
				"max_tokens": {
					Type: schema.Integer,
					Desc: "提取内容的最大token数量",
				},
			},
		),
	}, ExtractFunc)
	return t, nil
}

func GetAllSearchTools() ([]tool.BaseTool, error) {
	searchTool, err := SearchTool()
	if err != nil {
		return nil, err
	}
	extractTool, err := ExtractTool()
	if err != nil {
		return nil, err
	}
	return []tool.BaseTool{searchTool, extractTool}, nil
}

// genToolInfos generates tool information from tools config
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
