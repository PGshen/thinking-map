# MCP - Tavily 接入

本目录提供基于 CloudWeGo Eino 的 MCP（Model Context Protocol）工具接入，实现与 Tavily MCP Server 的连接，并将其提供的工具（search、extract、map、crawl）注入到 Agent 中使用。

## 参考文档

- Tavily MCP 文档（远程服务器 URL、客户端配置）：https://docs.tavily.com/documentation/mcp
- CloudWeGo Eino MCP 封装使用说明：https://www.cloudwego.io/zh/docs/eino/ecosystem_integration/tool/tool_mcp/

## 可用工具

默认加载以下工具（可通过配置筛选）：

- `tavily-search`：实时网络搜索
- `tavily-extract`：网页数据抽取
- `tavily-map`：网站结构化映射
- `tavily-crawl`：网页爬取

## 环境变量与配置

- `TAVILY_API_KEY`：必填，Tavily API Key（通常以 `tvly-` 开头）
- `mcp.tavily.api_key`：viper 配置项，优先于环境变量
- `mcp.tavily.base_url`：Tavily MCP Server 基础 URL，默认 `https://mcp.tavily.com/mcp/`

注：远程 MCP Server 使用 SSE 连接，API Key 采用查询参数 `tavilyApiKey` 传递，如：

```
https://mcp.tavily.com/mcp/?tavilyApiKey=tvly-xxxx
```

## 使用示例

```go
ctx := context.Background()

// 加载 Tavily MCP 工具（读取 env/viper 配置）
tools, err := mcp.GetTavilyTools(ctx, nil)
if err != nil {
    log.Fatalf("加载 Tavily MCP 工具失败: %v", err)
}

// 获取工具信息
infos, err := mcp.GetTavilyToolInfos(ctx, nil)
if err != nil {
    log.Fatalf("获取工具信息失败: %v", err)
}
for _, info := range infos {
    fmt.Printf("工具: %s\n描述: %s\n\n", info.Name, info.Desc)
}

// 注入到 Agent
// agent.AddTools(tools...)
```

## 注意事项

- 需要有效的 Tavily API Key，且网络可访问 Tavily MCP Server
- SSE 客户端需要显式启动与初始化握手（代码中已处理）
- 如需限制工具范围，可通过 `TavilyMCPConfig.ToolNames` 指定

## 故障排查

1. 连接失败：检查本机 Node/网络，无需本地安装 Tavily，只需远程 URL 可用
2. API Key 问题：确保以 `tvly-` 前缀，且在环境变量或 viper 配置中正确设置
3. 工具列表为空：稍后重试，或确认 `ToolNames` 未筛掉全部工具

---

本接入遵循 Eino 官方 MCP 封装流程，通过 mark3labs/mcp-go 的 SSE 客户端与 Tavily 远程 MCP Server 建立连接，并将工具以 `tool.BaseTool` 形式供 Agent 使用。