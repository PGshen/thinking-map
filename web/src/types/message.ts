import type { ApiResponse } from './response';

// 与后端 dto/message.go 对齐的消息类型定义
export type RoleType = 'system' | 'assistant' | 'user';

export type MessageType = 'text' | 'notice' | 'rag' | 'action' | 'thought' | 'plan';

export type PlanStepStatus = 'pending' | 'running' | 'completed' | 'skipped' | 'failed';

export type NoticeType =  'error' | 'warning' | 'success' | 'info';

// 通知信息
export interface Notice {
  type: NoticeType;
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

export interface Plan {
  steps: PlanStep[];
}

export interface PlanStep {
  id: string;
  name: string;
  description: string;
  status: PlanStepStatus;
  assignedSpecialist?: string;
}

export interface RAGResult {
  title: string;
  url?: string;
  content?: string;
  score?: number;
  raw_content?: string;
  favicon?: string;
}

export interface RAGRecord {
  id: string;
  query: string;
  answer: string;
  sources: string;
  results: RAGResult[];
}

// 消息内容 - 与后端 model.MessageContent 对齐
export interface MessageContent {
  text?: string;
  thought?: string;
  rag?: RAGRecord;
  notice?: Notice;
  action?: Action[];
  plan?: Plan;
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
