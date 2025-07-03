// 与后端 dto/response.go 对齐的通用响应类型定义
export interface Response<T = any> {
  code: number;
  message: string;
  data?: T;
  timestamp: string;
  requestId: string;
}

// 通用API响应类型
export interface ApiResponse<T = any> {
  code: number;
  message: string;
  data: T;
  timestamp?: string;
  requestId?: string;
} 