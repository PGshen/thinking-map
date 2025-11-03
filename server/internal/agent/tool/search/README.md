# Search Tools - 搜索工具集合

## 概述

这是一个基于CloudWeGo/Eino框架封装的搜索工具集合，包含DuckDuckGo和Google两种搜索引擎，提供了全面的网络搜索功能，可以用于Agent系统中获取实时信息和资料。

### 支持的搜索引擎

1. **DuckDuckGo搜索工具** - 隐私保护的搜索引擎，无需API密钥
2. **Google搜索工具** - 基于Google Custom Search API的高质量搜索

## 功能特性

### DuckDuckGo搜索工具
- **隐私保护**: 使用DuckDuckGo搜索引擎，不追踪用户搜索行为 <mcreference link="https://www.cloudwego.io/zh/docs/eino/ecosystem_integration/tool/tool_duckduckgo_search/" index="0">0</mcreference>
- **无需API Key**: 直接使用，无需申请和配置API密钥 <mcreference link="https://www.cloudwego.io/zh/docs/eino/ecosystem_integration/tool/tool_duckduckgo_search/" index="0">0</mcreference>
- **多语言支持**: 支持中文、英文等多种语言搜索
- **灵活配置**: 支持自定义搜索参数，如结果数量、页码等
- **错误处理**: 完善的错误处理和重试机制
- **缓存支持**: 内置缓存机制，提高搜索效率

### Google搜索工具
- **高质量结果**: 基于Google Custom Search API，提供高质量的搜索结果 <mcreference link="https://www.cloudwego.io/zh/docs/eino/ecosystem_integration/tool/tool_googlesearch/" index="0">0</mcreference>
- **多语言支持**: 支持多种语言的搜索界面和结果 <mcreference link="https://www.cloudwego.io/zh/docs/eino/ecosystem_integration/tool/tool_googlesearch/" index="0">0</mcreference>
- **分页搜索**: 支持结果分页和偏移量设置 <mcreference link="https://www.cloudwego.io/zh/docs/eino/ecosystem_integration/tool/tool_googlesearch/" index="0">0</mcreference>
- **灵活配置**: 可自定义搜索参数，如结果数量、语言、偏移量等 <mcreference link="https://www.cloudwego.io/zh/docs/eino/ecosystem_integration/tool/tool_googlesearch/" index="0">0</mcreference>
- **结构化结果**: 提供标题、链接、摘要、描述等结构化信息 <mcreference link="https://www.cloudwego.io/zh/docs/eino/ecosystem_integration/tool/tool_googlesearch/" index="0">0</mcreference>

## 使用方法

### DuckDuckGo搜索工具

#### 1. 基本使用

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

#### 2. 作为Eino工具使用

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

#### 3. 获取所有搜索工具

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

### Google搜索工具

#### 1. 环境配置

使用Google搜索工具前，需要先配置API密钥：

```bash
# 设置环境变量
export GOOGLE_API_KEY="your-google-api-key"
export GOOGLE_SEARCH_ENGINE_ID="your-search-engine-id"
```

#### 2. 基本使用

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
    
    // 创建Google搜索请求
    req := &search.GoogleSearchRequest{
        Query: "artificial intelligence machine learning",
        Num:   5,
        Lang:  "en",
    }
    
    // 执行搜索
    resp, err := search.GoogleSearchFunc(ctx, req)
    if err != nil {
        log.Fatalf("Google搜索失败: %v", err)
    }
    
    // 处理搜索结果
    fmt.Printf("搜索关键词: %s\n", resp.Query)
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
```

#### 3. 作为Eino工具使用

```go
package main

import (
    "context"
    "log"
    
    "github.com/PGshen/thinking-map/server/internal/agent/tool/search"
)

