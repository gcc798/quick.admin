import { request } from '@/utils/request';
import type { MenuRecord } from '@/types/menu';

export const menuApi = {
  getUserMenuTree: () => request.get<MenuRecord[]>('/api/v1/menu/user/tree'),
  getMenuTree: () => request.get<MenuRecord[]>('/api/v1/menu/tree'),
  getMenuList: () => request.get<MenuRecord[]>('/api/v1/menu'),
  detail: (id: number) => request.get<MenuRecord>(`/api/v1/menu/${id}`),
  create: (data: Partial<MenuRecord>) => request.post<string>('/api/v1/menu', data),
  update: (id: number, data: Partial<MenuRecord>) =>
    request.put<string>(`/api/v1/menu/${id}`, data),
  delete: (id: number) => request.delete<string>(`/api/v1/menu/${id}`),
};
