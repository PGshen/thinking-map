import type { ApiResponse } from './response';

// 与后端 dto/node_detail.go 对齐的节点详情类型定义
export interface NodeDetailResponse {
  id: string;
  nodeID: string;
  detailType: string;
  content: any; // TODO: 可根据 model.DetailContent 进一步细化
  status: number;
  metadata: Record<string, any>;
  createdAt: string;
  updatedAt: string;
}

export type NodeDetailListResponse = ApiResponse<{
  details: NodeDetailResponse[];
}>; 