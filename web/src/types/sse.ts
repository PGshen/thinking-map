// 与后端 dto/sse.go 对齐的 SSE 事件类型定义
export interface NodeCreatedEvent {
  nodeId: string;
  parentId: string;
  nodeType: string;
  question: string;
  target: string;
  position: any; // TODO: 可根据 model.Position 进一步细化
  timestamp: string;
}

export interface NodeUpdatedEvent {
  nodeId: string;
  updates: Record<string, any>;
  timestamp: string;
}

export interface ThinkingProgressEvent {
  nodeId: string;
  stage: string;
  progress: number;
  message: string;
  timestamp: string;
}

export interface ErrorEvent {
  nodeId: string;
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
  eventId: string;
  eventType: string;
  sentAt: string;
  message: string;
}

export type SSEEvent = NodeCreatedEvent | NodeUpdatedEvent | ThinkingProgressEvent | ErrorEvent; 