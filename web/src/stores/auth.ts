import { defineStore } from 'pinia';
import { authApi } from '@/api/auth';
import type { LoginParams, LoginResponse, UserInfo } from '@/types/api';

interface AuthState {
  accessToken: string;
  refreshToken: string;
  userInfo: UserInfo | null;
  isLoggedIn: boolean;
}

export const useAuthStore = defineStore('auth', {
  state: (): AuthState => ({
    accessToken: '',
    refreshToken: '',
    userInfo: null,
    isLoggedIn: false,
  }),

  getters: {
    userId: (state) => state.userInfo?.userId,
    username: (state) => state.userInfo?.username,
    nickname: (state) => state.userInfo?.nickname,
  },

  actions: {
    // 登录
    async login(credentials: LoginParams): Promise<LoginResponse> {
      try {
        const res = await authApi.login(credentials);
        
        // 适配后端返回的下划线命名
        this.accessToken = (res as any).access_token || res.accessToken;
        this.refreshToken = (res as any).refresh_token || res.refreshToken;
        this.userInfo = (res as any).user_info || res.userInfo;
        this.isLoggedIn = true;
        
        return res;
      } catch (error) {
        throw error;
      }
    },

    // 登出
    async logout() {
      try {
        await authApi.logout();
      } catch (error) {
        // 忽略登出接口错误（例如网络错误、服务器未启动等）
        console.warn('Logout API call failed, clearing local state anyway:', error);
      } finally {
        this.clearAuthState();
      }
    },

    // 清除本地认证状态（不调用后端接口）
    clearAuthState() {
      this.accessToken = '';
      this.refreshToken = '';
      this.userInfo = null;
      this.isLoggedIn = false;
      
      // 清除权限和路由
      import('./permission').then(({ usePermissionStore }) => {
        const permissionStore = usePermissionStore();
        permissionStore.resetRoutes();
      });
    },

    // 刷新 Token
    async refreshAccessToken() {
      try {
        const res = await authApi.refreshToken({
          refreshToken: this.refreshToken,
          clientKey: import.meta.env.VITE_CLIENT_KEY,
          clientSecret: import.meta.env.VITE_CLIENT_SECRET,
        });
        // 适配后端返回的下划线命名
        this.accessToken = (res as any).access_token || res.accessToken;
        this.refreshToken = (res as any).refresh_token || res.refreshToken;
      } catch (error) {
        // 刷新失败，清除状态
        this.accessToken = '';
        this.refreshToken = '';
        this.userInfo = null;
        this.isLoggedIn = false;
        throw error;
      }
    },
  },

  persist: {
    key: 'auth',
    storage: localStorage,
    paths: ['accessToken', 'refreshToken', 'userInfo', 'isLoggedIn'],
  },
});