func main() {
    ctx := context.Background()
    
    // 创建Google搜索工具
    tool, err := search.CreateGoogleSearchTool()
    if err != nil {
        log.Fatalf("创建Google搜索工具失败: %v", err)
    }
    
    // 准备搜索请求JSON
    reqJSON := `{"query":"CloudWeGo Eino framework","num":3,"lang":"zh-CN"}`
    
    // 执行搜索
    respJSON, err := tool.InvokableRun(ctx, reqJSON)
    if err != nil {
        log.Fatalf("执行Google搜索失败: %v", err)
    }
    
    fmt.Printf("搜索结果: %s\n", respJSON)
}
```

#### 4. 分页搜索

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
    
    // 第一页搜索
    req1 := &search.GoogleSearchRequest{
        Query:  "golang tutorial",
        Num:    3,
        Offset: 0, // 从第1个结果开始
        Lang:   "en",
    }
    
    resp1, err := search.GoogleSearchFunc(ctx, req1)
    if err != nil {
        log.Fatalf("第一页搜索失败: %v", err)
    }
    
    fmt.Println("=== 第一页搜索结果 ===")
    for i, item := range resp1.Items {
        fmt.Printf("%d. %s\n", i+1, item.Title)
    }
    
    // 第二页搜索
    req2 := &search.GoogleSearchRequest{
        Query:  "golang tutorial",
        Num:    3,
        Offset: 3, // 从第4个结果开始
        Lang:   "en",
    }
    
    resp2, err := search.GoogleSearchFunc(ctx, req2)
    if err != nil {
        log.Fatalf("第二页搜索失败: %v", err)
    }
    
    fmt.Println("=== 第二页搜索结果 ===")
    for i, item := range resp2.Items {
        fmt.Printf("%d. %s\n", i+4, item.Title) // 继续编号
    }
}
```

#### 5. 获取所有搜索工具（包括Google）

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
    
    // 获取所有搜索工具（DuckDuckGo + Google）
    tools, err := search.GetAllSearchToolsWithGoogle()
    if err != nil {
        log.Fatalf("获取搜索工具失败: %v", err)
    }
    
    fmt.Printf("可用的搜索工具数量: %d\n", len(tools))
    
    // 获取工具信息
    toolInfos, err := search.GetAllToolInfosWithGoogle(ctx)
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

### DuckDuckGo搜索工具API

#### SearchRequest

搜索请求参数结构体：

```go
type SearchRequest struct {
    Query      string `json:"query"`      // 搜索关键词（必填）
    Page       int    `json:"page"`       // 页码，默认为1
    MaxResults int    `json:"maxResults"` // 最大结果数量，默认为10，建议不超过20
}
```

#### SearchResponse

搜索响应结构体：

```go
type SearchResponse struct {
    Results []SearchResult `json:"results"` // 搜索结果列表
    Query   string         `json:"query"`   // 搜索关键词
    Page    int            `json:"page"`    // 当前页码
    Total   int            `json:"total"`   // 结果总数
}
```

#### SearchResult

单个搜索结果结构体：

```go
type SearchResult struct {
    Title       string `json:"title"`       // 标题
    Description string `json:"description"` // 描述
    Link        string `json:"link"`        // 链接
}
```

### Google搜索工具API

#### GoogleSearchRequest

Google搜索请求参数结构体：

```go
type GoogleSearchRequest struct {
    Query  string `json:"query"`  // 搜索关键词（必填）
    Num    int    `json:"num"`    // 返回结果数量，默认为5，最大为10
    Offset int    `json:"offset"` // 结果起始位置，用于分页，默认为0
    Lang   string `json:"lang"`   // 搜索语言，如zh-CN（中文）、en（英文），默认为zh-CN
}
```

#### GoogleSearchResponse

Google搜索响应结构体：

```go
type GoogleSearchResponse struct {
    Query string               `json:"query"` // 搜索关键词
    Items []GoogleSearchResult `json:"items"` // 搜索结果列表
    Total int                  `json:"total"` // 结果总数
}
```

#### GoogleSearchResult

Google搜索单个结果结构体：

