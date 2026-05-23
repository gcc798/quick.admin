import { request } from '@/utils/request';
import type { SnowflakeId } from '@/types/api';
import type { MenuRecord } from '@/types/menu';
import { normalizeMenuRecord, normalizeMenuTree } from '@/utils/menu';

export const menuApi = {
  getUserMenuTree: () =>
    request.get<MenuRecord[]>('/api/v1/menu/user/tree').then(normalizeMenuTree),
  getMenuTree: () =>
    request.get<MenuRecord[]>('/api/v1/menu/tree').then(normalizeMenuTree),
  getMenuList: () =>
    request.get<MenuRecord[]>('/api/v1/menu').then(normalizeMenuTree),
  detail: (id: SnowflakeId) =>
    request.get<MenuRecord>(`/api/v1/menu/${id}`).then(normalizeMenuRecord),
  create: (data: Partial<MenuRecord>) => request.post<string>('/api/v1/menu', data),
  update: (id: SnowflakeId, data: Partial<MenuRecord>) =>
    request.put<string>(`/api/v1/menu/${id}`, data),
  delete: (id: SnowflakeId) => request.delete<string>(`/api/v1/menu/${id}`),
};
