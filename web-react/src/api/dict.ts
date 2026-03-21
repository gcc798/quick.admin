import { request } from '@/utils/request';
import type { PageData, PageQuery } from '@/types/api';
import type { DictRecord } from '@/types/system';

export const dictApi = {
  page: (
    data: PageQuery & { dictType?: string; dictLabel?: string; status?: number },
  ) => request.post<PageData<DictRecord>>('/api/v1/dict/page', data),
  detail: (id: number) => request.get<DictRecord>(`/api/v1/dict/${id}`),
  create: (data: Partial<DictRecord>) => request.post<string>('/api/v1/dict', data),
  update: (id: number, data: Partial<DictRecord>) =>
    request.put<string>(`/api/v1/dict/${id}`, data),
  delete: (id: number) => request.delete<string>(`/api/v1/dict/${id}`),
  batchDelete: (ids: number[]) =>
    request.delete<string>('/api/v1/dict/batch', { data: { ids } }),
  getByType: (dictType: string, parentId?: number) =>
    request.get<DictRecord[]>('/api/v1/dict/type', { params: { dictType, parentId } }),
  getLabel: (dictType: string, dictValue: string) =>
    request.get<string>('/api/v1/dict/label', { params: { dictType, dictValue } }),
};
