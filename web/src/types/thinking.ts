// 与后端 dto/thinking.go 对齐的 AI 思考/推理相关类型定义
export interface ThinkingOptions {
  model: string; // 'gpt-4' | 'gpt-3.5-turbo'
  temperature: number;
}

export interface AnalyzeRequest {
  nodeID: string;
  context: string;
  options: ThinkingOptions;
}

export interface DecomposeRequest {
  nodeID: string;
  decomposeStrategy: string; // 'breadth_first' | 'depth_first'
  maxDepth: number;
}

export interface ConcludeRequest {
  nodeID: string;
  evidence: string[];
  reasoningType: string; // 'deductive' | 'inductive' | 'abductive'
}

export interface ChatRequest {
  nodeID: string;
  message: string;
  context: string; // 'decompose' | 'conclude'
}

export interface TaskResponse {
  taskID: string;
  nodeID: string;
  status: string;
  estimatedTime: number;
}

export interface ChatResponse {
  messageID: string;
  content: string;
  role: string;
  createdAt: string;
} 