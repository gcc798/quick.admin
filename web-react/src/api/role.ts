import { request } from '@/utils/request';
import type { PageData, PageQuery } from '@/types/api';
import type { RoleRecord } from '@/types/system';

export const roleApi = {
  page: (data: PageQuery & { roleName?: string; status?: number }) =>
    request.post<PageData<RoleRecord>>('/api/v1/role/page', data),
  detail: (roleId: number) => request.get<RoleRecord>(`/api/v1/role/${roleId}`),
  create: (data: Partial<RoleRecord>) => request.post<string>('/api/v1/role', data),
  update: (roleId: number, data: Partial<RoleRecord>) =>
    request.put<string>(`/api/v1/role/${roleId}`, data),
  delete: (roleId: number) => request.delete<string>(`/api/v1/role/${roleId}`),
  assignRole: (userId: number, roleId: number) =>
    request.post<string>('/api/v1/role/assign', { userId, roleId }),
  removeRole: (userId: number, roleId: number) =>
    request.delete<string>('/api/v1/role/remove', { params: { userId, roleId } }),
  getUserRoles: (userId: number) =>
    request.get<RoleRecord[]>('/api/v1/role/user', { params: { userId } }),
  getPermissions: (roleKey: string) =>
    request.get<string[]>('/api/v1/role/permissions', { params: { roleKey } }),
  getMenus: (roleId: number) => request.get<number[]>(`/api/v1/role/${roleId}/menus`),
  assignMenus: (roleId: number, menuIds: number[]) =>
    request.post<string>(`/api/v1/role/${roleId}/menus`, { menuIds }),
};
