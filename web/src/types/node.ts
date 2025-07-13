/*
 * @Date: 2025-07-07 22:05:27
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-07-08 00:39:50
 * @FilePath: /thinking-map/web/src/types/node.ts
 */
// 与后端 dto/node.go 对齐的节点类型定义
import type { ApiResponse } from './response';

export interface Position {
  x: number;
  y: number;
}

export interface Node {
  id: string;
  label: string;
  // 可根据实际需求扩展属性
}

export interface Edge {
  id: string;
  source: string;
  target: string;
  // 可根据实际需求扩展属性
}

export interface CustomNodeModel {
  id: string;
  parentId?: string;
  nodeType: string;
  question: string;
  target: string;
  conclusion?: string;
  status: 'pending' | 'running' | 'completed' | 'error';
  dependencies?: NodeContextItem[];
  context?: any;
  metadata?: any;
  selected?: boolean;
  childCount?: number;
  // 交互事件（由外部注入，非持久数据）
  onEdit?: (id: string, data: { question: string; target: string }) => void;
  onDelete?: (id: string) => void;
  onAddChild?: (id: string) => void;
  onSelect?: (id: string) => void;
  onDoubleClick?: (id: string) => void;
  onContextMenu?: (id: string, e: React.MouseEvent) => void;
}

// 节点上下文项类型
export interface NodeContextItem {
  question: string;
  target: string;
  conclusion: string;
  abstract: string;
}

export interface NodeResponse {
  id: string;
  mapID: string;
  parentID: string;
  nodeType: string;
  question: string;
  target: string;
  context: {
    ancestor: NodeContextItem[];
    prevSibling: NodeContextItem[];
    children: NodeContextItem[];
  };
  conclusion: string;
  status: number;
  position: Position;
  metadata: Record<string, any>;
  createdAt: string;
  updatedAt: string;
}

// 创建节点请求参数
export interface CreateNodeRequest {
  mapID: string;
  parentID: string;
  nodeType: string;
  question: string;
  target: string;
  position: Position;
}

// 更新节点请求参数
export interface UpdateNodeRequest {
  question?: string;
  target?: string;
  position?: Position;
}

// 更新节点上下文请求参数
export interface UpdateNodeContextRequest {
  context: {
    ancestor: NodeContextItem[];
    prevSibling: NodeContextItem[];
    children: NodeContextItem[];
  };
}

export type NodeListResponse = ApiResponse<{
  nodes: NodeResponse[];
}>;
export type CreateNodeResponse = ApiResponse<NodeResponse>;
export type UpdateNodeResponse = ApiResponse<NodeResponse>;
export type DeleteNodeResponse = ApiResponse<null>;
export type UpdateNodeContextResponse = ApiResponse<NodeResponse>;