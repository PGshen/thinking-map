// 与后端 dto/auth.go 对齐的用户类型定义
export interface AuthData {
  userID?: string;
  username?: string;
  email?: string;
  fullName?: string;
  accessToken?: string;
  refreshToken?: string;
  expiresIn?: number;
}

export interface TokenInfoDTO {
  userID: string;
  username: string;
  accessToken: string;
  expiresAt: string;
}

export interface User {
  userID: string;
  username: string;
  email: string;
  fullName: string;
} 