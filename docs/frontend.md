# ThinkingMap 前端技术文档

## 1. 技术栈说明

- **TypeScript**：类型安全，提升开发效率和可维护性。
- **React 18**：主流前端框架，组件化开发。
- **Next.js 14**：基于 React 的应用框架，支持文件路由、SSR/SSG、API 路由等。
- **Vite**：开发环境和构建工具，Next.js 内部已集成高效打包。
- **shadcn/ui**：基于 Radix UI 的 React 组件库，结合 Tailwind CSS，快速实现现代 UI。
- **Tailwind CSS**：原子化 CSS，灵活高效的样式方案。
- **ReactFlow**：强大的可视化流程图/节点图组件库。
- **Zustand**：轻量、易用的全局状态管理库。

---

## 2. 目录结构设计（推荐方案，结合详细实现）

```
/src
  /app                  # Next.js 14+ app 路由目录
    layout.tsx          # 全局布局（Provider、主题等）
    page.tsx            # 首页
    /history
      page.tsx          # 历史记录页
    /map/[id]
      page.tsx          # 工作区页面
      layout.tsx        # 工作区专用布局（可选）
  /layouts              # 各类布局组件（SidebarLayout、WorkspaceLayout等）
  /components           # 通用UI组件（Sidebar、Header、Button、Modal等）
  /features             # 业务模块（map、node、panel、chat、home等）
    /map                # 可视化区相关组件与逻辑
    /panel              # 操作面板相关组件与逻辑
    /home               # 首页问题输入、历史等
    /chat               # 聊天与消息相关
    ...
  /store                # Zustand状态管理（按模块拆分）
    mapStore.ts
    panelStore.ts
    globalStore.ts
    ...
  /api                  # API请求与SSE封装
    base.ts             # 通用请求封装
    mapApi.ts
    nodeApi.ts
    sse.ts              # SSE hook/工具
    ...
  /hooks                # 通用自定义hook（如useSSE、useApi、useRequest等）
  /types                # TypeScript类型定义（接口、事件、状态等）
  /utils                # 工具函数
  /styles               # 全局与模块样式（Tailwind、全局CSS等）
  /assets               # 静态资源（图片、SVG、字体等）
```

### 目录结构说明
- **/app**：Next.js 路由与页面入口，按页面/路由分目录，支持多布局
- **/layouts**：抽离各类布局组件，便于页面复用和切换
- **/components**：通用UI组件库，跨业务模块复用
- **/features**：按业务领域分模块，聚合相关组件、逻辑、样式，便于团队协作
- **/store**：Zustand状态管理，按业务/全局拆分
- **/api**：RESTful API与SSE封装，统一请求、类型安全
- **/hooks**：通用自定义hook，提升复用性
- **/types**：所有接口、事件、状态等TS类型定义，便于类型安全和自动生成
- **/utils**：通用工具函数
- **/styles**：全局与模块样式，Tailwind配置等
- **/assets**：静态资源，图片、SVG、字体等

> 该结构支持大型团队协作、模块化开发、易于扩展和维护。

---

## 3. 页面布局与路由方案

### 3.1 路由结构
- `/` 首页（带侧边栏）
- `/history` 历史记录页（带侧边栏）
- `/map/[id]` 工作区（全屏模式，无侧边栏）

### 3.2 布局实现
- `SidebarLayout`：包裹首页、历史页，左侧为侧边栏，右侧为主内容区。
- `WorkspaceLayout`：专用于工作区页面，隐藏侧边栏，全屏展示可视化与操作面板。
- 通过 Next.js 的 layout 机制实现不同页面的布局切换。

### 3.3 交互细节
- 侧边栏菜单项点击切换页面。
- 进入 `/map/[id]` 时，侧边栏隐藏，主工作区全屏。
- 工作区内提供返回按钮，返回首页或历史页时恢复侧边栏。

---

## 4. 状态管理方案（Zustand）

