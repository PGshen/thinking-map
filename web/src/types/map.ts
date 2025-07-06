import type { ApiResponse } from './response';

// 创建思维导图请求参数
export interface CreateMapRequest {
  problem: string; // 必填，问题描述，最大1000字符
  problemType?: string; // 可选，问题类型，最大50字符
  target?: string; // 可选，目标，最大1000字符
  keyPoints?: string[]; // 可选，关键点
  constraints?: string[]; // 可选，约束条件
}

// 创建思维导图响应数据
export interface MapDetail {
  id: string;
  rootNodeId: string;
  status: number;
  problem: string;
  problemType?: string;
  target?: string;
  keyPoints?: string[];
  constraints?: string[];
  conclusion?: string;
  metadata?: Record<string, any>;
  createdAt: string;
  updatedAt: string;
}

export type CreateMapResponse = ApiResponse<MapDetail>;

// 旧的 MapResponse 及 MapListResponse 可保留用于列表等场景
export interface MapResponse {
  id: string;
  title: string;
  description: string;
  rootQuestion: string;
  rootNodeId: string;
  status: number;
  metadata: any;
  nodeCount?: number;
  createdAt: string;
  updatedAt: string;
}

export type MapListResponse = ApiResponse<{
  total: number;
  page: number;
  limit: number;
  items: MapResponse[];
}>; 