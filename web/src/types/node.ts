// 与后端 dto/node.go 对齐的节点类型定义
import type { NodeDetailResponse } from './nodeDetail';
import type { ApiResponse } from './response';

export interface Position {
  x: number;
  y: number;
}

export interface DependencyInfo {
  nodeId: string;
  dependencyType: string;
  required: boolean;
}

export interface DependencyResponse {
  dependencies: DependencyInfo[];
  dependentNodes: DependencyInfo[];
}

export interface NodeResponse {
  id: string;
  parentId: string;
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