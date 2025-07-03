import type { ApiResponse } from './response';

// 与后端 dto/message.go 对齐的消息类型定义
export type RoleType = 'system' | 'assistant' | 'user';

export interface MessageResponse {
  id: string;
  parentId: string;
  chatId: string;
  messageType: string;
  role: RoleType;
  content: any; // TODO: 可根据 model.MessageContent 进一步细化
  metadata: any;
  createdAt: string;
  updatedAt: string;
}

export type MessageListResponse = ApiResponse<{
  total: number;
  page: number;
  limit: number;
  items: MessageResponse[];
}>; 