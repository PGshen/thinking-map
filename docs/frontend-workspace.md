<!--
 * @Date: 2025-01-27
 * @LastEditors: AI Assistant
 * @LastEditTime: 2025-01-27
 * @FilePath: /thinking-map/docs/frontend-workspace.md
-->
# ThinkingMap 前端：工作区系统

## 1. 总体布局与架构

工作区是 ThinkingMap 的核心界面，为用户提供完整的思维导图编辑和管理体验。整体采用三层布局结构：

- **顶部固定栏**：全局导航与操作控制
- **主体可视化区域**：节点图谱的展示与交互
- **右侧操作面板**：节点详情编辑与管理（按需展开）

### 1.1 响应式设计原则
- 支持不同屏幕尺寸的自适应布局
- 最小支持宽度 1024px，推荐 1440px 以上
- 不支持移动端

## 2. 顶部固定栏设计

### 2.1 布局结构
```
[退出按钮] [任务名称] ........................ [设置按钮]
```

### 2.2 组件详细设计

#### 左侧区域
- **退出按钮**：
  - 图标：返回/关闭图标（如 ArrowLeft 或 X）
  - 功能：退出当前工作区，返回任务列表或首页
  - 交互：点击前弹出确认对话框（如有未保存内容）
  - 样式：圆形按钮，悬浮时高亮

- **任务名称**：
  - 显示当前思维导图的标题
  - 支持点击编辑（内联编辑或弹窗编辑）
  - 最大显示长度限制，超出显示省略号
  - 字体：中等粗细，突出显示

#### 右侧区域
- **设置按钮**：
  - 图标：齿轮或三点菜单图标
  - 功能：打开设置下拉菜单或设置面板
  - 菜单项可能包括：
    - 导出思维导图
    - 布局设置
    - 显示选项
    - 帮助文档

### 2.3 样式规范
- 高度：固定 64px
- 背景：白色或浅色，带底部边框阴影
- 内边距：左右各 24px
- 字体：任务名称 16px，按钮图标 20px
- 颜色：遵循设计系统主色调

### 2.4 交互行为
- 固定定位，始终保持在视窗顶部
- 滚动时保持可见，不受可视化区域缩放影响
- 支持键盘快捷键（如 Ctrl+S 保存，Esc 退出等）

## 3. 节点可视化区域

### 3.1 区域定义
- 占据顶部固定栏以下的全部空间
- 右侧面板以抽屉形式展开时，不会影响可视化区域宽度

### 3.2 核心功能
基于 ReactFlow 实现的思维导图可视化，详细功能参考 <mcfile name="frontend-visual.md" path="/Users/shen/Me/Code/agent/thinking-map/docs/frontend-visual.md"></mcfile>：

- **节点渲染**：自定义节点样式，支持多种类型和状态
- **边连接**：表达节点间的层级和依赖关系
- **交互操作**：拖拽、缩放、选择、连接等
- **布局算法**：分层树状布局，兼顾层级与依赖关系
- **性能优化**：虚拟化渲染，支持大量节点

### 3.3 与面板的联动
- 节点双击：展开右侧操作面板，显示节点详情
- 节点选中：高亮显示，面板内容同步更新
- 面板操作：实时反映到可视化区域的节点状态
- 布局调整：面板展开/收起时，画布自动调整尺寸

### 3.4 状态管理
- 节点数据：存储于 Zustand store，支持实时更新
- 选中状态：局部状态管理，与面板联动
- SSE 同步：后端推送的变更自动更新画布

## 4. 右侧操作面板

### 4.1 展开触发
- **主要触发方式**：双击节点
- **其他触发方式**：
  - 节点右键菜单选择"编辑详情"
  - 快捷键（如 Enter 键）
  - 顶部工具栏的"节点详情"按钮

### 4.2 面板特性
- **抽屉式设计**：从右侧滑入，带遮罩层
- **宽度控制**：默认 35% 屏幕宽度，可拖拽调节（25%-50%）
- **折叠功能**：支持临时收起，保持选中状态
- **响应式**：小屏幕时可全屏覆盖

### 4.3 面板内容
详细功能参考 <mcfile name="frontend-panel.md" path="/Users/shen/Me/Code/agent/thinking-map/docs/frontend-panel.md"></mcfile>：

