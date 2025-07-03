// 与后端 dto/auth.go 对齐的用户类型定义
export interface AuthData {
  userId?: string;
  username?: string;
  email?: string;
  fullName?: string;
  accessToken?: string;
  refreshToken?: string;
  expiresIn?: number;
}

export interface TokenInfoDTO {
  userId: string;
  username: string;
  accessToken: string;
  expiresAt: string;
}

export interface User {
  id: string;
  name: string;
  // 可根据实际需求扩展属性
} 