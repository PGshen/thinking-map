// 与后端 dto/node.go 对齐的节点类型定义
import type { NodeDetailResponse } from './nodeDetail';

export interface Position {
  // 具体结构可根据 model.Position 细化
  [key: string]: any;
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
  mapId?: string;
  parentId: string;
  nodeType: string;
  question: string;
  target: string;
  context: string;
  status: number;
  position: Position;
  dependencies: any; // TODO: 可根据 model.Dependencies 进一步细化
  nodeDetails: NodeDetailResponse[];
  metadata: any;
  createdAt: string;
  updatedAt: string;
}

export interface NodeListResponse {
  nodes: NodeResponse[];
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