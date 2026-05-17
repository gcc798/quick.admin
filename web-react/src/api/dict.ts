import { request } from '@/utils/request';
import type { PageData, PageQuery, SnowflakeId } from '@/types/api';
import type { DictRecord } from '@/types/system';

export const dictApi = {
  page: (
    data: PageQuery & { dictType?: string; dictLabel?: string; status?: number },
  ) => request.post<PageData<DictRecord>>('/api/v1/dict/page', data),
  detail: (id: SnowflakeId) => request.get<DictRecord>(`/api/v1/dict/${id}`),
  create: (data: Partial<DictRecord>) => request.post<string>('/api/v1/dict', data),
  update: (id: SnowflakeId, data: Partial<DictRecord>) =>
    request.put<string>(`/api/v1/dict/${id}`, data),
  delete: (id: SnowflakeId) => request.delete<string>(`/api/v1/dict/${id}`),
  batchDelete: (ids: SnowflakeId[]) =>
    request.delete<string>('/api/v1/dict/batch', { data: { ids } }),
  getByType: (dictType: string, parentId?: SnowflakeId) =>
    request.get<DictRecord[]>('/api/v1/dict/type', { params: { dictType, parentId } }),
  getLabel: (dictType: string, dictValue: string) =>
    request.get<string>('/api/v1/dict/label', { params: { dictType, dictValue } }),
};
