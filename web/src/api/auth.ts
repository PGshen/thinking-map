/*
 * @Date: 2025-07-03 21:35:45
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-07-04 00:28:29
 * @FilePath: /thinking-map/web/src/api/auth.ts
 */
import { post } from "./request";
import type { RegisterParams, RegisterResponse, LoginParams, LoginResponse, RefreshTokenResponse } from "@/types/auth";
import type { ApiResponse } from "@/types/response";
import { AuthData, RefreshTokenData } from "@/types/auth";
import { getRefreshToken } from "@/lib/auth";
import { API_ENDPOINTS } from "@/api/endpoints";

// 注册
export async function registerUser(
  params: RegisterParams
): Promise<RegisterResponse> {
  return await post<AuthData>(API_ENDPOINTS.AUTH.REGISTER, params);
}

// 登录
export async function loginUser(
  params: LoginParams
): Promise<LoginResponse> {
  return await post<AuthData>(API_ENDPOINTS.AUTH.LOGIN, params);
}

// 刷新Token
export async function refreshToken(refreshToken: string): Promise<RefreshTokenResponse> {
  return await post<RefreshTokenData>(
    API_ENDPOINTS.AUTH.REFRESH,
    {},
    {
      headers: {
        Authorization: `Bearer ${refreshToken}`,
      },
    }
  );
}

// 登出
export async function logout(): Promise<ApiResponse<null>> {
  return await post<null>(
    API_ENDPOINTS.AUTH.LOGOUT,
    {},
    {
      headers: {
        "X-Refresh-Token": getRefreshToken()
      }
    }
  );
} 