/*
 * @Date: 2025-07-07 22:05:27
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-07-08 00:39:50
 * @FilePath: /thinking-map/web/src/types/node.ts
 */
// 与后端 dto/node.go 对齐的节点类型定义
import type { NodeDetailResponse } from './nodeDetail';
import type { ApiResponse } from './response';

export interface Position {
  x: number;
  y: number;
}

export interface DependencyInfo {
  name: string;
  status: string;
}

export interface DependencyResponse {
  dependencies: DependencyInfo[];
  dependentNodes: DependencyInfo[];
}

export interface NodeResponse {
  id: string;
  parentID: string;
  nodeType: string;
  question: string;
  target: string;
  status: number;
  position: Position;
}

export type NodeListResponse = ApiResponse<{
  nodes: NodeResponse[];
}>;

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
  dependencies?: DependencyInfo[];
  context?: any;
  metadata?: any;
  selected?: boolean;
  childCount?: number;
  // 交互事件（由外部注入，非持久数据）
  onEdit?: (id: string) => void;
  onDelete?: (id: string) => void;
  onAddChild?: (id: string) => void;
  onSelect?: (id: string) => void;
  onDoubleClick?: (id: string) => void;
  onContextMenu?: (id: string, e: React.MouseEvent) => void;
}