### 4.1 状态划分
- **全局状态**：用户信息、当前问题、节点树、SSE 连接状态、全局 loading/error。
- **局部状态**：表单输入、面板 Tab、节点选中、对话消息等。

### 4.2 使用建议
- 每个业务模块可单独维护 store，避免全局状态臃肿。
- 支持中大型项目的状态拆分与组合。
- 结合 React Context 提供跨组件访问。

---

## 5. 主要功能模块设计

### 5.1 问题输入与初步分析

#### 5.1.1 首页整体布局
- **采用侧边栏主布局（SidebarLayout）**，左侧为全局菜单，右侧为首页主内容区。
- 首页主内容区分为：
  - 顶部欢迎/产品简介区域（可选）
  - **问题输入区**（核心）
  - 历史问题快捷入口/列表
  - 产品亮点/引导说明（可选）

#### 5.1.2 问题输入区结构
- **多行文本输入框**：用于输入待分析的问题，支持粘贴、换行，自动聚焦。
- **问题类型选择**：下拉菜单或单选按钮，类型如"研究型/创意型/分析型/规划型"，与后端`problem_type`字段对应。
- **字数提示/建议**：实时显示当前字数，推荐范围（如50-200字），超限时高亮提示。
- **历史问题下拉**：可快速选择最近提交过的问题，支持搜索。
- **提交按钮**：高亮显示，输入有效时可用。
- **辅助说明**：如"请尽量描述清楚你的问题和目标"提示。

#### 5.1.3 交互与初步分析流程
1. 用户输入问题、选择类型，点击"提交"或"开始分析"。
2. 前端校验输入（字数、必填、类型等），通过后请求后端分析接口（如`POST /api/thinking/analyze`）。
3. 后端返回问题要点、目标、约束等初步分析结果。
4. 前端弹出/展示"系统理解确认区"：
   - 显示系统提取的核心要点、目标、约束
   - 用户可编辑/修正系统理解
   - 确认无误后点击"进入工作区"
5. 跳转到工作区（/map/[id]），初始化可视化图。

#### 5.1.4 与后端数据结构的映射
- `thinking_maps`表：
  - `problem`：用户输入的问题文本
  - `problem_type`：类型选择
  - `key_points`、`constraints`：后端分析提取的要点、约束
  - `target`：目标描述
- 初步分析结果通过API返回，前端展示并允许用户修正后再提交。

#### 5.1.5 样式与体验建议
- 问题输入区居中大卡片布局，突出主输入区域
- 输入框、下拉、按钮等使用shadcn/ui组件，Tailwind实现间距、圆角、阴影
- 字数提示、类型选择、历史下拉等辅助信息紧凑排列，便于快速操作
- 系统理解确认区采用弹窗或卡片，内容分区清晰，编辑区与只读区区分明显
- 错误提示、校验反馈即时、友好
- 支持键盘操作（回车提交、Tab切换等）
#### 5.1.6 推荐代码结构
- `HomePage.tsx`：首页主页面
- `QuestionInputCard.tsx`：问题输入卡片组件
- `ProblemTypeSelect.tsx`：类型选择组件
- `HistoryDropdown.tsx`：历史问题下拉组件
- `AnalyzeConfirmDialog.tsx`：系统理解确认弹窗/卡片
- `homeStore.ts`：首页相关状态管理

#### 5.1.7 示例布局（伪代码）
```jsx
<SidebarLayout>
  <div className="home-main">
    <QuestionInputCard>
      <Textarea ... />
      <ProblemTypeSelect ... />
      <HistoryDropdown ... />
      <CharCountHint ... />
      <Button>开始分析</Button>
    </QuestionInputCard>
    <AnalyzeConfirmDialog ... />
    <HistoryList ... />
  </div>
</SidebarLayout>
```

#### 5.1.8 其他建议
- 支持粘贴富文本自动转纯文本
- 历史问题可一键复用/编辑后再提交
- 首页可扩展为产品入口、引导页、FAQ等

### 5.2 可视化工作区（ReactFlow）

