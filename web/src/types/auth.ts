/*
 * @Date: 2025-07-03 23:19:45
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-07-04 00:13:06
 * @FilePath: /thinking-map/web/src/types/auth.ts
 */
import type { ApiResponse } from './response';

// 注册参数类型
export interface RegisterParams {
  username: string;
  email: string;
  password: string;
  fullName: string;
}

// 注册/登录响应data
export interface AuthData {
  userId: string;
  username: string;
  email: string;
  fullName: string;
  accessToken: string;
  refreshToken: string;
  expiresIn: number;
}

// 注册响应类型
export type RegisterResponse = ApiResponse<AuthData>;

// 登录参数类型
export interface LoginParams {
  email: string;
  password: string;
}

// 登录响应类型
export type LoginResponse = ApiResponse<AuthData>;

// 刷新Token响应类型
export interface RefreshTokenData {
  accessToken: string;
  refreshToken: string;
  expiresIn: number;
}
export type RefreshTokenResponse = ApiResponse<RefreshTokenData>;

// 错误响应data
export interface AuthErrorData {
  field: string;
  error: string;
} 