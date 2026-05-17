import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';
import { authApi } from '@/api/auth';
import { DEFAULT_CLIENT_ID } from '@/constants/auth';
import type { AuthLoginData, LoginReq, UserInfo } from '@/types/api';
import { usePermissionStore } from './permission';

interface AuthState {
  accessToken: string;
  refreshToken: string;
  userInfo: UserInfo | null;
  isLoggedIn: boolean;
  login: (payload: LoginReq) => Promise<AuthLoginData>;
  logout: () => Promise<void>;
  clearAuthState: () => void;
  refreshAccessToken: () => Promise<void>;
}

function readString(value: unknown) {
  return typeof value === 'string' ? value : '';
}

function readOptionalString(value: unknown) {
  return typeof value === 'string' ? value : undefined;
}

function readNumberOrString(value: unknown) {
  if (typeof value === 'number' || typeof value === 'string') {
    return value;
  }

  return '';
}

function normalizeUserInfo(raw: unknown): UserInfo {
  const source = (raw ?? {}) as Record<string, unknown>;

  return {
    userId: readNumberOrString(source.userId ?? source.user_id),
    username: readString(source.username ?? source.userName),
    nickname: readString(source.nickname ?? source.nickName),
    phonenumber: readOptionalString(source.phonenumber ?? source.phoneNumber),
    email: readOptionalString(source.email),
    avatar: readOptionalString(source.avatar),
    userType: readNumberOrString(source.userType ?? source.user_type),
  };
}

function normalizeAuthLoginData(raw: unknown): AuthLoginData {
  const source = (raw ?? {}) as Record<string, unknown>;

  return {
    // 实际联调时发现 native 仍然返回下划线字段，这里统一归一化，
    // 让 store 和页面始终只面对 React 工程内部约定的驼峰命名。
    accessToken: readString(source.accessToken ?? source.access_token),
    refreshToken: readString(source.refreshToken ?? source.refresh_token),
    expiresIn: Number(source.expiresIn ?? source.expires_in ?? 0),
    refreshExpiresIn: Number(
      source.refreshExpiresIn ?? source.refresh_expires_in ?? 0,
    ),
    userInfo: normalizeUserInfo(source.userInfo ?? source.user_info),
  };
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      accessToken: '',
      refreshToken: '',
      userInfo: null,
      isLoggedIn: false,
      login: async (payload) => {
        const rawData = await authApi.login(payload);
        const data = normalizeAuthLoginData(rawData);

        // 如果这里拿不到 token，继续放行只会在后续菜单请求里被 401 打回登录页。
        // 直接在登录阶段抛错，用户能更快定位到“登录返回结构不匹配”。
        if (!data.accessToken || !data.refreshToken) {
          throw new Error('登录响应缺少 token 字段，请检查后端返回结构');
        }

        set({
          accessToken: data.accessToken,
          refreshToken: data.refreshToken,
          userInfo: data.userInfo,
          isLoggedIn: true,
        });
        usePermissionStore.getState().reset();
        return data;
      },
      logout: async () => {
        try {
          await authApi.logout();
        } finally {
          get().clearAuthState();
        }
      },
      clearAuthState: () => {
        set({
          accessToken: '',
          refreshToken: '',
          userInfo: null,
          isLoggedIn: false,
        });
        usePermissionStore.getState().reset();
      },
      refreshAccessToken: async () => {
        const { refreshToken } = get();
        const rawData = await authApi.refreshToken({
          refreshToken,
          clientId: DEFAULT_CLIENT_ID,
        });
        const data = normalizeAuthLoginData(rawData);

        if (!data.accessToken || !data.refreshToken) {
          throw new Error('刷新 token 响应缺少必要字段');
        }

        set({
          accessToken: data.accessToken,
          refreshToken: data.refreshToken,
        });
      },
    }),
    {
      name: 'web-react-auth',
      storage: createJSONStorage(() => localStorage),
    },
  ),
);
