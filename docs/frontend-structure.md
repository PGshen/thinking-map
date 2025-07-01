# ThinkingMap 前端：目录结构与技术栈

## 1. 技术栈说明

- **TypeScript**：类型安全，提升开发效率和可维护性。
- **React 18**：主流前端框架，组件化开发。
- **Next.js 14**：基于 React 的应用框架，支持文件路由、SSR/SSG、API 路由等。
- **Vite**：开发环境和构建工具，Next.js 内部已集成高效打包。
- **shadcn/ui**：基于 Radix UI 的 React 组件库，结合 Tailwind CSS，快速实现现代 UI。
- **Tailwind CSS**：原子化 CSS，灵活高效的样式方案。
- **ReactFlow**：强大的可视化流程图/节点图组件库。
- **Zustand**：轻量、易用的全局状态管理库。

> 以上技术栈为前端开发的基础，建议团队成员熟悉相关文档与最佳实践。

---

## 2. 目录结构设计

推荐采用如下分层结构，便于大型团队协作、模块化开发、易于扩展和维护：

```text
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

> 目录结构应根据实际业务发展灵活调整，保持清晰分层和高内聚低耦合。 