```go
type GoogleSearchResult struct {
    Title   string `json:"title"`   // 标题
    Link    string `json:"link"`    // 链接
    Snippet string `json:"snippet"` // 摘要
    Desc    string `json:"desc"`    // 描述
}
```

## 配置说明

### DuckDuckGo搜索工具配置

工具内部使用以下配置：

- **搜索区域**: 全球搜索 (RegionWT)
- **安全搜索**: 关闭 (SafeSearchOff)
- **时间范围**: 全部时间 (TimeRangeAll)
- **超时时间**: 30秒
- **缓存**: 启用
- **最大重试次数**: 3次

### Google搜索工具配置

使用Google搜索工具需要配置以下环境变量：

- **GOOGLE_API_KEY**: Google API密钥（必需）
- **GOOGLE_SEARCH_ENGINE_ID**: Google自定义搜索引擎ID（必需）

工具内部配置：

- **默认结果数量**: 5个
- **最大结果数量**: 10个
- **默认语言**: zh-CN（中文）
- **支持分页**: 是，通过Offset参数
- **API基础URL**: https://customsearch.googleapis.com（可自定义）

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

### 通用注意事项

1. **网络依赖**: 搜索功能需要网络连接，请确保网络环境正常
2. **结果数量**: 建议合理控制搜索结果数量，以获得最佳性能
3. **错误处理**: 请妥善处理网络错误和搜索失败的情况

### DuckDuckGo搜索工具注意事项

1. **速率限制**: DuckDuckGo可能有速率限制，建议合理控制搜索频率
2. **结果数量**: 建议MaxResults不超过20，以获得最佳性能
3. **缓存机制**: 工具内置缓存，相同查询可能返回缓存结果

### Google搜索工具注意事项

1. **API配额**: Google Custom Search API有每日免费配额限制（100次/天）
2. **API密钥安全**: 请妥善保管API密钥，不要在代码中硬编码
3. **搜索引擎配置**: 需要在Google Cloud Console中创建自定义搜索引擎
4. **结果数量限制**: 单次搜索最多返回10个结果
5. **分页限制**: Google API对分页有限制，建议合理使用Offset参数

## 集成到Agent系统

这些搜索工具可以轻松集成到基于Eino框架的Agent系统中：

### 使用所有搜索工具

```go
// 在Agent初始化时添加所有搜索工具（DuckDuckGo + Google）
searchTools, err := search.GetAllSearchToolsWithGoogle()
if err != nil {
    return err
}

// 将搜索工具添加到Agent的工具列表中
agent.AddTools(searchTools...)
```

### 单独使用特定搜索工具

```go
// 只使用DuckDuckGo搜索工具
duckduckgoTool, err := search.CreateDuckDuckGoSearchTool()
if err != nil {
    return err
}
agent.AddTool(duckduckgoTool)

// 只使用Google搜索工具
googleTool, err := search.CreateGoogleSearchTool()
if err != nil {
    return err
}
agent.AddTool(googleTool)
```

## 扩展功能

未来可以考虑添加以下功能：

### 搜索引擎扩展
- 支持更多搜索引擎（Bing、百度、搜狗等）
- 添加搜索引擎选择策略
- 实现搜索结果聚合和去重

### 功能增强
- 添加搜索结果过滤和排序
- 支持图片和视频搜索
- 添加搜索历史记录
- 实现搜索结果的智能摘要
- 支持实时搜索和搜索建议

### 性能优化
- 实现分布式缓存
- 添加搜索结果预加载
- 优化并发搜索性能

## 相关文档

- [CloudWeGo Eino官方文档](https://www.cloudwego.io/zh/docs/eino/)
- [DuckDuckGo搜索工具文档](https://www.cloudwego.io/zh/docs/eino/ecosystem_integration/tool/tool_duckduckgo_search/)
- [Google搜索工具文档](https://www.cloudwego.io/zh/docs/eino/ecosystem_integration/tool/tool_googlesearch/)
- [Eino工具开发指南](https://www.cloudwego.io/zh/docs/eino/ecosystem_integration/tool/)