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
  ResetNodeContextResponse,
  ExecutableNodeResponse
} from '@/types/node';
import { MessageListResponse, MessageResponse } from '@/types/message';
import { ApiResponse } from '@/types/response';

// 获取思维导图的所有节点
export async function getMapNodes(mapID: string): Promise<NodeListResponse> {
  return get(API_ENDPOINTS.NODE.GET.replace(':mapID', mapID));
}

// 创建节点
export async function createNode(mapID: string, params: CreateNodeRequest): Promise<CreateNodeResponse> {
  return post(API_ENDPOINTS.NODE.CREATE.replace(':mapID', mapID), params);
}

// 更新节点
export async function updateNode(mapID: string, nodeID: string, params: UpdateNodeRequest): Promise<UpdateNodeResponse> {
  return put(API_ENDPOINTS.NODE.UPDATE.replace(':mapID', mapID).replace(':nodeID', nodeID), params);
}

// 更新节点上下文
export async function updateNodeContext(mapID: string, nodeID: string, params: UpdateNodeContextRequest): Promise<UpdateNodeContextResponse> {
  return put(API_ENDPOINTS.NODE.CONTEXT.replace(':mapID', mapID).replace(':nodeID', nodeID), params);
}

// 重置节点上下文
export async function resetNodeContext(mapID: string, nodeID: string): Promise<ResetNodeContextResponse> {
  return put(API_ENDPOINTS.NODE.CONTEXT_RESET.replace(':mapID', mapID).replace(':nodeID', nodeID));
}

// 删除节点
export async function deleteNode(mapID: string, nodeID: string): Promise<DeleteNodeResponse> {
  return del(API_ENDPOINTS.NODE.DELETE.replace(':mapID', mapID).replace(':nodeID', nodeID));
}

// 可执行节点
export async function executableNodes(mapID: string, nodeID?: string): Promise<ExecutableNodeResponse> {
  const url = API_ENDPOINTS.NODE.EXECUTABLE_NODES.replace(':mapID', mapID);
  const params = nodeID ? { nodeID } : undefined;
  return get(url, { params });
}

// 查询消息列表
export async function getMessages(mapID: string, nodeID: string, conversationType: string): Promise<ApiResponse<MessageResponse[]>> {
  return get(API_ENDPOINTS.NODE.MESSAGES.replace(':mapID', mapID).replace(':nodeID', nodeID), {
    params: {
      conversationType,
    }
  });
}

// 拆解
export async function decomposition(nodeID: string, clarification: string, isDecomposed: boolean): Promise<ApiResponse<any>> {
  return post(API_ENDPOINTS.THINKING.DECOMPOSITION, {
    nodeID,
    clarification,
    isDecomposed,
  });
}

// 结论
export async function conclusion(nodeID: string, reference: string, instruction: string): Promise<ApiResponse<any>> {
  return post(API_ENDPOINTS.THINKING.CONCLUSION, {
    nodeID,
    reference,
    instruction,
  });
}
