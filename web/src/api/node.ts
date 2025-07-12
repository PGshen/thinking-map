/*
 * @Date: 2025-07-08 00:39:50
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-07-08 00:39:50
 * @FilePath: /thinking-map/web/src/api/node.ts
 */
import { get, post, put, del } from './request';
import API_ENDPOINTS from './endpoints';
import type { 
  NodeListResponse, 
  NodeResponse, 
  CreateNodeRequest,
  UpdateNodeRequest,
  UpdateNodeContextRequest,
  CreateNodeResponse,
  UpdateNodeResponse,
  UpdateNodeContextResponse,
  DeleteNodeResponse
} from '@/types/node';
import type { ApiResponse } from '@/types/response';

// 获取思维导图的所有节点
export async function getMapNodes(mapId: string): Promise<NodeListResponse> {
  return get(`${API_ENDPOINTS.NODE.GET}/${mapId}/nodes`);
}

// 创建节点
export async function createNode(mapId: string, params: CreateNodeRequest): Promise<CreateNodeResponse> {
  return post(`${API_ENDPOINTS.NODE.CREATE}/${mapId}/nodes`, params);
}

// 更新节点
export async function updateNode(mapId: string, nodeId: string, params: UpdateNodeRequest): Promise<UpdateNodeResponse> {
  return put(`${API_ENDPOINTS.NODE.UPDATE}/${mapId}/nodes/${nodeId}`, params);
}

// 更新节点上下文
export async function updateNodeContext(mapId: string, nodeId: string, params: UpdateNodeContextRequest): Promise<UpdateNodeContextResponse> {
  return put(`${API_ENDPOINTS.NODE.CONTEXT}/${mapId}/nodes/${nodeId}/context`, params);
}

// 删除节点
export async function deleteNode(mapId: string, nodeId: string): Promise<DeleteNodeResponse> {
  return del(`${API_ENDPOINTS.NODE.DELETE}/${mapId}/nodes/${nodeId}`);
}
