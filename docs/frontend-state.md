# ThinkingMap 前端：状态管理（Zustand）

## 1. 状态划分
- **全局状态**：用户信息、当前问题、节点树、SSE 连接状态、全局 loading/error。
- **局部状态**：表单输入、面板 Tab、节点选中、对话消息等。

## 2. Store 拆分与设计原则
- 每个业务模块可单独维护 store，避免全局状态臃肿。
- 支持中大型项目的状态拆分与组合。
- 结合 React Context 提供跨组件访问。
- 所有状态、action、selector 定义 TypeScript 类型。
- 只存 UI 相关状态，后端数据通过 API/SSE 获取，store 只存 UI 需要的部分。

## 3. Store 结构与示例
```ts
// store/mapStore.ts
interface MapStore {
  nodes: Node[];
  edges: Edge[];
  selectedNodeId: string | null;
  setNodes: (nodes: Node[]) => void;
  setEdges: (edges: Edge[]) => void;
  selectNode: (id: string | null) => void;
  // ...更多action
}

// store/panelStore.ts
interface PanelStore {
  isOpen: boolean;
  activeTab: 'info' | 'decompose' | 'conclusion';
  formState: InfoFormState;
  chatMessages: ChatMessage[];
  setActiveTab: (tab: PanelTab) => void;
  // ...
}

// store/globalStore.ts
interface GlobalStore {
  user: User | null;
  mapId: string | null;
  loading: boolean;
  error: string | null;
  setUser: (user: User | null) => void;
  // ...
}
```

## 4. 与 API/SSE 的集成
- API 请求/响应通过 action 更新 store 状态（如节点树、用户信息等）。
- SSE 事件直接写入/更新相关 store（如节点变更、消息推送、连接状态等）。
- 支持 selector 订阅局部状态，避免无关组件重渲染。

## 5. 持久化与中间件
- 推荐使用 zustand/middleware 的 persist、immer、devtools 等。
- 关键状态（如用户登录态、主题）可本地持久化，其他状态随页面刷新重置。
- 开发环境启用 devtools，便于调试。

## 6. 性能优化
- 拆分 store，按需订阅，减少全局重渲染。
- 使用 selector 精细订阅，避免无关组件更新。
- 大数据（如节点树）可用 shallow 比较或 memo 优化。
- 只在必要时更新状态，避免高频写入。

## 7. 类型定义与代码结构
- `/store` 目录下按模块拆分 store 文件。
- `/types` 目录定义所有状态、action、数据结构类型。
- store 内部 action、selector、初始状态全部类型化。

## 8. 最佳实践
- 只在组件树顶层注入 Provider（如有 context）。
- 组件内通过 hook 访问和操作 store。
- 复杂业务逻辑封装为 action，组件只负责 UI。
- 状态变更与后端同步解耦，便于测试和维护。

## 9. 示例伪代码
```ts
// store/mapStore.ts
import { create } from 'zustand';
import { persist, devtools } from 'zustand/middleware';

export const useMapStore = create<MapStore>()(
  devtools(
    persist(
      (set, get) => ({
        nodes: [],
        edges: [],
        selectedNodeId: null,
        setNodes: (nodes) => set({ nodes }),
        setEdges: (edges) => set({ edges }),
        selectNode: (id) => set({ selectedNodeId: id }),
        // ...
      }),
      { name: 'map-store' }
    )
  )
);
```

> 状态管理方案应兼顾类型安全、性能和可维护性。 