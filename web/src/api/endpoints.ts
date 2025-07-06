// 统一管理后端 API 地址
// 可根据实际后端路由结构进行分组

const API_BASE = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://127.0.0.1:8080/api';

export const API_ENDPOINTS = {
  AUTH: {
    REGISTER: `${API_BASE}/v1/auth/register`,
    LOGIN: `${API_BASE}/v1/auth/login`,
    REFRESH: `${API_BASE}/v1/auth/refresh`,
    LOGOUT: `${API_BASE}/v1/auth/logout`,
  },
  THINKING: {
    REPEAT: `${API_BASE}/v1/thinking/repeat`,
    UNDERSTANDING: `${API_BASE}/v1/thinking/understanding`
  },
  MAP: {
    CREATE: `${API_BASE}/v1/maps`,
  },
  // 其他模块分组
};

export default API_ENDPOINTS; 