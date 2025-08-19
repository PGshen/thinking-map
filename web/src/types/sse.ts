import { Action } from './message';
import { Position } from './node';

// 与后端 dto/sse.go 对齐的 SSE 事件类型定义
export interface NodeCreatedEvent {
  nodeID: string;
  parentID: string;
  nodeType: string;
  question: string;
  target: string;
  position: Position;
  timestamp: string;
}

export interface NodeUpdatedEvent {
  nodeID: string;
  mode: string; // 更新模式：replace/append
  updates: Record<string, any>;
  timestamp: string;
}

export interface ThinkingProgressEvent {
  nodeID: string;
  stage: string;
  progress: number;
  message: string;
  timestamp: string;
}

export interface MessageActionEvent {
  nodeID: string;
  messageID: string;
  actions: Action[];
  timestamp: string;
}

export interface MessageTextEvent {
  nodeID: string;
  messageID: string;
  message: string;
  mode: 'append' | 'replace';
  timestamp: string;
}

export interface MessageThoughtEvent {
  nodeID: string;
  messageID: string;
  message: string;
  mode: 'append' | 'replace';
  timestamp: string;
}

export interface ErrorEvent {
  nodeID: string;
  errorCode: string;
  errorMessage: string;
  timestamp: string;
}

export interface TestEventRequest {
  eventType: 'node_created' | 'node_updated' | 'thinking_progress' | 'error' | 'custom';
  data: Record<string, any>;
  delay: number;
}

export interface TestEventResponse {
  eventID: string;
  eventType: string;
  sentAt: string;
  message: string;
}

export type SSEEvent = NodeCreatedEvent | NodeUpdatedEvent | ThinkingProgressEvent | ErrorEvent;