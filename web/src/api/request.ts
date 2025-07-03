import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse, InternalAxiosRequestConfig } from 'axios';
import { useGlobalStore } from '@/store/globalStore';
import {
  getToken,
  getRefreshToken,
  setToken,
  removeToken,
  refreshTokenApi,
  refreshState,
  onRefreshed,
  addRefreshSubscriber
} from '@/lib/auth';

const apiBaseURL = process.env.NEXT_PUBLIC_API_BASE_URL || '/api';

const instance: AxiosInstance = axios.create({
  baseURL: apiBaseURL,
  timeout: 15000,
  withCredentials: true,
});

// 请求拦截器：自动附加token、设置全局loading
instance.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = getToken();
    if (token && config.headers) {
      config.headers['Authorization'] = `Bearer ${token}`;
    }
    // 可选：全局loading
    if (typeof window !== 'undefined') {
      const { setLoading } = useGlobalStore.getState();
      setLoading(true);
    }
    return config;
  },
  (error: any) => {
    if (typeof window !== 'undefined') {
      const { setLoading, setError } = useGlobalStore.getState();
      setLoading(false);
      setError(error.message);
    }
    return Promise.reject(error);
  }
);

// 响应拦截器：统一处理后端响应格式、全局loading/error
instance.interceptors.response.use(
  (response: AxiosResponse) => {
    if (typeof window !== 'undefined') {
      const { setLoading } = useGlobalStore.getState();
      setLoading(false);
    }
    // 假设后端返回 { code, message, data }
    const { code, message, data } = response.data;
    if (code !== 200) {
      if (typeof window !== 'undefined') {
        const { setError } = useGlobalStore.getState();
        setError(message || '请求错误');
      }
      return Promise.reject(new Error(message || '请求错误'));
    }
    return data;
  },
  async (error: any) => {
    const originalRequest = error.config;
    if (error.response && error.response.status === 401 && !originalRequest._retry) {
      if (typeof window !== 'undefined') {
        const { setLoading, setError } = useGlobalStore.getState();
        setLoading(false);
      }
      if (refreshState.isRefreshing) {
        // 队列等待token刷新完成
        return new Promise((resolve, reject) => {
          addRefreshSubscriber((token: string) => {
            originalRequest.headers['Authorization'] = 'Bearer ' + token;
            originalRequest._retry = true;
            resolve(instance(originalRequest));
          });
        });
      }
      originalRequest._retry = true;
      refreshState.isRefreshing = true;
      try {
        const newToken = await refreshTokenApi();
        onRefreshed(newToken);
        refreshState.isRefreshing = false;
        originalRequest.headers['Authorization'] = 'Bearer ' + newToken;
        return instance(originalRequest);
      } catch (refreshErr) {
        refreshState.isRefreshing = false;
        if (typeof window !== 'undefined') {
          const { setError } = useGlobalStore.getState();
          setError('登录已过期，请重新登录');
        }
        removeToken();
        return Promise.reject(refreshErr);
      }
    }
    if (typeof window !== 'undefined') {
      const { setLoading, setError } = useGlobalStore.getState();
      setLoading(false);
      setError(error.message);
    }
    return Promise.reject(error);
  }
);

export default instance; 