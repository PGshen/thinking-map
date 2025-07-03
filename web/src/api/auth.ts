/*
 * @Date: 2025-07-03 21:35:45
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-07-04 00:28:29
 * @FilePath: /thinking-map/web/src/api/auth.ts
 */
import instance from "./request";
import type { RegisterParams, RegisterResponse, LoginParams, LoginResponse, RefreshTokenResponse } from "@/types/auth";
import type { ApiResponse } from "@/types/response";

// 注册
export async function registerUser(
  params: RegisterParams
): Promise<RegisterResponse> {
  const res = await instance.post<RegisterResponse>("/v1/auth/register", params);
  return res.data;
}

// 登录
export async function loginUser(
  params: LoginParams
): Promise<LoginResponse> {
  const res = await instance.post<LoginResponse>("/v1/auth/login", params);
  return res.data;
}

// 刷新Token
export async function refreshToken(refreshToken: string): Promise<RefreshTokenResponse> {
  const res = await instance.post<RefreshTokenResponse>(
    "/v1/auth/refresh",
    {},
    {
      headers: {
        Authorization: `Bearer ${refreshToken}`,
      },
    }
  );
  return res.data;
}

// 登出
export async function logout(accessToken: string): Promise<ApiResponse<null>> {
  const res = await instance.post<ApiResponse<null>>(
    "/v1/auth/logout"
  );
  return res.data;
} 