# useSSEConnection Hook 使用文档

## 概述

`useSSEConnection` 是一个用于管理 Server-Sent Events (SSE) 连接的 React Hook，支持连接复用、动态回调注册和自动重连等功能。

## 基本用法

### 1. 基础连接

```typescript
import { useSSEConnection } from '@/hooks/use-sse-connection';

function MyComponent() {
  const { connect, disconnect, isConnected } = useSSEConnection({
    mapID: 'your-map-id',
    callbacks: [
      {
        eventType: 'nodeCreated',
        callback: (event) => {
          console.log('节点创建:', event);
          // event 结构: { event: string, data: string, id?: string, retry?: number }
        }
      },
      {
        eventType: 'nodeUpdated',
        callback: (event) => {
          console.log('节点更新:', event);
        }
      }
    ],
    onOpen: () => {
      console.log('连接已建立');
    },
    onError: (error) => {
      console.error('连接错误:', error);
    }
  });

  return (
    <div>
      <p>连接状态: {isConnected ? '已连接' : '未连接'}</p>
      <button onClick={connect}>连接</button>
      <button onClick={disconnect}>断开</button>
    </div>
  );
}
```

### 2. 处理不同类型的事件

```typescript
const { connect, disconnect, isConnected } = useSSEConnection({
  mapID: 'your-map-id',
  callbacks: [
    {
      eventType: 'nodeCreated',
      callback: (event) => {
        const nodeData = JSON.parse(event.data);
        console.log('节点创建:', nodeData);
      }
    },
    {
      eventType: 'nodeUpdated',
      callback: (event) => {
        const updateData = JSON.parse(event.data);
        console.log('节点更新:', updateData);
      }
    },
    {
      eventType: 'error',
      callback: (event) => {
        const errorData = JSON.parse(event.data);
        console.error('业务错误:', errorData);
      }
    },
    {
      eventType: 'ping',
      callback: (event) => {
        // 心跳事件，通常不需要处理
        console.log('收到心跳');
      }
    }
  ]
});
```

### 3. 动态注册回调

```typescript
function MyComponent() {
  const { connect, disconnect, isConnected, registerEventCallbacks } = useSSEConnection({
    mapID: 'your-map-id',
    callbacks: [
      {
        eventType: 'nodeCreated',
        callback: (event) => {
          console.log('默认节点创建处理:', event);
        }
      }
    ]
  });

  const handleSpecialMode = () => {
    // 动态注册新的回调函数
    registerEventCallbacks([
      {
        eventType: 'nodeUpdated',
        callback: (event) => {
          console.log('特殊模式节点更新处理:', event);
          // 处理特殊业务逻辑
        }
      }
    ]);
  };

  return (
    <div>
      <button onClick={handleSpecialMode}>启用特殊模式</button>
    </div>
  );
}
```

## 高级用法

### 1. 使用 useSSECallbackRegistration 进行外部控制

```typescript
import { useSSECallbackRegistration } from '@/hooks/use-sse-connection';

function ExternalController() {
  const { registerEventCallbacksForMap, getConnectionStatus, disconnectMap } = useSSECallbackRegistration();

  const handleRegisterCallbacks = () => {
    registerEventCallbacksForMap('your-map-id', [
      {
        eventType: 'nodeCreated',
        callback: (event) => {
          console.log('外部注册的节点创建回调:', event);
        }
      },
      {
        eventType: 'error',
        callback: (event) => {
          console.log('外部注册的错误回调:', event);
        }
      }
    ]);
  };

  const checkStatus = () => {
    const isConnected = getConnectionStatus('your-map-id');
    console.log('连接状态:', isConnected);
  };

  const forceDisconnect = () => {
    disconnectMap('your-map-id');
  };

  return (
    <div>
      <button onClick={handleRegisterCallbacks}>注册外部回调</button>
      <button onClick={checkStatus}>检查状态</button>
      <button onClick={forceDisconnect}>强制断开</button>
    </div>
  );
}
```

### 2. 连接复用示例

