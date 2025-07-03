// 与后端 dto/map.go 对齐的思维导图类型定义
export interface MapResponse {
  id: string;
  rootNodeId?: string;
  status: number;
  problem: string;
  problemType: string;
  target: string;
  keyPoints: any; // TODO: 可根据 model.KeyPoints 进一步细化
  constraints: any; // TODO: 可根据 model.Constraints 进一步细化
  conclusion: string;
  metadata: any;
  createdAt: string; // ISO 时间字符串
  updatedAt: string; // ISO 时间字符串
}

export interface MapListResponse {
  total: number;
  page: number;
  limit: number;
  items: MapResponse[];
} 