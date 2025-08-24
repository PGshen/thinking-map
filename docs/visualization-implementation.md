# 思维导图可视化实现方案

## 概述

本文档详细描述了思维导图可视化系统的技术实现方案，包括节点渲染、自动布局、动画系统、交互处理等核心功能。

## 架构设计

### 核心组件

1. **VisualizationArea** - 主可视化区域组件
2. **CustomNode** - 自定义节点组件
3. **useAutoLayout** - 自动布局Hook
4. **ReactFlow** - 底层图形渲染引擎

### 技术栈

- **React 18** - 前端框架
- **ReactFlow** - 图形可视化库
- **Zustand** - 状态管理
- **TypeScript** - 类型安全
- **Tailwind CSS** - 样式系统

## 详细实现

### 1. 可视化区域 (VisualizationArea)

#### 文件位置
`/src/features/workspace/components/visualization-area/visualization-area.tsx`

#### 核心功能

##### 1.1 状态管理
```typescript
// ReactFlow状态
const [nodes, setNodes, onNodesChange] = useNodesState([]);
const [edges, setEdges, onEdgesChange] = useEdgesState([]);

// 交互状态
const [isDragging, setIsDragging] = useState(false);
const [layoutType, setLayoutType] = useState<LayoutType>('local');
const [pendingLayout, setPendingLayout] = useState<{ nodeId?: string; type: LayoutType } | null>(null);
```

##### 1.2 数据同步机制
- **从Store到ReactFlow**: 监听workspace store变化，同步节点和边数据
- **从ReactFlow到Store**: 拖拽结束时同步位置变更
- **动画状态保持**: 在数据同步时保留节点的动画状态

```typescript
// 同步节点数据，保留动画状态
useEffect(() => {
  if (nodesData.length > 0) {
    setNodes((currentNodes) => {
      return nodesData.map((node) => {
        const currentNode = currentNodes.find(n => n.id === node.id);
        const currentAnimationState = currentNode?.data ? {
          isAnimating: currentNode.data.isAnimating,
          animationDuration: currentNode.data.animationDuration,
          animationEasing: currentNode.data.animationEasing
        } : {};
        
        return {
          ...node,
          data: {
            ...node.data,
            ...currentAnimationState,
            // 事件处理函数
          }
        };
      });
    });
  }
}, [nodesData, selectedNodeIDs]);
```

##### 1.3 自动布局触发机制
- **节点创建时**: 通过SSE事件触发局部布局
- **手动触发**: 用户点击全局布局按钮
- **延迟执行**: 使用100ms延迟避免频繁布局

```typescript
// 监听节点变化，触发自动布局
useEffect(() => {
  if (nodesData.length > 0 && pendingLayout && !isDragging && !isLayouting) {
    const { nodeId, type } = pendingLayout;
    setPendingLayout(null);
    
    triggerLayoutTimeoutRef.current = setTimeout(async () => {
      const layoutResult = await applyAutoLayout(
        nodesData as Node<CustomNodeModel>[],
        edgesData,
        type,
        nodeId,
        nodeSizesRef.current
      );
      
      // 处理布局结果和动画
    }, 100);
  }
}, [nodesData, isDragging, isLayouting, pendingLayout]);
```

##### 1.4 动画系统

**两阶段动画实现**:
1. **第一阶段**: 设置目标位置并启用动画样式
2. **第二阶段**: 动画完成后移除动画样式并同步到store

```typescript
if (layoutResult.animationState) {
  // 第一步：添加动画样式并设置到目标位置
  setNodes((nds) => 
    nds.map((node) => {
      const layoutedNode = layoutResult.nodes.find(n => n.id === node.id);
      if (layoutedNode) {
        return { 
          ...node, 
          position: layoutedNode.position,
          data: {
            ...node.data,
            isAnimating: true,
            animationDuration: 500,
            animationEasing: 'cubic-bezier(0.4, 0, 0.2, 1)'
          }
        };
      }
      return node;
    })
  );
  
  // 第二步：动画完成后移除动画样式并同步到store
  setTimeout(() => {
    setNodes((nds) => 
      nds.map((node) => {
        const layoutedNode = layoutResult.nodes.find(n => n.id === node.id);
        if (layoutedNode) {
          actions.updateNode(node.id, { position: layoutedNode.position });
          actions.addChangedNodePosition(node.id); // 记录位置变更
          return { 
            ...node, 
            data: {
              ...node.data,
              isAnimating: false
            }
          };
        }
        return node;
      })
    );
    finishAnimation();
  }, 500);
}
```

