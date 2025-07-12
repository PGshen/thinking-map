import type { ApiResponse } from './response';

// 创建思维导图请求参数
export interface CreateMapRequest {
  title: string; // 必填，标题概述
  problem: string; // 必填，问题描述，最大1000字符
  problemType?: string; // 可选，问题类型，最大50字符
  target?: string; // 可选，目标，最大1000字符
  keyPoints?: string[]; // 可选，关键点
  constraints?: string[]; // 可选，约束条件
}

// 思维导图列表查询参数
export interface MapListQuery {
  page: number;
  limit: number;
  status?: number; // 0:全部, 1:进行中, 2:已完成
  problemType?: string;
  dateRange?: string;
  search?: string;
}

// 更新思维导图请求参数
export interface UpdateMapRequest {
  status?: number; // 1:进行中, 2:已完成
  title?: string;
  problem?: string;
  problemType?: string;
  target?: string;
  keyPoints?: string[];
  constraints?: string[];
  conclusion?: string;
}


// 创建思维导图响应数据
export interface Map {
  id: string;
  status: number;
  title: string;
  problem: string;
  problemType?: string;
  target?: string;
  keyPoints?: string[];
  constraints?: string[];
  conclusion?: string;
  metadata?: Record<string, any>;
  createdAt: string;
  updatedAt: string;
}

export interface MapList {
  total: number;
  page: number;
  limit: number;
  items: Map[];
}

// API响应类型
export type CreateMapResponse = ApiResponse<Map>;
export type MapListResponse = ApiResponse<MapList>;
export type MapDetailResponse = ApiResponse<Map>;
export type UpdateMapResponse = ApiResponse<Map>;