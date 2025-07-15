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
  DeleteNodeResponse,
  ResetNodeContextResponse
} from '@/types/node';

// 获取思维导图的所有节点
export async function getMapNodes(mapId: string): Promise<NodeListResponse> {
  return get(API_ENDPOINTS.NODE.GET.replace(':mapId', mapId));
}

// 创建节点
export async function createNode(mapId: string, params: CreateNodeRequest): Promise<CreateNodeResponse> {
  return post(API_ENDPOINTS.NODE.CREATE.replace(':mapId', mapId), params);
}

// 更新节点
export async function updateNode(mapId: string, nodeId: string, params: UpdateNodeRequest): Promise<UpdateNodeResponse> {
  return put(API_ENDPOINTS.NODE.UPDATE.replace(':mapId', mapId).replace(':nodeId', nodeId), params);
}

// 更新节点上下文
export async function updateNodeContext(mapId: string, nodeId: string, params: UpdateNodeContextRequest): Promise<UpdateNodeContextResponse> {
  return put(API_ENDPOINTS.NODE.CONTEXT.replace(':mapId', mapId).replace(':nodeId', nodeId), params);
}

// 重置节点上下文
export async function resetNodeContext(mapId: string, nodeId: string): Promise<ResetNodeContextResponse> {
  return put(API_ENDPOINTS.NODE.CONTEXT_RESET.replace(':mapId', mapId).replace(':nodeId', nodeId));
}

// 删除节点
export async function deleteNode(mapId: string, nodeId: string): Promise<DeleteNodeResponse> {
  return del(API_ENDPOINTS.NODE.DELETE.replace(':mapId', mapId).replace(':nodeId', nodeId));
}
