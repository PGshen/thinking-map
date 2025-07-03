import instance from "./request";

// 注册
export async function registerUser(params: {
  username: string;
  email: string;
  password: string;
  fullName: string;
}) {
  return instance.post("/v1/auth/register", params);
}

// 登录
export async function loginUser(params: {
  username: string;
  password: string;
}) {
  return instance.post("/v1/auth/login", params);
} 