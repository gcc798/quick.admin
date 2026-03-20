// API 响应基础类型
export interface ApiResponse<T = any> {
  code: number;
  msg: string;
  data: T;
}

// 分页响应（统一格式，对应后端 pagination.Page）
export interface PageResponse<T = any> {
  records: T[];      // 数据列表
  total: number;     // 总记录数
  size: number;      // 每页显示条数
  current: number;   // 当前页
  pages: number;     // 总页数
}

// 分页请求参数
export interface PageParams {
  pageNum?: number;
  pageSize?: number;
}

// 用户信息
export interface UserInfo {
  userId: number;
  username: string;
  nickname: string;
  email?: string;
  phonenumber?: string;
  avatar?: string;
  userType: string;
}

// 登录参数
export interface LoginParams {
  grantType: 'password' | 'email' | 'xcx';
  username?: string;
  password?: string;
  email?: string;
  code?: string;
  uuid?: string;
  wxCode?: string;
  phonenumber?: string;
  clientKey: string;
  clientSecret: string;
}

// 登录响应
export interface LoginResponse {
  accessToken: string;
  refreshToken: string;
  expiresIn: number;
  refreshExpiresIn: number;
  userInfo: UserInfo;
}

// 刷新 Token 参数
export interface RefreshTokenParams {
  refreshToken: string;
  clientKey: string;
  clientSecret: string;
}

// 刷新 Token 响应
export interface RefreshTokenResponse {
  accessToken: string;
  refreshToken: string;
  expiresIn: number;
  refreshExpiresIn: number;
}

// 验证码数据
export interface CaptchaData {
  id: string;
  type: string;
  data: Record<string, any>;
  expireAt: string;
}

// 验证码类型
export type CaptchaType = 'image' | 'sms' | 'email';
