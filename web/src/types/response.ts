// 与后端 dto/response.go 对齐的通用响应类型定义
export interface Response<T = any> {
  code: number;
  message: string;
  data?: T;
  timestamp: string;
  requestId: string;
}

export interface PaginationResponse<T = any> {
  total: number;
  page: number;
  pageSize: number;
  data: T;
} 