import { request } from '@/utils/request';
import type {
  AuthLoginData,
  CaptchaData,
  CaptchaType,
  LoginReq,
  RefreshTokenReq,
} from '@/types/api';

export const authApi = {
  login: (data: LoginReq) => request.post<AuthLoginData>('/login', data),
  logout: () => request.post<string>('/logout'),
  refreshToken: (data: RefreshTokenReq) =>
    request.post<Omit<AuthLoginData, 'userInfo'>>('/auth/refresh', data),
  getCaptchaTypes: () => request.get<CaptchaType[]>('/captcha/enabled-types'),
  getImageCaptcha: () => request.get<CaptchaData>('/captcha/image'),
};
