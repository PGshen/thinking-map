import axios, { AxiosError, AxiosInstance, AxiosRequestConfig, AxiosResponse, InternalAxiosRequestConfig } from 'axios';
import { useGlobalStore } from '@/store/globalStore';
import {
  getToken,
  removeToken,
  refreshTokenApi,
  refreshState,
  onRefreshed,
  addRefreshSubscriber
} from '@/lib/auth';
import { toast } from 'sonner';
import { ApiResponse } from '@/types/response';

const apiBaseURL = process.env.NEXT_PUBLIC_API_BASE_URL || '/api';

const instance: AxiosInstance = axios.create({
  baseURL: apiBaseURL,
  timeout: 15000,
  withCredentials: true,
});

// 扩展 InternalAxiosRequestConfig 以支持 _retry
interface CustomAxiosRequestConfig extends InternalAxiosRequestConfig {
  _retry?: boolean;
  _retryCount?: number;
}

// 泛型请求方法
export async function request<T = any>(config: AxiosRequestConfig): Promise<ApiResponse<T>> {
  const res = await instance.request<ApiResponse<T>>(config);
  return res.data;
}

// 也可导出常用方法
export function get<T = any>(url: string, config?: AxiosRequestConfig) {
  return request<T>({ ...config, method: 'get', url });
}
export function post<T = any>(url: string, data?: any, config?: AxiosRequestConfig) {
  return request<T>({ ...config, method: 'post', url, data });
}
export function put<T = any>(url: string, data?: any, config?: AxiosRequestConfig) {
  return request<T>({ ...config, method: 'put', url, data });
}
export function del<T = any>(url: string, config?: AxiosRequestConfig) {
  return request<T>({ ...config, method: 'delete', url });
}

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
  async (response: AxiosResponse<ApiResponse<any>>) => {
    if (typeof window !== 'undefined') {
      const { setLoading } = useGlobalStore.getState();
      setLoading(false);
    }
    const { code, message, data } = response.data;
    if (code !== 200) {
      if (typeof window !== 'undefined') {
        const { setError } = useGlobalStore.getState();
        setError(message || '请求错误');
        toast.error(message || '请求错误');
      }
      if (code == 401) {
        // 重定向到登录页
        if (typeof window !== 'undefined') {
          window.location.href = '/login';
        }
      } else {
        const originalRequest = response.config as CustomAxiosRequestConfig;
        originalRequest._retryCount = (originalRequest._retryCount || 0) + 1;
        if (originalRequest._retryCount <= 3) {
          await new Promise((r) => setTimeout(r, 500));
          return instance(originalRequest);
        } else {
          if (typeof window !== 'undefined') {
            const { setError } = useGlobalStore.getState();
            setError('服务异常，请稍后再试');
            toast.error('服务异常，请稍后再试');
          }
        }
      }
      return Promise.reject(new Error(message || '请求错误'));
    }
    return response;
  },
  async (error: AxiosError) => {
    const originalRequest = error.config as CustomAxiosRequestConfig | undefined;
    if (
      error.response &&
      error.response.status === 401 &&
      originalRequest &&
      !originalRequest._retry
    ) {
      if (typeof window !== 'undefined') {
        const { setLoading, setError } = useGlobalStore.getState();
        setLoading(false);
      }
      if (refreshState.isRefreshing) {
        // 队列等待token刷新完成
        return new Promise((resolve, reject) => {
          addRefreshSubscriber((token: string) => {
            if (!originalRequest) return reject(new Error('No original request'));
            originalRequest.headers = originalRequest.headers || {};
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
        if (originalRequest) {
          originalRequest.headers = originalRequest.headers || {};
          originalRequest.headers['Authorization'] = 'Bearer ' + newToken;
          return instance(originalRequest);
        }
      } catch (refreshErr) {
        refreshState.isRefreshing = false;
        if (typeof window !== 'undefined') {
          const { setError } = useGlobalStore.getState();
          setError('登录已过期，请重新登录');
          window.location.href = '/login';
        }
        removeToken();
        return Promise.reject(refreshErr);
      }
    }
    if (
      originalRequest &&
      (!error.response || (error.response && error.response.status !== 401))
    ) {
      originalRequest._retryCount = (originalRequest._retryCount || 0) + 1;
      if (originalRequest._retryCount <= 3) {
        await new Promise((r) => setTimeout(r, 500));
        return instance(originalRequest);
      } else {
        if (typeof window !== 'undefined') {
          const { setError } = useGlobalStore.getState();
          setError('服务异常，请稍后再试');
          toast.error('服务异常，请稍后再试');
        }
      }
    }
    if (typeof window !== 'undefined') {
      const { setLoading, setError } = useGlobalStore.getState();
      setLoading(false);
      let message = '请求错误';
      if (error.response && error.response.data) {
        const data = error.response.data as Partial<ApiResponse<any>>;
        message = data.message || error.message || '请求错误';
      } else if (error.message) {
        message = error.message;
      }
      setError(message);
      toast.error(message);
    }
    return Promise.reject(error);
  }
);

export default instance;