- **信息 Tab**：节点基础信息编辑
- **拆解 Tab**：问题拆解的聊天式交互
- **结论 Tab**：结论生成与确认
- **依赖管理**：显示和管理节点依赖关系

### 4.4 关闭行为
- 点击遮罩层关闭
- 按 Esc 键关闭
- 面板内关闭按钮
- 切换到其他节点时自动切换内容

## 5. 整体交互流程

### 5.1 工作区初始化
1. 加载任务基础信息，显示在顶部栏
2. 从后端获取节点树数据
3. 渲染可视化画布，应用默认布局
4. 建立 SSE 连接，监听实时更新

### 5.2 节点操作流程
1. **浏览模式**：用户查看整体思维导图结构
2. **选择节点**：单击节点进行选中，显示操作按钮
3. **编辑模式**：双击节点展开操作面板
4. **详情编辑**：在面板中编辑节点信息、执行拆解或生成结论
5. **保存同步**：操作结果实时同步到可视化区域

### 5.3 多节点协作
- 支持同时选中多个节点进行批量操作
- 面板显示多选状态下的通用操作
- 实时显示其他用户的操作状态（如有协作功能）

## 6. 状态管理架构

### 6.1 全局状态（Zustand）
```typescript
interface WorkspaceStore {
  // 任务信息
  taskInfo: TaskInfo
  
  // 节点数据
  nodes: ThinkingNode[]
  edges: NodeEdge[]
  
  // 选中状态
  selectedNodeIDs: string[]
  
  // 面板状态
  panelOpen: boolean
  panelWidth: number
  activeNodeID: string | null
  
  // 操作方法
  actions: {
    selectNode: (nodeID: string) => void
    openPanel: (nodeID: string) => void
    closePanel: () => void
    updateNode: (nodeID: string, data: Partial<ThinkingNode>) => void
    // ...
  }
}
```

### 6.2 组件状态
- 局部 UI 状态（如动画、临时输入等）使用 React useState
- 表单状态使用 react-hook-form 管理
- 复杂交互状态可使用 useReducer

## 7. 代码结构建议

### 7.1 目录结构
```
src/features/workspace/
├── components/
│   ├── WorkspaceLayout.tsx          # 工作区整体布局
│   ├── TopBar/
│   │   ├── TopBar.tsx               # 顶部固定栏
│   │   ├── ExitButton.tsx           # 退出按钮
│   │   ├── TaskTitle.tsx            # 任务标题
│   │   └── SettingsButton.tsx       # 设置按钮
│   ├── VisualizationArea/
│   │   ├── MapCanvas.tsx            # 可视化画布
│   │   ├── CustomNode.tsx           # 自定义节点
│   │   ├── CustomEdge.tsx           # 自定义边
│   │   └── NodeActionButtons.tsx    # 节点操作按钮
│   └── OperationPanel/
│       ├── PanelDrawer.tsx          # 操作面板抽屉
│       ├── PanelTabs.tsx            # Tab 导航
│       ├── InfoTab.tsx              # 信息 Tab
│       ├── DecomposeTab.tsx         # 拆解 Tab
│       └── ConclusionTab.tsx        # 结论 Tab
├── hooks/
│   ├── useWorkspaceData.ts          # 工作区数据获取
│   ├── useNodeSelection.ts          # 节点选择逻辑
│   └── usePanelState.ts             # 面板状态管理
├── store/
│   └── workspaceStore.ts            # 工作区状态管理
└── types/
    └── workspace.ts                 # 类型定义
```

### 7.2 主要组件实现

#### WorkspaceLayout.tsx
```jsx
export function WorkspaceLayout() {
  return (
    <div className="h-screen flex flex-col">
      <TopBar />
      <div className="flex-1 flex relative">
        <VisualizationArea />
        <OperationPanel />
      </div>
    </div>
  )
}
```

## 8. 性能优化策略

### 8.1 渲染优化
- 使用 React.memo 避免不必要的重渲染
- 虚拟化大量节点的渲染
- 防抖处理高频交互事件

### 8.2 数据优化
- 懒加载节点详细信息
- 缓存常用数据，减少 API 调用
- 使用 SSE 替代轮询，减少网络开销

### 8.3 用户体验优化
- 骨架屏显示加载状态
- 乐观更新，先更新 UI 再同步后端
- 错误边界处理异常情况
