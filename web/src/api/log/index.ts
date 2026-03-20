import { request } from '@/utils/request';
import type { PageParams, PageResponse } from '@/types/api';

export interface LoginLog {
  id: number;
  userName: string;
  ipaddr: string;
  loginLocation?: string;
  browser?: string;
  os?: string;
  status: number;
  msg?: string;
  loginTime: string;
  clientId?: string;
  tenantId?: string;
}

export interface OperLog {
  id: number;
  title: string;
  businessType: string;
  method: string;
  requestMethod: string;
  operatorType: string;
  operName: string;
  operUrl: string;
  operIp: string;
  operLocation?: string;
  operParam?: string;
  jsonResult?: string;
  status: string;
  errorMsg?: string;
  operTime: string;
  costTime?: number;
  userAgent?: string;
}

export const logApi = {
  // 获取登录日志列表
  loginLogPage: (params: PageParams & { userName?: string; ipaddr?: string; status?: number | null; startTime?: string; endTime?: string }) =>
    request.post<PageResponse<LoginLog>>('/api/v1/loginLog/page', params),

  // 获取操作日志列表
  operLogPage: (params: PageParams & { title?: string; operName?: string; businessType?: string; status?: string | null; startTime?: string; endTime?: string }) =>
    request.post<PageResponse<OperLog>>('/api/v1/operLog/page', params),

  // 获取登录日志详情
  loginLogDetail: (id: number) =>
    request.get<LoginLog>(`/api/v1/loginLog/${id}`),

  // 获取操作日志详情
  operLogDetail: (id: number) =>
    request.get<OperLog>(`/api/v1/operLog/${id}`),

  // 清空登录日志
  cleanLoginLog: (params: { days: number } = { days: 30 }) =>
    request.post('/api/v1/loginLog/clean', params),

  // 清空操作日志
  cleanOperLog: (params: { days: number } = { days: 30 }) =>
    request.post('/api/v1/operLog/clean', params),

  // 删除登录日志
  deleteLoginLog: (ids: number[]) =>
    request.delete('/api/v1/loginLog/batch', { data: { ids } }),

  // 删除操作日志
  deleteOperLog: (ids: number[]) =>
    request.delete('/api/v1/operLog/batch', { data: { ids } }),
};
