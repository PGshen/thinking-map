# 节点操作工具 (Node Operation Tools)

本包提供了基于 CloudWeGo/Eino 框架的思维导图节点操作工具，包括创建、更新和删除节点的功能。

## 功能特性

- **创建节点** (`createNode`): 在思维导图中创建新的节点
- **更新节点** (`updateNode`): 更新现有节点的信息和位置
- **删除节点** (`deleteNode`): 删除指定的节点

## 工具列表

### 1. CreateNodeTool - 创建节点工具

**功能**: 在思维导图中创建新的节点

**参数**:
- `mapID` (必需): 思维导图ID
- `parentID` (可选): 父节点ID，根节点可为空
- `nodeType` (必需): 节点类型，如root、analysis、conclusion等
- `question` (必需): 节点问题描述
- `target` (可选): 节点目标描述
- `x` (必需): 节点X坐标位置
- `y` (必需): 节点Y坐标位置

**返回**: `dto.NodeResponse` - 创建的节点信息

### 2. UpdateNodeTool - 更新节点工具

**功能**: 更新现有思维导图节点的信息

**参数**:
- `nodeID` (必需): 要更新的节点ID
- `question` (可选): 更新的节点问题描述
- `target` (可选): 更新的节点目标描述
- `x` (可选): 更新的节点X坐标位置
- `y` (可选): 更新的节点Y坐标位置

**返回**: `dto.NodeResponse` - 更新后的节点信息

### 3. DeleteNodeTool - 删除节点工具

**功能**: 删除指定的思维导图节点

**参数**:
- `nodeID` (必需): 要删除的节点ID

**返回**: `DeleteNodeResponse` - 删除操作结果

## 使用示例

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/PGshen/thinking-map/server/internal/agent/tool/node"
)

func main() {
    // 获取所有节点操作工具
    tools, err := node.GetAllNodeTools()
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("成功创建 %d 个节点操作工具\n", len(tools))
    
    // 单独创建工具
    createTool, err := node.CreateNodeTool()
    if err != nil {
        log.Fatal(err)
    }
    
    updateTool, err := node.UpdateNodeTool()
    if err != nil {
        log.Fatal(err)
    }
    
    deleteTool, err := node.DeleteNodeTool()
    if err != nil {
        log.Fatal(err)
    }
    
    // 工具现在可以在 Eino 框架中使用
    fmt.Println("节点操作工具已准备就绪")
}
```

## 架构设计

### 依赖关系

```
node/operator.go
├── global.NodeOperator (业务逻辑层)
├── model/dto (数据传输对象)
├── cloudwego/eino (工具框架)
└── model.Position (位置模型)
```

### 工具创建流程

1. **参数定义**: 使用 `schema.ParameterInfo` 定义工具参数
2. **函数实现**: 实现具体的业务逻辑函数
3. **工具封装**: 使用 `utils.NewTool` 创建 Eino 工具
4. **错误处理**: 统一的错误处理和返回

## 注意事项

1. **坐标系统**: 节点位置使用浮点数坐标系统 (X, Y)
2. **节点类型**: 支持多种节点类型，具体类型由业务逻辑定义
3. **父子关系**: 创建节点时需要指定父节点，根节点的父节点为空
4. **错误处理**: 所有操作都包含完整的错误处理机制
5. **事务安全**: 底层使用 `global.NodeOperator` 确保数据一致性

## 测试

运行测试:

```bash
go test ./internal/agent/tool/node/ -v
```

所有工具都包含完整的单元测试，确保功能正确性。