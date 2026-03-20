import { request } from '@/utils/request';
import type { User } from '@/types/system';
import type { PageParams, PageResponse } from '@/types/api';

export const userApi = {
  // 获取用户列表（分页）
  list: (params: PageParams & { username?: string; status?: number }) =>
    request.post<PageResponse<User>>('/api/v1/user/page', params),

  // 获取用户详情
  detail: (id: number) => request.get<User>(`/api/v1/user/${id}`),

  // 创建用户
  create: (data: Partial<User>) => request.post('/api/v1/user', data),

  // 更新用户
  update: (id: number, data: Partial<User>) =>
      request.put(`/api/v1/user/${id}`, data),

  // 删除用户
  delete: (id: number) => request.delete(`/api/v1/user/${id}`),

  // 批量删除
  batchDelete: (ids: number[]) =>
    request.delete('/api/v1/user/batch', { data: { ids } }),

  // 重置密码（管理员）
  resetPassword: (id: number, newPassword: string) =>
    request.put(`/api/v1/user/${id}/password`, { newPassword }),

  // 修改密码（用户自己）
  changePassword: (oldPassword: string, newPassword: string) =>
    request.post('/api/v1/user/password/change', { oldPassword, newPassword }),

  // 批量导入用户
  import: (users: Partial<User>[]) =>
    request.post('/api/v1/user/import', { users }),

  // 获取用户角色
  getRoles: (userId: number) =>
    request.get<any[]>(`/api/v1/role/user`, { params: { userId } }),

  // 分配角色（单个）
  assignRole: (userId: number, roleId: number, _orgId?: number) =>
    request.post(`/api/v1/role/assign`, { userId, roleId }),

  // 移除用户角色
  removeRole: (userId: number, roleId: number) =>
    request.delete(`/api/v1/role/remove`, { params: { userId, roleId } }),
};
