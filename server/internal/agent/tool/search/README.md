# Search Tool - DuckDuckGo搜索工具

## 概述

这是一个基于CloudWeGo/Eino框架封装的DuckDuckGo搜索工具，提供了网络搜索功能，可以用于Agent系统中获取实时信息和资料。

## 功能特性

- **隐私保护**: 使用DuckDuckGo搜索引擎，不追踪用户搜索行为 <mcreference link="https://www.cloudwego.io/zh/docs/eino/ecosystem_integration/tool/tool_duckduckgo_search/" index="0">0</mcreference>
- **无需API Key**: 直接使用，无需申请和配置API密钥 <mcreference link="https://www.cloudwego.io/zh/docs/eino/ecosystem_integration/tool/tool_duckduckgo_search/" index="0">0</mcreference>
- **多语言支持**: 支持中文、英文等多种语言搜索
- **灵活配置**: 支持自定义搜索参数，如结果数量、页码等
- **错误处理**: 完善的错误处理和重试机制
- **缓存支持**: 内置缓存机制，提高搜索效率

## 使用方法

### 1. 基本使用

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/PGshen/thinking-map/server/internal/agent/tool/search"
)

func main() {
    ctx := context.Background()
    
    // 创建搜索请求
    req := &search.SearchRequest{
        Query:      "Go programming language",
        Page:       1,
        MaxResults: 5,
    }
    
    // 执行搜索
    resp, err := search.DuckDuckGoSearchFunc(ctx, req)
    if err != nil {
        log.Fatalf("搜索失败: %v", err)
    }
    
    // 处理搜索结果
    fmt.Printf("搜索关键词: %s\n", resp.Query)
    fmt.Printf("找到 %d 个结果:\n", resp.Total)
    
    for i, result := range resp.Results {
        fmt.Printf("%d. %s\n", i+1, result.Title)
        fmt.Printf("   链接: %s\n", result.Link)
        fmt.Printf("   描述: %s\n\n", result.Description)
    }
}
```

### 2. 作为Eino工具使用

```go
package main

import (
    "context"
    "log"
    
    "github.com/PGshen/thinking-map/server/internal/agent/tool/search"
)

func main() {
    ctx := context.Background()
    
    // 创建搜索工具
    tool, err := search.CreateDuckDuckGoSearchTool()
    if err != nil {
        log.Fatalf("创建搜索工具失败: %v", err)
    }
    
    // 准备搜索请求JSON
    reqJSON := `{"query":"人工智能发展趋势","page":1,"maxResults":10}`
    
    // 执行搜索
    respJSON, err := tool.InvokableRun(ctx, reqJSON)
    if err != nil {
        log.Fatalf("执行搜索失败: %v", err)
    }
    
    fmt.Printf("搜索结果: %s\n", respJSON)
}
```

### 3. 获取所有搜索工具

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/PGshen/thinking-map/server/internal/agent/tool/search"
)

func main() {
    ctx := context.Background()
    
    // 获取所有搜索工具
    tools, err := search.GetAllSearchTools()
    if err != nil {
        log.Fatalf("获取搜索工具失败: %v", err)
    }
    
    fmt.Printf("可用的搜索工具数量: %d\n", len(tools))
    
    // 获取工具信息
    toolInfos, err := search.GetAllToolInfos(ctx)
    if err != nil {
        log.Fatalf("获取工具信息失败: %v", err)
    }
    
    for _, info := range toolInfos {
        fmt.Printf("工具名称: %s\n", info.Name)
        fmt.Printf("工具描述: %s\n\n", info.Desc)
    }
}
```

## API参考

### SearchRequest

搜索请求参数结构体：

```go
type SearchRequest struct {
    Query      string `json:"query"`      // 搜索关键词（必填）
    Page       int    `json:"page"`       // 页码，默认为1
    MaxResults int    `json:"maxResults"` // 最大结果数量，默认为10，建议不超过20
}
```

### SearchResponse

搜索响应结构体：

```go
type SearchResponse struct {
    Results []SearchResult `json:"results"` // 搜索结果列表
    Query   string         `json:"query"`   // 搜索关键词
    Page    int            `json:"page"`    // 当前页码
    Total   int            `json:"total"`   // 结果总数
}
```

### SearchResult

单个搜索结果结构体：

```go
type SearchResult struct {
    Title       string `json:"title"`       // 标题
    Description string `json:"description"` // 描述
    Link        string `json:"link"`        // 链接
}
```

## 配置说明

工具内部使用以下配置：

- **搜索区域**: 全球搜索 (RegionWT)
- **安全搜索**: 关闭 (SafeSearchOff)
- **时间范围**: 全部时间 (TimeRangeAll)
- **超时时间**: 30秒
- **缓存**: 启用
- **最大重试次数**: 3次

## 测试

运行单元测试：

```bash
cd server
go test ./internal/agent/tool/search -v
```

运行集成测试：

```bash
cd server
go test ./internal/agent/tool/search -v -run TestSearchToolIntegration
```

运行基准测试：

```bash
cd server
go test ./internal/agent/tool/search -bench=.
```

## 注意事项

1. **网络依赖**: 搜索功能需要网络连接，请确保网络环境正常
2. **速率限制**: DuckDuckGo可能有速率限制，建议合理控制搜索频率
3. **结果数量**: 建议MaxResults不超过20，以获得最佳性能
4. **错误处理**: 请妥善处理网络错误和搜索失败的情况
5. **缓存机制**: 工具内置缓存，相同查询可能返回缓存结果

## 集成到Agent系统

这个搜索工具可以轻松集成到基于Eino框架的Agent系统中：

```go
// 在Agent初始化时添加搜索工具
searchTools, err := search.GetAllSearchTools()
if err != nil {
    return err
}

// 将搜索工具添加到Agent的工具列表中
agent.AddTools(searchTools...)
```

## 扩展功能

未来可以考虑添加以下功能：

- 支持更多搜索引擎（Google、Bing等）
- 添加搜索结果过滤和排序
- 支持图片和视频搜索
- 添加搜索历史记录
- 实现搜索结果的智能摘要

## 相关文档

- [CloudWeGo Eino官方文档](https://www.cloudwego.io/zh/docs/eino/)
- [DuckDuckGo搜索工具文档](https://www.cloudwego.io/zh/docs/eino/ecosystem_integration/tool/tool_duckduckgo_search/)
- [Eino工具开发指南](https://www.cloudwego.io/zh/docs/eino/ecosystem_integration/tool/)