#### 5.2.1 结构分层
- **WorkspacePage（页面级）**：负责整体布局、数据加载、权限校验、全局事件处理。
- **MapCanvas（画布层）**：承载 ReactFlow 实例，负责节点、边的渲染与交互。
- **Node/Edge 自定义组件**：根据节点类型/状态自定义外观与交互。
- **Toolbar/Topbar**：画布顶部操作栏（如返回、缩放、布局、导出等）。
- **ContextMenu/快捷操作**：节点/画布右键菜单。

#### 5.2.2 核心组件
- `<ReactFlowProvider>`：包裹画布，提供上下文。
- `<ReactFlow>`：主画布组件，配置节点、边、交互事件。
- `CustomNode`：自定义节点（含类型图标、状态标识、内容省略、动画等）。
- `CustomEdge`：自定义边（支持平滑曲线、动画、状态高亮）。
- `NodeActionButtons`：节点单击后弹出的操作按钮组（编辑、删除、加子节点等）。
- `MiniMap`、`Controls`、`Background`：辅助组件，提升可用性。

#### 5.2.3 数据流与状态管理
- 节点树、边数据存储于 Zustand store，支持高频实时更新。
- 画布交互（如拖拽、节点编辑）通过 action 分发到 store，并同步后端。
- SSE 事件驱动节点/边的增删改，自动刷新画布。
- 节点选中、面板联动状态存于局部 store。

#### 5.2.4 关键交互实现
- **节点拖拽/缩放/平移**：ReactFlow 内置，支持自定义约束与动画。
- **节点单击**：弹出操作按钮组，定位于节点上方，自动避让边界。
- **节点双击**：高亮节点并展开右侧操作面板。
- **节点右键**：弹出快捷菜单（如复制、粘贴、删除、导出等）。
- **画布空白处点击**：取消节点选中，隐藏操作按钮组。
- **节点连接**：支持拖拽创建父子关系，自动校验合法性。

#### 5.2.5 动画与性能优化
- 节点/边变更采用平滑动画（如新增、删除、状态切换）。
- 大量节点时采用虚拟化/分层渲染，提升性能。
- 事件节流/防抖，避免高频更新导致卡顿。
- 只在必要时重渲染节点（如状态/内容变化）。

#### 5.2.6 与操作面板的联动
- 节点选中/双击时，右侧操作面板自动切换到对应节点详情。
- 面板内编辑节点信息，保存后实时同步到画布。
- 节点执行/拆解/结论生成等操作，触发节点状态和内容的实时更新。

#### 5.2.7 与后端的同步
- 画布初始化时拉取节点树和边数据。
- 所有节点/边的增删改操作均通过 API 同步后端。
- SSE 实时接收后端推送的节点/边变更事件，自动刷新画布。
- 支持断线重连和数据一致性校验。

#### 5.2.8 其他建议
- 支持导出图片/PDF、复制节点结构等高级功能。
- 画布自适应窗口大小，支持响应式和全屏。
- 关键交互（如删除节点）需二次确认，防止误操作。

#### 5.2.9 自定义节点实现细节

##### 1. 节点数据结构与后端映射
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

##### 2. 节点整体布局设计
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

##### 3. 节点状态与样式
- **待执行**：灰色边框，静态图标
- **执行中**：蓝色边框，loading动画
- **已完成**：绿色边框，完成图标
- **错误**：红色边框，警告/错误图标
- 状态切换带平滑过渡动画

##### 4. 交互与事件
- 单击节点：高亮并弹出操作按钮组
- 双击节点：展开右侧操作面板，节点高亮
- 右键节点：弹出快捷菜单
- 拖拽节点：更新位置并同步后端
- 拖拽连接：创建父子关系，自动校验
- 悬浮时显示完整内容tooltip

##### 5. 响应式与适配
- 节点宽度自适应内容，最小/最大宽度限制
- 支持多行内容自动截断与省略
- 适配不同缩放级别，保证主要信息可读

