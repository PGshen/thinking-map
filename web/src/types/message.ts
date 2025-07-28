import type { ApiResponse } from './response';

// 与后端 dto/message.go 对齐的消息类型定义
export type RoleType = 'system' | 'assistant' | 'user';

export interface MessageResponse {
  id: string;
  parentID: string;
  conversationID: string;
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

export interface ChatMsg {
  type: 'text' | 'tool' | 'action';
  textMsg?: ChatTextMessage;
  toolMsg?: ChatToolMessage;
  actionMsg?: ChatActionMessage;
};

export interface ChatTextMessage {
  id: string;
  role: RoleType;
  content: string;
}

export interface ChatToolMessage {
  id: string;
  role: RoleType;
  toolCall: ToolCall;
}

export interface ToolCall {
  id: string;
  type: string;
  function: {
    name: string;
    arguments: string;
  };
}

export interface ChatActionMessage {
  id: string;
  role: RoleType;
  actions: ChatAction[];
}

export interface ChatAction {
  name: string;
  url: string;
  arguments: string;
}
