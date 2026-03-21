export interface CommonResp<T = unknown> {
  code: number;
  msg: string;
  data?: T;
}

export interface PageData<T = unknown> {
  records: T[];
  total: number;
  size: number;
  current: number;
  pages: number;
}

export interface LoginReq {
  clientKey: string;
  clientSecret: string;
  grantType: 'password' | 'email' | 'xcx';
  username?: string;
  password?: string;
  code?: string;
  phonenumber?: string;
  email?: string;
  wxCode?: string;
  uuid?: string;
}

export interface RefreshTokenReq {
  refreshToken: string;
  clientKey: string;
  clientSecret: string;
}

export interface UserInfo {
  userId: number;
  username: string;
  nickname: string;
  phonenumber?: string;
  email?: string;
  avatar?: string;
  userType: number | string;
}

export interface AuthLoginData {
  accessToken: string;
  refreshToken: string;
  expiresIn: number;
  refreshExpiresIn: number;
  userInfo: UserInfo;
}

export interface CaptchaData {
  id: string;
  type: string;
  data: Record<string, unknown>;
  expireAt: string;
}

export type CaptchaType = 'image' | 'sms' | 'email';

export interface PageQuery {
  pageNum?: number;
  pageSize?: number;
}