##### 1.5 拖拽处理
- **拖拽开始**: 禁用节点动画，避免移动延迟
- **拖拽结束**: 同步位置到store，记录位置变更

```typescript
const onNodeDragStart = useCallback((event: any, draggedNode: any) => {
  setIsDragging(true);
  // 拖拽时禁用节点动画
  setNodes((nds) => 
    nds.map((node) => {
      if (node.id === draggedNode.id) {
        return {
          ...node,
          data: { ...node.data, isAnimating: false }
        };
      }
      return node;
    })
  );
}, [setNodes]);

const onNodeDragStop = useCallback((event: any, draggedNode: any) => {
  setIsDragging(false);
  actions.updateNode(draggedNode.id, { position: draggedNode.position });
  actions.addChangedNodePosition(draggedNode.id);
}, [actions]);
```

##### 1.6 SSE事件处理
- **节点创建事件**: 创建新节点并触发布局
- **节点更新事件**: 支持replace和append两种更新模式
- **连接状态管理**: 显示SSE连接状态

### 2. 自定义节点 (CustomNode)

#### 文件位置
`/src/features/workspace/components/custome-node/custom-node.tsx`

#### 核心功能

##### 2.1 节点尺寸测量
```typescript
// 测量节点实际尺寸并通知布局系统
useEffect(() => {
  if (nodeRef.current && data.onSizeChange) {
    const { clientWidth, clientHeight } = nodeRef.current;
    data.onSizeChange(data.id, { width: clientWidth, height: clientHeight });
  }
}, [data.question, data.target, data.conclusion, data.isEditing]);
```

##### 2.2 动画样式应用
```typescript
// 动画样式
const animationStyle = data.isAnimating ? {
  transition: `all ${data.animationDuration || 500}ms ${data.animationEasing || 'cubic-bezier(0.4, 0, 0.2, 1)'}`
} : {};

<div
  ref={nodeRef}
  className={`...节点样式... ${data.isAnimating ? '' : 'transition-all duration-200'}`}
  style={animationStyle}
>
```

##### 2.3 编辑模式
- **内联编辑**: 双击节点进入编辑模式
- **表单验证**: 保存时验证必填字段
- **状态管理**: 区分新增和编辑状态

##### 2.4 节点类型
- **problem**: 问题节点
- **information**: 信息节点
- **analysis**: 分析节点
- **generation**: 生成节点
- **evaluation**: 评估节点

### 3. 自动布局系统 (useAutoLayout)

#### 文件位置
`/src/features/workspace/hooks/use-auto-layout.ts`

#### 核心功能

##### 3.1 布局类型
- **global**: 全局布局，重新排列所有节点
- **local**: 局部布局，只调整新增节点及其兄弟节点

##### 3.2 动画状态检测
```typescript
// 检查是否有位置变化
const fromPositions = new Map<string, { x: number; y: number }>();
const toPositions = new Map<string, { x: number; y: number }>();
let hasPositionChanges = false;

layoutResult.nodes.forEach((newNode) => {
  const oldNode = nodes.find(n => n.id === newNode.id);
  if (oldNode) {
    const positionChanged = 
      Math.abs(newNode.position.x - oldNode.position.x) > 1 ||
      Math.abs(newNode.position.y - oldNode.position.y) > 1;
    
    if (positionChanged) {
      fromPositions.set(newNode.id, oldNode.position);
      toPositions.set(newNode.id, newNode.position);
      hasPositionChanges = true;
    }
  }
});
```

##### 3.3 布局配置
```typescript
interface LayoutConfig {
  direction: 'TB' | 'LR' | 'BT' | 'RL';
  nodeSpacing: number;
  rankSpacing: number;
  align: 'UL' | 'UR' | 'DL' | 'DR';
}

const DEFAULT_LAYOUT_CONFIG: LayoutConfig = {
  direction: 'TB',
  nodeSpacing: 100,
  rankSpacing: 150,
  align: 'UL'
};
```

