# ThinkingMap 前端：可视化与交互（ReactFlow）

## 1. 结构分层
- **WorkspacePage（页面级）**：负责整体布局、数据加载、权限校验、全局事件处理。
- **MapCanvas（画布层）**：承载 ReactFlow 实例，负责节点、边的渲染与交互。
- **Node/Edge 自定义组件**：根据节点类型/状态自定义外观与交互。
- **Toolbar/Topbar**：画布顶部操作栏（如返回、缩放、布局、导出等）。
- **ContextMenu/快捷操作**：节点/画布右键菜单。

## 2. 核心组件
- `<ReactFlowProvider>`：包裹画布，提供上下文。
- `<ReactFlow>`：主画布组件，配置节点、边、交互事件。
- `CustomNode`：自定义节点（含类型图标、状态标识、内容省略、动画等）。
- `CustomEdge`：自定义边（支持平滑曲线、动画、状态高亮）。
- `NodeActionButtons`：节点单击后弹出的操作按钮组（编辑、删除、加子节点等）。
- `MiniMap`、`Controls`、`Background`：辅助组件，提升可用性。

## 3. 数据流与状态管理
- 节点树、边数据存储于 Zustand store，支持高频实时更新。
- 画布交互（如拖拽、节点编辑）通过 action 分发到 store，并同步后端。
- SSE 事件驱动节点/边的增删改，自动刷新画布。
- 节点选中、面板联动状态存于局部 store。

## 4. 关键交互实现
- 节点拖拽/缩放/平移：ReactFlow 内置，支持自定义约束与动画。
- 节点单击：弹出操作按钮组，定位于节点上方，自动避让边界。
- 节点双击：高亮节点并展开右侧操作面板。
- 节点右键：弹出快捷菜单（如复制、粘贴、删除、导出等）。
- 画布空白处点击：取消节点选中，隐藏操作按钮组。
- 节点连接：支持拖拽创建父子关系，自动校验合法性。

## 5. 动画与性能优化
- 节点/边变更采用平滑动画（如新增、删除、状态切换）。
- 大量节点时采用虚拟化/分层渲染，提升性能。
- 事件节流/防抖，避免高频更新导致卡顿。
- 只在必要时重渲染节点（如状态/内容变化）。

## 6. 与操作面板的联动
- 节点选中/双击时，右侧操作面板自动切换到对应节点详情。
- 面板内编辑节点信息，保存后实时同步到画布。
- 节点执行/拆解/结论生成等操作，触发节点状态和内容的实时更新。

## 7. 与后端的同步
- 画布初始化时拉取节点树和边数据。
- 所有节点/边的增删改操作均通过 API 同步后端。
- SSE 实时接收后端推送的节点/边变更事件，自动刷新画布。
- 支持断线重连和数据一致性校验。

## 8. 自定义节点实现细节
### 8.1 节点数据结构与后端映射
- 节点数据来源于后端 `thinking_nodes`、`node_details` 表，字段包括：
  - `id`：节点唯一标识
  - `parent_id`：父节点ID
  - `node_type`：节点类型（root/analysis/conclusion/custom等）
  - `question`：节点问题描述
  - `target`：目标描述
  - `conclusion`：结论内容（已完成时显示）
  - `status`：节点状态（0待执行/1执行中/2已完成/3错误等）
  - `position`：节点在画布上的坐标
  - `dependencies`：依赖信息
  - `context`：上下文信息（如父/子节点摘要）
  - `metadata`：扩展字段
- 详细信息（如上下文、拆解、结论Tab内容）通过 `node_details` 关联获取。

### 8.2 节点整体布局设计
- **外层容器**：圆角矩形/卡片，阴影、边框颜色根据状态变化
- **顶部区域**：
  - 类型图标 + 类型文字（如"分析"、"结论"）
  - 状态标识（灰/蓝/绿/红，带动画/图标）
- **主体内容**：
  - 问题概述（最多40字，超出省略号）
  - 目标概述（最多30字，超出省略号）
  - 结论内容（仅已完成时显示，最多50字，省略号）
- **底部区域**：
  - 依赖状态（如有未完成依赖，显示提示/图标）
  - 子节点数量/分支提示（可选）
- **交互按钮**：
  - 悬浮/选中时显示操作按钮组（编辑、删除、加子节点等）

### 8.3 节点状态与样式
- 待执行：灰色边框，静态图标
- 执行中：蓝色边框，loading动画
- 已完成：绿色边框，完成图标
- 错误：红色边框，警告/错误图标
- 状态切换带平滑过渡动画

### 8.4 交互与事件
- 单击节点：高亮并弹出操作按钮组
- 双击节点：展开右侧操作面板，节点高亮
- 右键节点：弹出快捷菜单
- 拖拽节点：更新位置并同步后端
- 拖拽连接：创建父子关系，自动校验
- 悬浮时显示完整内容tooltip

### 8.5 响应式与适配
- 节点宽度自适应内容，最小/最大宽度限制
- 支持多行内容自动截断与省略
- 适配不同缩放级别，保证主要信息可读

### 8.6 动画与性能
- 节点/边新增、删除、状态切换均有平滑动画
- 选中/高亮节点有明显视觉反馈
- 大量节点时仅渲染视窗内节点，提升性能

### 8.7 代码结构建议
- `CustomNode.tsx`：自定义节点主组件
- `NodeStatusIcon.tsx`：状态图标与动画
- `NodeActionButtons.tsx`：操作按钮组
- `NodeTooltip.tsx`：内容tooltip
- `nodeTypes.ts`：节点类型与样式配置

### 8.8 示例节点布局（伪代码）
```jsx
<div className={`node-card status-${status}`}> 
  <div className="node-header">
    <TypeIcon type={node_type} />
    <span>{typeText}</span>
    <NodeStatusIcon status={status} />
  </div>
  <div className="node-content">
    <div className="node-question" title={question}>{truncate(question, 40)}</div>
    <div className="node-target" title={target}>{truncate(target, 30)}</div>
    {status === 'completed' && <div className="node-conclusion" title={conclusion}>{truncate(conclusion, 50)}</div>}
  </div>
  <div className="node-footer">
    {hasUnmetDependencies && <DependencyIcon />}
    <span>{childCount}分支</span>
  </div>
  {selected && <NodeActionButtons ... />}
</div>
```

### 8.9 样式建议
- 使用 Tailwind CSS utility class 实现配色、圆角、阴影、动画等
- 关键内容（问题、目标、结论）采用不同字号/颜色区分
- 状态色彩与后端/产品文档保持一致
> 可视化区的设计应兼顾性能、交互体验和可扩展性。 