import type { ApiResponse } from './response';

// 与后端 dto/map.go 对齐的思维导图类型定义
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