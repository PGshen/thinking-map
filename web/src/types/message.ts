import type { ApiResponse } from './response';

// 与后端 dto/message.go 对齐的消息类型定义
export type RoleType = 'system' | 'assistant' | 'user';

export type MessageType = 'text' | 'notice' | 'rag' | 'action' | 'thought' | 'plan';

// 通知信息
export interface Notice {
  type: string;
  name: string;
  content: string;
}

// 动作信息
export interface Action {
  name: string;
  url: string;
  method: string;
  param?: Record<string, any>;
}

// 消息内容 - 与后端 model.MessageContent 对齐
export interface MessageContent {
  text?: string;
  thought?: string;
  rag?: string[];
  notice?: Notice[];
  action?: Action[];
}

export interface MessageResponse {
  id: string;
  parentID?: string;
  conversationID?: string;
  messageType: MessageType;
  role: RoleType;
  content: MessageContent;
  metadata?: any;
  createdAt?: string;
  updatedAt?: string;
}

export type MessageListResponse = ApiResponse<{
  total: number;
  page: number;
  limit: number;
  items: MessageResponse[];
}>;
