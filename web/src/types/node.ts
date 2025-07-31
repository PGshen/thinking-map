/*
 * @Date: 2025-07-07 22:05:27
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-07-08 00:39:50
 * @FilePath: /thinking-map/web/src/types/node.ts
 */
// 与后端 dto/node.go 对齐的节点类型定义
import { MessageResponse } from './message';
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
  parentID?: string;
  nodeType: string;
  question: string;
  target: string;
  status: 'initial' | 'pending' | 'running' | 'completed' | 'error';
  context?: DependentContext;
  decomposition?: Decomposition;
  conclusion?: Conclusion;
  metadata?: any;
  selected?: boolean;
  isEditing?: boolean;
  // 交互事件（由外部注入，非持久数据）
  onEdit?: (mapID: string|null, id: string, data: Partial<CustomNodeModel>) => Promise<NodeResponse>;
  onDelete?: (mapID: string|null, id: string) => void;
  onAddChild?: (id: string) => void;
  onSelect?: (id: string) => void;
  onDoubleClick?: (id: string) => void;
  onContextMenu?: (id: string, e: React.MouseEvent) => void;
  onUpdateID?: (mapID: string|null, oldID: string, newID: string) => void;
}

export interface DependentContext {
  ancestor: NodeContextItem[];
  prevSibling: NodeContextItem[];
  children: NodeContextItem[];
}

// 节点上下文项类型
export interface NodeContextItem {
  question: string;
  target: string;
  conclusion?: string;
  abstract?: string;
  status: string;
}

export interface Decomposition {
  isDecompose: boolean;
  lastMessageID: string;
  messages: MessageResponse[];
}

export interface Conclusion {
  lastMessageID: string;
  content: string;
  messages: MessageResponse[];
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
  decomposition: Decomposition;
  conclusion: Conclusion;
  status: string;
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
export type ResetNodeContextResponse = ApiResponse<NodeResponse>;