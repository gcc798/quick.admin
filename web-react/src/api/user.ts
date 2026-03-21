import { request } from '@/utils/request';
import type { PageData, PageQuery } from '@/types/api';
import type { UserFormData, UserRecord } from '@/types/system';

export const userApi = {
  page: (data: PageQuery & { username?: string; phonenumber?: string; status?: number }) =>
    request.post<PageData<UserRecord>>('/api/v1/user/page', data),
  detail: (id: number) => request.get<UserRecord>(`/api/v1/user/${id}`),
  create: (data: UserFormData) => request.post<string>('/api/v1/user', data),
  update: (id: number, data: Partial<UserFormData>) =>
    request.put<string>(`/api/v1/user/${id}`, data),
  delete: (id: number) => request.delete<string>(`/api/v1/user/${id}`),
  batchDelete: (ids: number[]) =>
    request.delete<string>('/api/v1/user/batch', { data: { ids } }),
  resetPassword: (id: number, newPassword: string) =>
    request.put<string>(`/api/v1/user/${id}/password`, { newPassword }),
};