##### 6. 动画与性能
- 节点/边新增、删除、状态切换均有平滑动画
- 选中/高亮节点有明显视觉反馈
- 大量节点时仅渲染视窗内节点，提升性能

##### 7. 代码结构建议
- `CustomNode.tsx`：自定义节点主组件
- `NodeStatusIcon.tsx`：状态图标与动画
- `NodeActionButtons.tsx`：操作按钮组
- `NodeTooltip.tsx`：内容tooltip
- `nodeTypes.ts`：节点类型与样式配置

##### 8. 示例节点布局（伪代码）
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

##### 9. 样式建议
- 使用 Tailwind CSS utility class 实现配色、圆角、阴影、动画等
- 关键内容（问题、目标、结论）采用不同字号/颜色区分
- 状态色彩与后端/产品文档保持一致

### 5.3 操作面板系统

#### 5.3.1 总体布局与结构
- **右侧抽屉式/侧栏面板**，默认占界面右侧 35% 宽度，可拖拽调节（最小 25%，最大 50%），支持折叠/展开。
- **Tab 垂直排列**于面板左侧，内容区右侧。
- **面板内容区**根据当前激活 Tab 展示不同内容。
- **顶部区域**可显示节点类型、状态、返回/关闭按钮。

#### 5.3.2 Tab 设计与内容

##### 1. 信息 Tab（节点信息）
- **数据来源**：`node_details`（detail_type: info）、`thinking_nodes` 基础字段
- **展示内容**：
  - 当前问题（多行文本，支持编辑）
  - 目标描述（多行文本，支持编辑）
  - 上下文背景（富文本，支持格式化，内容结构见后端 context 字段）
  - 结论内容（仅在已完成状态显示，只读）
- **依赖检查**：
  - 展示所有父节点/依赖节点的状态（已完成/未完成）
  - 未满足依赖时禁用"开始执行"按钮，并高亮提示
- **执行控制**：
  - 满足依赖时显示"开始执行"按钮
  - 点击后请求后端判断是否需要拆解，自动跳转到拆解Tab或结论Tab
- **操作按钮**：
  - 保存修改、重置、删除节点（需二次确认）

##### 2. 拆解 Tab（问题拆解）
- **数据来源**：`node_details`（detail_type: decompose）、`messages`（type: 拆解相关）
- **显示条件**：后端判断当前节点需要拆解时显示
- **展示内容**：
  - 聊天式对话区，消息类型包括：
    - 系统消息（拆解进度、节点创建通知等）
    - AI消息（拆解分析、子问题建议等）
    - 用户消息（引导调整、确认修改等）
  - 拆解流程进度（如RAG查询、AI分析、节点创建等步骤）
  - 子问题建议列表，用户可调整/确认
- **交互**：
  - Tab激活后自动开始拆解流程
  - 用户可通过对话输入调整拆解建议
  - 确认后创建子节点，实时同步可视化区
  - 聊天消息与节点变更通过SSE实时同步

##### 3. 结论 Tab（结论生成）
- **数据来源**：`node_details`（detail_type: conclusion）、`messages`（type: 结论相关）
- **显示条件**：节点无需拆解或拆解完成后
- **展示内容**：
  - 聊天式对话区，消息类型包括：
    - AI结论生成、用户追问/补充、系统提示等
  - 信息汇总区：当前节点所有相关信息（问题、目标、上下文、子节点结论等）
  - 结论内容编辑/确认区（最终结论确认后只读）
- **交互**：
  - 用户可提问、要求AI调整或补充结论
  - 结论确认后节点状态变为已完成，内容同步到可视化区

#### 5.3.3 交互与状态管理
- Tab切换时自动保存未提交内容或弹出确认
- 面板与节点选中状态联动，切换节点时自动刷新内容
- SSE事件驱动面板内容实时更新（如节点状态、对话消息、结论变更等）
- 支持快捷键（如Ctrl+S保存、Esc关闭面板）

