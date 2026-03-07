import { request } from '@/utils/request';
import type { Role } from '@/types/system';
import type { PageParams, PageResponse } from '@/types/api';

export const roleApi = {
  // 获取角色列表（分页）
  list: (params?: PageParams & { roleName?: string; status?: number }) =>
    request.post<PageResponse<Role>>('/api/v1/role/page', params),

  // 获取角色详情
  detail: (roleId: number) => request.get<Role>(`/api/v1/role/${roleId}`),

  // 创建角色
  create: (data: Partial<Role>) => request.post('/api/v1/role', data),

  // 更新角色
  update: (data: Partial<Role>) =>
    request.put(`/api/v1/role/${data.roleId}`, data),

  // 删除角色
  delete: (roleId: number) => request.delete(`/api/v1/role/${roleId}`),

  // 批量删除角色
  batchDelete: (roleIds: number[]) =>
    Promise.all(roleIds.map(id => roleApi.delete(id))),

  // 为用户分配角色
  assignToUser: (userId: number, roleId: number, _orgId?: number) =>
    request.post('/api/v1/role/assign', { userId, roleId }),

  // 移除用户角色
  removeFromUser: (userId: number, roleId: number) =>
    request.delete('/api/v1/role/remove', { params: { userId, roleId } }),

  // 获取用户的所有角色
  getUserRoles: (userId: number) =>
    request.get<Role[]>('/api/v1/role/user', { params: { userId } }),

  // 获取角色权限列表
  getPermissions: (roleKey: string) => 
    request.get<any[]>(`/api/v1/role/permissions`, { params: { roleKey } }),

  // 添加角色权限
  addPermission: (roleKey: string, resource: string, action: string, _orgId?: number) =>
    request.post(`/api/v1/role/permission`, { roleKey, resource, action }),

  // 删除角色权限
  removePermission: (roleKey: string, resource: string, action: string) =>
    request.delete(`/api/v1/role/permission`, { params: { roleKey, resource, action } }),
};
