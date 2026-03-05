import { request } from '@/utils/request';
import type {
  LoginParams,
  LoginResponse,
  RefreshTokenParams,
  RefreshTokenResponse,
  UserInfo,
} from '@/types/api';

export const authApi = {
  // 登录
  login: (params: LoginParams) => request.post<LoginResponse>('/login', params),

  // 登出
  logout: () => request.post('/logout'),

  // 刷新 Token
  refreshToken: (params: RefreshTokenParams) =>
    request.post<RefreshTokenResponse>('/auth/refresh', params),

  // 获取当前用户信息
  me: () => request.get<{ userId: number }>('/me'),

  // 发送短信验证码
  sendSmsCode: (phonenumber: string) =>
    request.post('/captcha/sms', { phone: phonenumber }),

  // 发送邮箱验证码
  sendEmailCode: (email: string) =>
    request.post('/captcha/email', { email }),
};