#### 5.3.4 样式与体验建议
- 使用shadcn/ui的Tabs、Drawer、Button、Input、Textarea等组件，结合Tailwind CSS实现响应式和美观布局
- Tab标题高亮当前激活项，禁用不可用Tab
- 聊天区消息气泡区分AI/用户/系统，支持滚动、加载动画
- 依赖状态、执行按钮等有明显视觉反馈
- 面板内容区支持自适应高度、溢出滚动

#### 5.3.5 代码结构建议
- `PanelDrawer.tsx`：面板主组件，控制开关、宽度、折叠
- `PanelTabs.tsx`：Tab导航与切换
- `PanelInfoTab.tsx`：信息Tab内容与表单
- `PanelDecomposeTab.tsx`：拆解Tab聊天与流程
- `PanelConclusionTab.tsx`：结论Tab聊天与确认
- `PanelDependencyList.tsx`：依赖状态展示
- `PanelChatMessage.tsx`：聊天消息气泡
- `panelStore.ts`：面板相关状态管理

#### 5.3.6 示例布局（伪代码）
```jsx
<PanelDrawer>
  <PanelTabs>
    <PanelInfoTab />
    <PanelDecomposeTab />
    <PanelConclusionTab />
  </PanelTabs>
</PanelDrawer>
```

#### 5.3.7 与后端数据结构的映射
- `node_details`：按detail_type区分Tab内容（info/decompose/conclusion）
- `messages`：按message_type区分聊天消息（系统/AI/用户）
- 依赖、上下文、结论等字段结构详见后端文档

#### 5.3.8 其他建议
- 支持面板内容的懒加载和缓存，提升性能
- 关键操作（如删除、结论确认）需二次确认
- 聊天区支持消息历史回溯与搜索
- 面板可扩展更多Tab（如操作日志、AI分析等）

### 5.4 实时同步与后端交互
- SSE 连接管理（节点变更、思考进度等事件）。
- API 封装（节点 CRUD、思考流程、对话等）。
- 请求拦截、错误处理、全局 loading。
- SSE 事件分发与本地状态同步。

### 5.5 用户体验与辅助功能
- 响应式布局、面板/节点自适应。
- 加载、错误、状态反馈。
- 快捷键、右键菜单、拖拽连接。

---

## 6. 关键技术实现要点与建议

### 6.1 Next.js 页面与布局

#### 6.1.1 app 目录结构与页面划分
- 采用 Next.js 14+ 的 app 目录路由（推荐）
- 主要页面：
  - `/` 首页（带侧边栏）
  - `/history` 历史记录页（带侧边栏）
  - `/map/[id]` 工作区（全屏模式，无侧边栏）
- 目录示例：
```
/src/app
  layout.tsx           # 全局布局（如全局Provider、主题等）
  page.tsx             # 首页
  /history
    page.tsx           # 历史记录页
  /map/[id]
    page.tsx           # 工作区页面
    layout.tsx         # 工作区专用布局（可选）
```

#### 6.1.2 布局实现方式
- **全局 layout.tsx**：包裹全站，注入全局 Provider、主题、全局样式等
- **SidebarLayout**：首页、历史页专用布局，包含侧边栏和主内容区
- **WorkspaceLayout**：工作区专用布局，隐藏侧边栏，全屏展示
- 通过 Next.js layout 机制，按页面自动切换布局
- 可通过 context/provider 传递全局状态（如用户、主题、SSE连接等）

#### 6.1.3 侧边栏与全屏工作区的切换
- 侧边栏组件（Sidebar.tsx）在 SidebarLayout 中渲染，菜单项支持高亮与跳转
- 进入 `/map/[id]` 时，切换为 WorkspaceLayout，侧边栏不渲染，主内容区全屏
- 工作区内可有返回按钮，返回首页或历史页时恢复侧边栏

#### 6.1.4 页面跳转与参数传递
- 使用 Next.js 的 `useRouter`、`Link` 组件进行页面跳转
- `/map/[id]` 动态路由，id 通过 params 传递给工作区组件
- 页面间可通过 URL 参数、Zustand 全局状态或 context 传递必要信息

