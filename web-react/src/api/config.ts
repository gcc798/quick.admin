import { request } from '@/utils/request';
import type { PageData, PageQuery, SnowflakeId } from '@/types/api';
import type { ConfigRecord } from '@/types/system';

export const configApi = {
  page: (data: PageQuery & { name?: string; code?: string }) =>
    request.post<PageData<ConfigRecord>>('/api/v1/config/page', data),
  detail: (id: SnowflakeId) => request.get<ConfigRecord>(`/api/v1/config/${id}`),
  getByCode: (code: string) =>
    request.get<ConfigRecord[]>('/api/v1/config/code', { params: { code } }),
  getData: (code: string) =>
    request.get<unknown>('/api/v1/config/data', { params: { code } }),
  create: (data: Partial<ConfigRecord>) => request.post<string>('/api/v1/config', data),
  update: (id: SnowflakeId, data: Partial<ConfigRecord>) =>
    request.put<string>(`/api/v1/config/${id}`, data),
  delete: (id: SnowflakeId) => request.delete<string>(`/api/v1/config/${id}`),
  batchDelete: (ids: SnowflakeId[]) =>
    request.delete<string>('/api/v1/config/batch', { data: { ids } }),
};