## CSS动画系统

### 动画类定义
```css
/* 只在animating状态下启用节点动画 */
.react-flow__node.animating {
  transition: var(--rf-node-transition);
}

/* 只在animating状态下启用边动画 */
.react-flow__edge.animating path {
  transition: var(--rf-edge-transition);
}
```

### 动态类名应用
```typescript
// 节点动态类名
nodes={nodes.map(node => ({
  ...node,
  className: node.data.isAnimating ? 'animating' : ''
}))}

// 边动态类名
edges={edges.map(edge => ({
  ...edge,
  className: isDragging ? '' : 'animating'
}))}
```

## 性能优化

### 1. 事件处理优化
- 使用`useCallback`缓存事件处理函数
- 使用`useRef`存储稳定的处理函数引用
- 避免在渲染过程中创建新的函数实例

### 2. 布局计算优化
- 使用节点实际DOM尺寸进行精确布局
- 缓存节点尺寸信息，避免重复测量
- 延迟布局触发，避免频繁计算

### 3. 动画性能优化
- 使用CSS transition而非JavaScript动画
- 只在必要时启用动画样式
- 拖拽时禁用动画，提升响应性

### 4. 状态同步优化
- 分离ReactFlow状态和业务状态
- 使用批量更新减少重渲染
- 保留动画状态避免闪现

## 交互体验

### 1. 拖拽体验
- **即时响应**: 拖拽时禁用动画，节点立即跟随鼠标
- **位置同步**: 拖拽结束后同步位置到store和后端
- **视觉反馈**: 拖拽过程中提供清晰的视觉反馈

### 2. 布局动画
- **平滑过渡**: 自动布局时节点平滑移动到目标位置
- **同步动画**: 节点和边的动画保持同步
- **性能优化**: 使用硬件加速的CSS动画

### 3. 编辑体验
- **内联编辑**: 双击节点直接进入编辑模式
- **实时预览**: 编辑过程中实时显示内容变化
- **智能保存**: 区分新增和编辑操作，智能处理保存逻辑

## 数据流

### 1. 节点创建流程
```
SSE事件 → handleNodeCreated → 创建节点 → 触发布局 → 动画执行 → 同步到store
```

### 2. 节点更新流程
```
SSE事件 → handleNodeUpdated → 更新节点数据 → 重新渲染
```

### 3. 拖拽流程
```
拖拽开始 → 禁用动画 → 实时更新位置 → 拖拽结束 → 同步到store → 记录变更
```

### 4. 自动布局流程
```
触发条件 → 计算布局 → 检测变化 → 执行动画 → 同步状态 → 记录变更
```

## 错误处理

### 1. 布局计算错误
- 捕获布局计算异常
- 回退到原始位置
- 记录错误日志

### 2. 动画中断处理
- 支持动画取消机制
- 清理定时器和状态
- 避免内存泄漏

### 3. SSE连接错误
- 自动重连机制
- 错误状态显示
- 降级处理方案

## 扩展性设计

### 1. 布局算法扩展
- 插件化布局算法
- 自定义布局配置
- 动态布局切换

### 2. 节点类型扩展
- 可配置节点类型
- 自定义节点渲染
- 插件化节点功能

### 3. 动画系统扩展
- 自定义动画配置
- 多种动画效果
- 动画性能监控

## 总结

本可视化实现方案通过ReactFlow提供强大的图形渲染能力，结合自定义的节点组件、自动布局系统和动画机制，实现了高性能、高交互性的思维导图可视化体验。系统设计注重性能优化、用户体验和扩展性，为复杂的思维导图应用提供了坚实的技术基础。

### 关键特性
- **高性能渲染**: 基于ReactFlow的硬件加速渲染
- **智能布局**: 自动布局算法，支持全局和局部布局
- **流畅动画**: CSS动画系统，支持拖拽时禁用动画
- **实时同步**: SSE实时数据同步，支持多用户协作
- **良好体验**: 内联编辑、拖拽交互、视觉反馈
- **可扩展性**: 模块化设计，支持功能扩展