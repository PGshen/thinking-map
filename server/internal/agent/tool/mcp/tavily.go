package mcp

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	mcpclient "github.com/mark3labs/mcp-go/client"
	mcpproto "github.com/mark3labs/mcp-go/mcp"

	einomcp "github.com/cloudwego/eino-ext/components/tool/mcp"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/spf13/viper"
)

// TavilyMCPConfig 定义 Tavily MCP 工具接入配置
type TavilyMCPConfig struct {
	// Tavily API Key（必须，格式通常为 tvly-xxxx）
	APIKey string
	// Tavily MCP Server 基础地址，默认：https://mcp.tavily.com/mcp/
	BaseURL string
	// 需要加载的工具名称列表，可选：tavily-search、tavily-extract、tavily-map、tavily-crawl
	ToolNames []string
	// 初始化超时时间
	InitTimeout time.Duration
}

// DefaultTavilyMCPConfig 读取配置（优先 viper，其次环境变量），生成默认配置
func DefaultTavilyMCPConfig() *TavilyMCPConfig {
	apiKey := viper.GetString("mcp.tavily.api_key")
	if apiKey == "" {
		apiKey = os.Getenv("TAVILY_API_KEY")
	}

	baseURL := viper.GetString("mcp.tavily.base_url")
	if baseURL == "" {
		baseURL = "https://mcp.tavily.com/mcp/"
	}

	// 默认加载 Tavily 官方提供的四个工具
	toolNames := []string{"tavily-search", "tavily-extract", "tavily-map", "tavily-crawl"}

	return &TavilyMCPConfig{
		APIKey:      apiKey,
		BaseURL:     baseURL,
		ToolNames:   toolNames,
		InitTimeout: 20 * time.Second,
	}
}

// GetTavilyTools 创建 MCP Client 并获取 Tavily 的工具集合，返回可直接用于 Eino Agent 的工具列表
func GetTavilyTools(ctx context.Context, cfg *TavilyMCPConfig) ([]tool.BaseTool, error) {
	if cfg == nil {
		cfg = DefaultTavilyMCPConfig()
	}

	if cfg.APIKey == "" {
		return nil, fmt.Errorf("缺少 Tavily API Key：请在环境变量 TAVILY_API_KEY 或 viper 配置 mcp.tavily.api_key 中设置")
	}

	// 构造带 apiKey 的远程 MCP URL
	// 例如：https://mcp.tavily.com/mcp/?tavilyApiKey=tvly-xxxx
	fullURL := buildTavilyURL(cfg.BaseURL, cfg.APIKey)

	// 创建 SSE MCP 客户端并启动异步通信
	cli, err := mcpclient.NewSSEMCPClient(fullURL)
	if err != nil {
		return nil, fmt.Errorf("创建 Tavily MCP SSE 客户端失败: %w", err)
	}

	// SSE 需要显式 Start
	startCtx, cancel := context.WithTimeout(ctx, cfg.InitTimeout)
	defer cancel()
	if err = cli.Start(startCtx); err != nil {
		return nil, fmt.Errorf("启动 Tavily MCP 客户端失败: %w", err)
	}

	// 执行 MCP 初始化握手
	initReq := mcpproto.InitializeRequest{}
	initReq.Params.ProtocolVersion = mcpproto.LATEST_PROTOCOL_VERSION
	initReq.Params.ClientInfo = mcpproto.Implementation{ // 客户端信息
		Name:    "thinking-map-client",
		Version: "1.0.0",
	}

	initCtx, initCancel := context.WithTimeout(ctx, cfg.InitTimeout)
	defer initCancel()
	if _, err = cli.Initialize(initCtx, initReq); err != nil {
		return nil, fmt.Errorf("初始化 Tavily MCP 客户端失败: %w", err)
	}

	// 通过 CloudWeGo Eino 封装拉取 MCP Tools
	tools, err := einomcp.GetTools(ctx, &einomcp.Config{
		Cli:          cli,
		ToolNameList: cfg.ToolNames,
	})
	if err != nil {
		return nil, fmt.Errorf("获取 Tavily MCP 工具失败: %w", err)
	}
	return tools, nil
}

// GetTavilyToolInfos 获取工具的元信息（名称、描述、参数等）
func GetTavilyToolInfos(ctx context.Context, cfg *TavilyMCPConfig) ([]*schema.ToolInfo, error) {
	tools, err := GetTavilyTools(ctx, cfg)
	if err != nil {
		return nil, err
	}
	infos := make([]*schema.ToolInfo, 0, len(tools))
	for _, t := range tools {
		info, err := t.Info(ctx)
		if err != nil {
			return nil, fmt.Errorf("读取工具信息失败: %w", err)
		}
		infos = append(infos, info)
	}
	return infos, nil
}

// buildTavilyURL 规范化基础地址并拼接 tavilyApiKey 查询参数
func buildTavilyURL(base, apiKey string) string {
	b := strings.TrimSpace(base)
	// 去掉可能的多余问号
	b = strings.TrimSuffix(b, "?")
	// 确保以 / 结尾的 base 也能正确拼接
	if !strings.Contains(b, "?") {
		return fmt.Sprintf("%s?tavilyApiKey=%s", strings.TrimRight(b, "/"), apiKey)
	}
	// 已包含查询参数的情况，追加
	return fmt.Sprintf("%s&tavilyApiKey=%s", b, apiKey)
}

// Example:
//  ctx := context.Background()
//  tools, err := GetTavilyTools(ctx, nil)
//  if err != nil { /* handle */ }
//  // 将 tools 添加到 Agent：agent.AddTools(tools...)
