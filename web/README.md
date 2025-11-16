# ThinkingMap 前端

本项目为 ThinkingMap 的前端实现，基于 Next.js 14、TypeScript、shadcn/ui、Radix UI、Tailwind CSS、ReactFlow、Zustand 等现代技术栈，支持可视化 AI 问题解决。

## 技术栈
- Next.js 14 (React 18)
- TypeScript
- Tailwind CSS
- shadcn/ui + Radix UI
- Zustand
- ReactFlow
- pnpm

## 目录结构

```text
/src
  /app                  # Next.js 14+ app 路由目录
  /layouts              # 各类布局组件
  /components           # 通用UI组件
  /features             # 业务模块（map、panel、home、chat等）
  /store                # Zustand状态管理
  /api                  # API请求与SSE封装
  /hooks                # 通用自定义hook
  /types                # TypeScript类型定义
  /utils                # 工具函数
  /styles               # 全局与模块样式
  /assets               # 静态资源
```

## 快速开始

1. 安装依赖
   ```sh
   pnpm install
   ```
2. 启动开发环境
   ```sh
   pnpm dev
   ```
3. 访问 http://localhost:3030

## 代码规范
- 使用 Prettier 统一代码风格
- 推荐使用 VSCode + Volar/ESLint 插件

## 相关文档
- [产品需求文档](../docs/prd.md)
- [前端技术文档](../docs/frontend.md)
- [目录结构与技术栈](../docs/frontend-structure.md)
