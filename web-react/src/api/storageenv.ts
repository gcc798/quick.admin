import { request } from '@/utils/request';
import type { PageData, PageQuery } from '@/types/api';
import type { StorageEnvRecord } from '@/types/system';

export const storageEnvApi = {
  page: (data: PageQuery & { name?: string; storageType?: string }) =>
    request.post<PageData<StorageEnvRecord>>('/api/v1/storage-env/page', data),
  detail: (id: number) => request.get<StorageEnvRecord>(`/api/v1/storage-env/${id}`),
  default: () => request.get<StorageEnvRecord>('/api/v1/storage-env/default'),
  create: (data: Partial<StorageEnvRecord>) =>
    request.post<string>('/api/v1/storage-env', data),
  update: (id: number, data: Partial<StorageEnvRecord>) =>
    request.put<string>(`/api/v1/storage-env/${id}`, data),
  delete: (id: number) => request.delete<string>(`/api/v1/storage-env/${id}`),
  setDefault: (id: number) => request.post<string>('/api/v1/storage-env/default', { id }),
  testConnection: (id: number) => request.post<string>(`/api/v1/storage-env/${id}/test`),
};