```typescript
// 组件 A
function ComponentA() {
  const { isConnected } = useSSEConnection({
    mapID: 'shared-map-id',
    callbacks: [
      {
        eventType: 'nodeCreated',
        callback: (event) => {
          console.log('组件A收到节点创建消息:', event);
        }
      }
    ]
  });
  
  return <div>组件A: {isConnected ? '已连接' : '未连接'}</div>;
}

// 组件 B
function ComponentB() {
  const { isConnected } = useSSEConnection({
    mapID: 'shared-map-id', // 相同的 mapID，会复用连接
    callbacks: [
      {
        eventType: 'nodeUpdated',
        callback: (event) => {
          console.log('组件B收到节点更新消息:', event);
        }
      }
    ]
  });
  
  return <div>组件B: {isConnected ? '已连接' : '未连接'}</div>;
}
```

## API 参考

### useSSEConnection 参数

```typescript
interface SSECallbackConfig {
  eventType: SSEEventType;
  callback: SSEEventCallback;
}

interface SSEConnectionOptions {
  mapID: string;  // 必需，用于标识连接的唯一ID
  callbacks?: SSECallbackConfig[];  // 事件回调配置数组
  onOpen?: () => void;
  onError?: (error: any) => void;
}
```

### useSSEConnection 返回值

```typescript
{
  connect: () => void;           // 建立连接
  disconnect: () => void;        // 断开连接
  isConnected: boolean;          // 连接状态
  registerEventCallbacks: (configs: SSECallbackConfig[]) => void; // 动态注册事件回调
}
```

### useSSECallbackRegistration 返回值

```typescript
{
  registerEventCallbacksForMap: (mapID: string, configs: SSECallbackConfig[]) => string; // 注册回调，返回回调ID
  getConnectionStatus: (mapID: string) => boolean;  // 获取连接状态
  disconnectMap: (mapID: string) => void;           // 断开指定连接
}
```

## 特性

### 1. 连接复用
- 相同 `mapID` 的多个组件会共享同一个 SSE 连接
- 自动管理引用计数，最后一个组件卸载时才真正断开连接

### 2. 精确事件回调
- 每个事件类型只能注册一个回调函数，避免重复执行
- 支持在运行时动态注册和更新回调函数
- 每个组件可以有独立的回调处理逻辑

### 3. 自动重连
- 连接断开时自动尝试重连
- 智能的重连策略，避免频繁重连

### 4. 错误处理
- 完善的错误处理机制
- 区分客户端错误和服务器错误

### 5. 类型安全
- 完整的 TypeScript 类型定义
- 编译时类型检查

## 注意事项

1. **mapID 唯一性**: 确保每个思维导图使用唯一的 `mapID`
2. **事件类型唯一性**: 每个事件类型只能注册一个回调，重复注册会覆盖之前的回调
3. **回调函数稳定性**: 避免在回调函数中频繁创建新对象，可能影响性能
4. **错误处理**: 始终提供 `onError` 回调来处理连接错误
5. **组件卸载**: Hook 会自动处理组件卸载时的清理工作
6. **数据解析**: 回调中的 `event.data` 是字符串，需要根据业务需求进行 JSON 解析

## 迁移指南

如果你正在从旧版本的 SSE Hook 迁移，主要变化包括：

### 从 onMessage 迁移到 callbacks

**旧版本:**
```typescript
const { isConnected } = useSSEConnection({
  mapID: 'your-map-id',
  onMessage: (event) => {
    if (event.event === 'nodeCreated') {
      // 处理节点创建
    } else if (event.event === 'nodeUpdated') {
      // 处理节点更新
    }
  }
});
```

**新版本:**
```typescript
const { isConnected } = useSSEConnection({
  mapID: 'your-map-id',
  callbacks: [
    {
      eventType: 'nodeCreated',
      callback: (event) => {
        // 处理节点创建
      }
    },
    {
      eventType: 'nodeUpdated',
      callback: (event) => {
        // 处理节点更新
      }
    }
  ]
});
```

### 动态注册回调的变化

**旧版本:**
```typescript
const { registerCallbacks } = useSSEConnection({ mapID: 'your-map-id' });
registerCallbacks({
  onMessage: (event) => { /* ... */ }
});
```

**新版本:**
```typescript
const { registerEventCallbacks } = useSSEConnection({ mapID: 'your-map-id' });
registerEventCallbacks([
  {
    eventType: 'nodeCreated',
    callback: (event) => { /* ... */ }
  }
]);
```

### 主要改进

1. **类型安全**: 事件类型现在有明确的 TypeScript 类型定义
2. **精确回调**: 每个事件类型只能注册一个回调，避免重复执行
3. **更好的性能**: 减少了不必要的回调调用
4. **清晰的 API**: 事件处理逻辑更加清晰和可维护