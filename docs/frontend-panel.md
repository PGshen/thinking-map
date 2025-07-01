# ThinkingMap 前端：操作面板系统

## 1. 总体布局与结构
- 右侧抽屉式/侧栏面板，默认占界面右侧 35% 宽度，可拖拽调节（最小 25%，最大 50%），支持折叠/展开。
- Tab 垂直排列于面板左侧，内容区右侧。
- 面板内容区根据当前激活 Tab 展示不同内容。
- 顶部区域可显示节点类型、状态、返回/关闭按钮。

## 2. Tab 设计与内容
### 2.1 信息 Tab（节点信息）
- 数据来源：`node_details`（detail_type: info）、`thinking_nodes` 基础字段
- 展示内容：当前问题、目标描述、上下文背景、结论内容、依赖检查、执行控制、操作按钮

### 2.2 拆解 Tab（问题拆解）
- 数据来源：`node_details`（detail_type: decompose）、`messages`（type: 拆解相关）
- 显示条件：后端判断当前节点需要拆解时显示
- 展示内容：聊天式对话区、拆解流程进度、子问题建议列表、交互

### 2.3 结论 Tab（结论生成）
- 数据来源：`node_details`（detail_type: conclusion）、`messages`（type: 结论相关）
- 显示条件：节点无需拆解或拆解完成后
- 展示内容：聊天式对话区、信息汇总区、结论内容编辑/确认区、交互

## 3. 交互与状态管理
- Tab切换时自动保存未提交内容或弹出确认
- 面板与节点选中状态联动，切换节点时自动刷新内容
- SSE事件驱动面板内容实时更新（如节点状态、对话消息、结论变更等）
- 支持快捷键（如Ctrl+S保存、Esc关闭面板）

## 4. 样式与体验建议
- 使用shadcn/ui的Tabs、Drawer、Button、Input、Textarea等组件，结合Tailwind CSS实现响应式和美观布局
- Tab标题高亮当前激活项，禁用不可用Tab
- 聊天区消息气泡区分AI/用户/系统，支持滚动、加载动画
- 依赖状态、执行按钮等有明显视觉反馈
- 面板内容区支持自适应高度、溢出滚动

## 5. 代码结构建议
- `PanelDrawer.tsx`：面板主组件，控制开关、宽度、折叠
- `PanelTabs.tsx`：Tab导航与切换
- `PanelInfoTab.tsx`：信息Tab内容与表单
- `PanelDecomposeTab.tsx`：拆解Tab聊天与流程
- `PanelConclusionTab.tsx`：结论Tab聊天与确认
- `PanelDependencyList.tsx`：依赖状态展示
- `PanelChatMessage.tsx`：聊天消息气泡
- `panelStore.ts`：面板相关状态管理

## 6. 示例布局（伪代码）
```jsx
<PanelDrawer>
  <PanelTabs>
    <PanelInfoTab />
    <PanelDecomposeTab />
    <PanelConclusionTab />
  </PanelTabs>
</PanelDrawer>
```

> 操作面板系统应支持内容懒加载、缓存、关键操作二次确认、消息历史回溯等高级功能。 