#### 6.1.5 响应式与无障碍建议
- 布局采用 Tailwind CSS 实现响应式，支持最小宽度 1200px
- 侧边栏可收缩/自适应，主内容区自适应宽度
- 所有交互元素（按钮、菜单、Tab等）支持键盘操作和aria标签

#### 6.1.6 代码结构与复用
- 布局组件（SidebarLayout、WorkspaceLayout）放在 `/layouts` 目录，便于复用
- 侧边栏、顶部栏、面板等通用组件放在 `/components`
- 页面内容与布局解耦，便于后续扩展更多页面或布局

#### 6.1.7 与状态管理的集成
- 全局 Provider（如 Zustand、ThemeProvider）在全局 layout.tsx 注入
- 页面/布局可直接访问全局状态（如用户、SSE连接、主题等）
- 页面跳转时可触发全局状态变更（如切换当前 mapId、重置面板等）

#### 6.1.8 示例伪代码
```tsx
// src/app/layout.tsx
<Providers>
  {children}
</Providers>

// src/layouts/SidebarLayout.tsx
<div className="flex">
  <Sidebar />
  <main>{children}</main>
</div>

// src/layouts/WorkspaceLayout.tsx
<div className="w-full h-full">
  <WorkspaceHeader />
  <main>{children}</main>
</div>

// src/app/map/[id]/page.tsx
<WorkspaceLayout>
  <MapWorkspace mapId={params.id} />
</WorkspaceLayout>
```

### 6.2 Zustand 状态管理

#### 6.2.1 状态划分与设计原则
- **全局状态**：用户信息、全局 loading/error、SSE 连接状态、当前 mapId、主题等
- **模块状态**：
  - 可视化区（节点树、选中节点、画布缩放/平移等）
  - 操作面板（当前Tab、表单内容、依赖状态、对话消息等）
  - 首页（问题输入、类型选择、历史列表等）
- **拆分 store**：每个业务模块单独维护 store，避免全局状态臃肿，便于维护和测试
- **类型安全**：所有状态、action、selector 定义 TypeScript 类型
- **只存 UI 相关状态**：后端数据通过 API/SSE 获取，store 只存 UI 需要的部分

#### 6.2.2 Store 结构与示例
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

#### 6.2.3 与 API/SSE 的集成
- API 请求/响应通过 action 更新 store 状态（如节点树、用户信息等）
- SSE 事件直接写入/更新相关 store（如节点变更、消息推送、连接状态等）
- 支持 selector 订阅局部状态，避免无关组件重渲染

#### 6.2.4 持久化与中间件
- 推荐使用 zustand/middleware 的 persist、immer、devtools 等
- 关键状态（如用户登录态、主题）可本地持久化，其他状态随页面刷新重置
- 开发环境启用 devtools，便于调试

#### 6.2.5 性能优化
- 拆分 store，按需订阅，减少全局重渲染
- 使用 selector 精细订阅，避免无关组件更新
- 大数据（如节点树）可用 shallow 比较或 memo 优化
- 只在必要时更新状态，避免高频写入

#### 6.2.6 类型定义与代码结构
- `/store` 目录下按模块拆分 store 文件
- `/types` 目录定义所有状态、action、数据结构类型
- store 内部 action、selector、初始状态全部类型化

#### 6.2.7 最佳实践
- 只在组件树顶层注入 Provider（如有 context）
- 组件内通过 hook 访问和操作 store
- 复杂业务逻辑封装为 action，组件只负责 UI
- 状态变更与后端同步解耦，便于测试和维护

#### 6.2.8 示例伪代码
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

### 6.3 ReactFlow 可视化
- 节点数据结构与后端同步。
- 节点状态（待执行/执行中/已完成/错误）样式。
- 节点内容省略、图标、动画。
- 节点事件与面板联动。

