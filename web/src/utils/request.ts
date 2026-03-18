import axios, { type AxiosInstance, type AxiosRequestConfig, type AxiosResponse } from 'axios';
import { message } from 'ant-design-vue';
import JSONbig from 'json-bigint';
import { useAuthStore } from '@/stores/auth';
import router from '@/router';
import type { ApiResponse } from '@/types/api';

// 创建 JSONbig 实例，将大数转换为字符串
const JSONbigString = JSONbig({ storeAsString: true });

function readRequestPayload(config: AxiosRequestConfig): Record<string, any> {
  const raw = config.data;
  if (!raw) {
    return {};
  }
  if (typeof raw === 'string') {
    try {
      return JSON.parse(raw);
    } catch {
      return {};
    }
  }
  if (typeof raw === 'object') {
    return raw as Record<string, any>;
  }
  return {};
}

// 创建 axios 实例
const service: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json;charset=UTF-8',
  },
  // 使用 json-bigint 来解析响应，避免大数精度丢失
  transformResponse: [
    (data) => {
      if (typeof data === 'string') {
        try {
          return JSONbigString.parse(data);
        } catch (err) {
          return data;
        }
      }
      return data;
    },
  ],
});

// 请求拦截器
service.interceptors.request.use(
  (config) => {
    const authStore = useAuthStore();
    
    // 添加 Token
    if (authStore.accessToken) {
      config.headers.Authorization = `Bearer ${authStore.accessToken}`;
    }
    
    return config;
  },
  (error) => {
    console.error('Request error:', error);
    return Promise.reject(error);
  },
);

// 响应拦截器
service.interceptors.response.use(
  async (response: AxiosResponse<ApiResponse>) => {
    if (response.config.responseType === 'blob' || response.config.responseType === 'arraybuffer') {
      return response.data as any;
    }

    const { code, data, msg, rows, total } = (response.data || {}) as any;
    
    // 成功响应
    if (code === 200) {
      if (typeof rows !== 'undefined' || typeof total !== 'undefined') {
        const payload = readRequestPayload(response.config);
        const pageSize = Number(payload.pageSize) || 0;
        const current = Number(payload.pageNum) || 1;
        const totalCount = Number(total) || 0;
        const size = pageSize > 0 ? pageSize : Array.isArray(rows) ? rows.length : 0;
        const pages = size > 0 ? Math.ceil(totalCount / size) : 0;
        return {
          records: rows || [],
          total: totalCount,
          size,
          current,
          pages,
        };
      }
      return data;
    }
    
    // Token 过期，尝试刷新（但排除登录和刷新接口）
    if (code === 401 && !response.config.url?.includes('/login') && !response.config.url?.includes('/refresh')) {
      const authStore = useAuthStore();
      
      // 如果没有 refreshToken，直接跳转登录页
      if (!authStore.refreshToken) {
        authStore.clearAuthState(); // 使用本地清除
        router.push('/login');
        message.error('登录已过期，请重新登录');
        return Promise.reject(new Error('No refresh token'));
      }
      
      try {
        // 刷新 Token
        await authStore.refreshAccessToken();
        
        // 重试原请求
        const config = response.config;
        config.headers.Authorization = `Bearer ${authStore.accessToken}`;
        return service.request(config);
      } catch (error) {
        // 刷新失败，跳转登录页
        authStore.clearAuthState(); // 使用本地清除
        router.push('/login');
        message.error('登录已过期，请重新登录');
        return Promise.reject(new Error('Token refresh failed'));
      }
    }
    
    // 如果是刷新接口返回 401，直接跳转登录页
    // if (code === 401 && response.config.url?.includes('/refresh')) {
    if (code === 401) {
      const authStore = useAuthStore();
      authStore.clearAuthState();
      router.push('/login');
      message.error('登录已过期，请重新登录');
      return Promise.reject(new Error('Refresh token expired'));
    }
    
    // 其他错误 - 将错误信息附加到 Error 对象上，供调用方使用
    const error = new Error(msg || 'Request failed') as any;
    error.code = code;
    error.message = msg || 'Request failed';
    // 不在这里显示错误消息，让调用方决定如何处理
    // message.error(msg || '请求失败');
    return Promise.reject(error);
  },
  (error) => {
    console.error('Response error:', error);
    
    if (error.response) {
      const { status } = error.response;
      
      switch (status) {
        case 401:
          // 如果是登录或刷新接口的 401，不做特殊处理
          if (!error.config?.url?.includes('/login') && !error.config?.url?.includes('/refresh')) {
            message.error('未授权，请重新登录');
            const authStore = useAuthStore();
            authStore.clearAuthState(); // 使用本地清除，避免调用后端
            router.push('/login');
          }
          break;
        case 403:
          message.error('拒绝访问');
          break;
        case 404:
          message.error('请求的资源不存在');
          break;
        case 500:
          message.error('服务器错误');
          break;
        default:
          message.error(error.response.data?.msg || '请求失败');
      }
    } else if (error.request) {
      // 网络错误（例如后端服务未启动）
      message.error('网络错误，请检查后端服务是否启动');
      // 不要在这里调用 logout，避免无限循环
    } else {
      message.error('请求配置错误');
    }
    
    return Promise.reject(error);
  },
);

// 封装请求方法
export const request = {
  get<T = any>(url: string, config?: AxiosRequestConfig): Promise<T> {
    return service.get(url, config);
  },
  
  post<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    return service.post(url, data, config);
  },
  
  put<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    return service.put(url, data, config);
  },
  
  delete<T = any>(url: string, config?: AxiosRequestConfig): Promise<T> {
    return service.delete(url, config);
  },
};

export default service;
