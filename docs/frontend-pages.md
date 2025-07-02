<!--
 * @Date: 2025-07-01 09:33:59
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-07-01 23:49:22
 * @FilePath: /thinking-map/docs/frontend-pages.md
-->
# ThinkingMap 前端：页面与路由

## 1. 路由结构
- `/` 首页（带侧边栏）
- `/history` 历史记录页（带侧边栏）
- `/map/[id]` 工作区（全屏模式，无侧边栏）

## 2. 布局实现
- `SidebarLayout`：包裹首页、历史页，左侧为侧边栏，右侧为主内容区。
- `WorkspaceLayout`：专用于工作区页面，隐藏侧边栏，全屏展示可视化与操作面板。
- 通过 Next.js 的 layout 机制实现不同页面的布局切换。

## 3. 交互细节
- 侧边栏菜单项点击切换页面。
- 进入 `/map/[id]` 时，侧边栏隐藏，主工作区全屏。
- 工作区内提供返回按钮，返回首页或历史页时恢复侧边栏。

## 4. 页面跳转与参数传递
- 使用 Next.js 的 `useRouter`、`Link` 组件进行页面跳转。
- `/map/[id]` 动态路由，id 通过 params 传递给工作区组件。
- 页面间可通过 URL 参数、Zustand 全局状态或 context 传递必要信息。

## 5. 响应式与无障碍建议
- 布局采用 Tailwind CSS 实现响应式，支持最小宽度 1200px。
- 侧边栏可收缩/自适应，主内容区自适应宽度。
- 所有交互元素（按钮、菜单、Tab等）支持键盘操作和aria标签。

## 6. 代码结构与复用
- 布局组件（SidebarLayout、WorkspaceLayout）放在 `/layouts` 目录，便于复用。
- 侧边栏、顶部栏、面板等通用组件放在 `/components`。
- 页面内容与布局解耦，便于后续扩展更多页面或布局。

## 7. 示例伪代码
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

> 页面与路由设计应兼顾易用性、可扩展性和团队协作需求。 