### 6.4 shadcn/ui + Tailwind
- 统一 UI 风格，快速开发表单、弹窗、按钮、Tab、侧边栏等。
- Tailwind utility class 实现响应式和自适应。

### 6.5 API 与 SSE 封装

#### 6.5.1 API 请求层设计
- **统一 API 封装**：所有 RESTful 请求通过统一的 api 层发起，便于维护、拦截、Mock 和测试。
- **请求工具**：推荐使用 fetch 封装或 axios，支持拦截器、超时、取消、全局错误处理。
- **接口分模块管理**：如 mapApi、nodeApi、thinkingApi、userApi 等，按后端接口聚合。
- **类型安全**：所有请求/响应参数定义 TypeScript 类型，结合后端接口文档自动生成或手动维护。
- **通用响应处理**：统一处理后端 code/message/data 格式，自动抛出/捕获业务错误。
- **鉴权与Token管理**：自动附加 JWT Token，失效时自动跳转登录或刷新。
- **全局 loading/error 状态**：结合 Zustand 或 React Context 管理全局请求状态。

##### 示例结构
```ts
// api/base.ts
export async function request<T>(url: string, options: RequestInit): Promise<T> { ... }

// api/mapApi.ts
export const createMap = (params: CreateMapParams) => request<Map>("/api/maps", { method: "POST", body: ... });
```

#### 6.5.2 SSE 封装设计
- **SSE 连接管理**：封装为自定义 hook（如 useSSE），支持自动重连、断线检测、连接状态指示。
- **事件分发机制**：收到事件后按 type 分发到对应 handler，或推送到 Zustand store。
- **事件类型与数据结构**：严格对齐后端事件格式（如 node_created、node_updated、thinking_progress、error 等），定义 TypeScript 类型。
- **多路复用与订阅**：支持按 mapId、userId 等参数建立/切换连接，或多页面复用同一连接。
- **性能与健壮性**：
  - 心跳包/定时重连，防止断线
  - 事件去重与顺序保证
  - 错误事件友好提示与自动恢复

##### 示例结构
```ts
// api/sse.ts
export function useSSE({ url, onEvent }: { url: string, onEvent: (event: SSEEvent) => void }) { ... }

// store/sseStore.ts
export const useSSEStore = create<...>(...)
```

#### 6.5.3 类型定义与接口映射
- 所有 API 请求/响应、SSE 事件数据均定义 TS 类型（如 Node, Map, SSEEvent, NodeCreatedEvent 等）
- 类型与后端接口/事件格式保持同步，推荐自动生成（如 openapi-typescript）或手动维护

#### 6.5.4 与 Zustand 的集成
- API 请求/事件结果通过 action 分发到 Zustand store，驱动 UI 实时更新
- SSE 事件直接写入/更新全局状态（如节点树、消息列表、连接状态等）
- 支持局部/全局 loading、错误、连接状态等 UI 反馈

#### 6.5.5 代码结构建议
- `/api`：RESTful API 封装、SSE hook
- `/types`：接口与事件类型定义
- `/store`：全局/模块状态管理
- `/hooks`：通用自定义 hook（如 useSSE、useApi、useRequest 等）

#### 6.5.6 健壮性与性能建议
- 支持请求重试、超时、取消
- SSE 支持断线重连、事件去重、顺序校验
- 全局错误捕获与用户友好提示
- 日志与调试信息可选输出

#### 6.5.7 示例伪代码
```ts
// useSSE.ts
export function useSSE({ url, onEvent }) {
  useEffect(() => {
    const es = new EventSource(url);
    es.onmessage = (e) => onEvent(JSON.parse(e.data));
    es.onerror = () => {/* 自动重连等 */};
    return () => es.close();
  }, [url]);
}

// api/base.ts
export async function request<T>(url: string, options: RequestInit): Promise<T> {
  // 统一鉴权、错误处理、类型校验
}
```

### 6.6 目录结构与代码组织
- 业务模块、通用组件、hooks、store、api、types 分层清晰。
- 便于多人协作和后期维护。
