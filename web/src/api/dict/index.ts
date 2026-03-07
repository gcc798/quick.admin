import { request } from '@/utils/request';
import type { PageParams, PageResponse } from '@/types/api';

export interface Dict {
  id: number;
  dictType: string;
  dictLabel: string;
  dictValue: string;
  sort: number;
  parentId?: number;
  isDefault?: boolean;
  status: number;
  remark?: string;
  createBy?: number;
  updateBy?: number;
  createdTime?: string;
  updatedTime?: string;
}

export const dictApi = {
  // 获取字典列表（分页）
  page: (params: PageParams & { dictType?: string; dictLabel?: string; status?: number }) =>
    request.post<PageResponse<Dict>>('/api/v1/dict/page', params),

  // 根据字典类型获取字典数据
  getByType: (dictType: string, parentId?: number) =>
    request.get<Dict[]>(`/api/v1/dict/type`, { params: { dictType, parentId } }),

  // 根据字典类型和键值获取标签
  getLabel: (dictType: string, dictValue: string) =>
    request.get<string>('/api/v1/dict/label', { params: { dictType, dictValue } }),

  // 获取字典详情
  detail: (id: number) =>
    request.get<Dict>(`/api/v1/dict/${id}`),

  // 创建字典
  create: (data: Partial<Dict>) =>
    request.post('/api/v1/dict', data),

  // 更新字典
  update: (data: Partial<Dict>) =>
    request.put(`/api/v1/dict/${data.id}`, data),

  // 删除字典
  delete: (id: number) =>
    request.delete(`/api/v1/dict/${id}`),

  // 批量删除字典
  batchDelete: (ids: number[]) =>
    request.delete('/api/v1/dict/batch', { data: { ids } }),
};
