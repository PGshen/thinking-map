# SSE 实时连接功能

本文档说明了工作区中SSE（Server-Sent Events）实时连接功能的实现和使用方法。

## 功能概述

当用户进入工作区后，系统会自动建立SSE长连接，用于接收后端推送的实时事件，包括：

- **节点创建事件** (`nodeCreated`): 当后端创建新节点时推送
- **节点更新事件** (`nodeUpdated`): 当节点状态或内容更新时推送
- **思考进度事件** (`thinkingProgress`): 当AI思考过程中推送进度信息
- **错误事件** (`error`): 当发生错误时推送错误信息

## 技术实现

### 1. 类型定义

在 `src/types/sse.ts` 中定义了与后端对应的事件类型：

```typescript
export interface NodeCreatedEvent {
  nodeID: string;
  parentID: string;
  nodeType: string;
  question: string;
  target: string;
  position: Position;
  timestamp: string;
}

export interface NodeUpdatedEvent {
  nodeID: string;
  mode: string; // 更新模式：replace/append
  updates: Record<string, any>;
  timestamp: string;
}
```

### 2. SSE连接Hook

`src/hooks/use-sse-connection.ts` 提供了SSE连接的核心功能：

```typescript
export function useSSEConnection({
  mapID,
  onNodeCreated,
  onNodeUpdated,
  onThinkingProgress,
  onError,
  onConnectionEstablished,
}: SSEConnectionOptions)
```

**主要特性：**
- 自动连接管理（连接、断开、重连）
- 事件类型解析和分发
- 错误处理和重连机制
- 通过URL参数传递认证token

### 3. 可视化区域集成

在 `VisualizationArea` 组件中集成了SSE功能：

```typescript
// 建立SSE连接
const { isConnected } = useSSEConnection({
  mapID,
  onNodeCreated: handleNodeCreated,
  onNodeUpdated: handleNodeUpdated,
  onThinkingProgress: handleThinkingProgress,
  onError: handleSSEError,
  onConnectionEstablished: handleConnectionEstablished,
});
```

## 事件处理逻辑

### 节点创建事件处理

当接收到 `nodeCreated` 事件时：
1. 创建新的ReactFlow节点对象
2. 添加到画布中
3. 如果有父节点，创建连接边
4. 同步到Zustand store

### 节点更新事件处理

当接收到 `nodeUpdated` 事件时：
1. 根据更新模式（replace/append）处理数据
2. 更新对应节点的数据
3. 同步到store

### 思考进度事件处理

当接收到 `thinkingProgress` 事件时：
1. 更新节点状态为 `running`
2. 在节点metadata中存储进度信息
3. 可用于显示进度条或状态指示器

## 使用方法

### 在工作区页面中使用

SSE连接已经自动集成到工作区中，无需额外配置：

```typescript
// /app/(workspace)/map/[id]/page.tsx
export default function TestWorkspacePage() {
  const { id: mapID } = useParams<{ id: string }>();
  
  return (
    <WorkspaceLayout mapID={mapID} />
  );
}
```

### 自定义事件处理

如果需要自定义事件处理逻辑，可以直接使用 `useSSEConnection` hook：

```typescript
const { isConnected } = useSSEConnection({
  mapID: 'your-map-id',
  onNodeCreated: (event) => {
    console.log('New node created:', event);
    // 自定义处理逻辑
  },
  onNodeUpdated: (event) => {
    console.log('Node updated:', event);
    // 自定义处理逻辑
  },
});
```

## 认证机制

由于EventSource不支持自定义headers，认证token通过URL参数传递：

```
GET /api/v1/sse/connect?map_id={mapID}&token={token}
```

后端需要支持从query参数中获取和验证token。

## 错误处理和重连

- **连接错误**: 自动在3秒后尝试重连
- **解析错误**: 记录错误日志，不影响其他事件处理
- **认证失败**: 需要用户重新登录

## 调试和监控

所有SSE事件都会在控制台输出日志，便于开发调试：

```
SSE connection established: {...}
Received node created event: {...}
Received node updated event: {...}
```

## 注意事项

1. **浏览器兼容性**: EventSource在所有现代浏览器中都有良好支持
2. **连接限制**: 浏览器对同一域名的SSE连接数有限制（通常6个）
3. **内存管理**: 组件卸载时会自动断开连接，避免内存泄漏
4. **网络状态**: 网络断开时会自动重连，恢复后继续接收事件

## 后端要求

后端需要实现以下SSE端点和事件格式：

- **连接端点**: `GET /api/v1/sse/connect?map_id={mapID}&token={token}`
- **事件格式**: 符合 `src/types/sse.ts` 中定义的接口
- **认证**: 支持通过URL参数验证token
- **错误处理**: 发送标准化的错误事件