# ThinkingMap 前端：API 与实时同步

## 1. API 请求层设计
- 统一 API 封装：所有 RESTful 请求通过统一的 api 层发起，便于维护、拦截、Mock 和测试。
- 请求工具：推荐使用 fetch 封装或 axios，支持拦截器、超时、取消、全局错误处理。
- 接口分模块管理：如 mapApi、nodeApi、thinkingApi、userApi 等，按后端接口聚合。
- 类型安全：所有请求/响应参数定义 TypeScript 类型，结合后端接口文档自动生成或手动维护。
- 通用响应处理：统一处理后端 code/message/data 格式，自动抛出/捕获业务错误。
- 鉴权与Token管理：自动附加 JWT Token，失效时自动跳转登录或刷新。
- 全局 loading/error 状态：结合 Zustand 或 React Context 管理全局请求状态。

## 2. SSE 封装设计
- SSE 连接管理：封装为自定义 hook（如 useSSE），支持自动重连、断线检测、连接状态指示。
- 事件分发机制：收到事件后按 type 分发到对应 handler，或推送到 Zustand store。
- 事件类型与数据结构：严格对齐后端事件格式，定义 TypeScript 类型。
- 多路复用与订阅：支持按 mapId、userId 等参数建立/切换连接，或多页面复用同一连接。
- 性能与健壮性：心跳包/定时重连，事件去重与顺序保证，错误事件友好提示与自动恢复。

## 3. 类型定义与接口映射
- 所有 API 请求/响应、SSE 事件数据均定义 TS 类型（如 Node, Map, SSEEvent, NodeCreatedEvent 等）。
- 类型与后端接口/事件格式保持同步，推荐自动生成（如 openapi-typescript）或手动维护。

## 4. 与 Zustand 的集成
- API 请求/事件结果通过 action 分发到 Zustand store，驱动 UI 实时更新。
- SSE 事件直接写入/更新全局状态（如节点树、消息列表、连接状态等）。
- 支持局部/全局 loading、错误、连接状态等 UI 反馈。

## 5. 代码结构建议
- `/api`：RESTful API 封装、SSE hook
- `/types`：接口与事件类型定义
- `/store`：全局/模块状态管理
- `/hooks`：通用自定义 hook（如 useSSE、useApi、useRequest 等）

## 6. 健壮性与性能建议
- 支持请求重试、超时、取消
- SSE 支持断线重连、事件去重、顺序校验
- 全局错误捕获与用户友好提示
- 日志与调试信息可选输出

> API 与实时同步设计应兼顾类型安全、健壮性和高性能。 