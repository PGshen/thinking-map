/*
 * @Date: 2025-07-03 22:03:37
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-07-04 00:33:39
 * @FilePath: /thinking-map/web/src/lib/auth.ts
 */
// token工具方法

import axios from "axios";

const TOKEN_KEY = 'token';
const REFRESH_TOKEN_KEY = 'refreshToken';
const USER_KEY = 'user';
const apiBaseURL = process.env.NEXT_PUBLIC_API_BASE_URL || '/api';

export function setToken(token: string, refreshToken: string): void {
  if (typeof window !== 'undefined') {
    localStorage.setItem(TOKEN_KEY, token || '');
    localStorage.setItem(REFRESH_TOKEN_KEY, refreshToken || '');
  }
}

export function getToken(): string {
  if (typeof window !== 'undefined') {
    return localStorage.getItem(TOKEN_KEY) || '';
  }
  return '';
}

export function getRefreshToken(): string {
  if (typeof window !== 'undefined') {
    return localStorage.getItem(REFRESH_TOKEN_KEY) || '';
  }
  return '';
}

export function removeToken(): void {
  if (typeof window !== 'undefined') {
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(REFRESH_TOKEN_KEY);
    localStorage.removeItem(USER_KEY);
  }
}

// 用户信息存储和获取
export function setUser(user: any): void {
  if (typeof window !== 'undefined') {
    localStorage.setItem(USER_KEY, JSON.stringify(user));
  }
}

export function getUser(): any {
  if (typeof window !== 'undefined') {
    const userStr = localStorage.getItem(USER_KEY);
    if (userStr) {
      try {
        return JSON.parse(userStr);
      } catch (e) {
        return null;
      }
    }
  }
  return null;
}

export function removeUser(): void {
  if (typeof window !== 'undefined') {
    localStorage.removeItem(USER_KEY);
  }
}

// token刷新相关
export const refreshState = {
  isRefreshing: false,
  refreshSubscribers: [] as Array<(token: string) => void>,
};

export function onRefreshed(token: string): void {
  refreshState.refreshSubscribers.forEach(cb => cb(token));
  refreshState.refreshSubscribers = [];
}

export function addRefreshSubscriber(cb: (token: string) => void): void {
  refreshState.refreshSubscribers.push(cb);
}

// 后端返回格式
interface RefreshTokenResponse {
  code: number;
  message: string;
  data: {
    accessToken: string;
    refreshToken: string;
    expiresIn: number;
  };
  timestamp: string;
  request_id: string;
}

export async function refreshTokenApi(): Promise<string> {
  const refreshToken = getRefreshToken();
  if (!refreshToken) throw new Error('No refresh token');
  const token = getToken();
  if (!token) throw new Error('No token');
  const res = await axios.post<RefreshTokenResponse>(`${apiBaseURL}/v1/auth/refresh`, null, {
    headers: {
      'Authorization': `Bearer ${token}`,
      'X-Refresh-Token': refreshToken,
    },
  });
  const { code, data, message } = res.data;
  if (code !== 200) throw new Error(message || '刷新token失败');
  setToken(data.accessToken || '', data.refreshToken || '');
  return data.accessToken;
}
