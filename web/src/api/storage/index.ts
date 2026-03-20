import { request } from '@/utils/request';
import type { StorageEnv, Attachment } from '@/types/system';
import type { PageParams, PageResponse } from '@/types/api';

// 存储环境 API
export const storageEnvApi = {
  // 获取存储环境列表（分页）
  list: (params?: PageParams & { name?: string; storageType?: string }) =>
    request.post<PageResponse<StorageEnv>>('/api/v1/storage-env/page', params),

  // 获取默认存储环境
  getDefault: () => request.get<StorageEnv>('/api/v1/storage-env/default'),

  // 获取存储环境详情
  detail: (id: number) => request.get<StorageEnv>(`/api/v1/storage-env/${id}`),

  // 创建存储环境
  create: (data: Partial<StorageEnv>) => request.post('/api/v1/storage-env', data),

  // 更新存储环境
  update: (data: Partial<StorageEnv>) =>
    request.put(`/api/v1/storage-env/${data.id}`, data),

  // 删除存储环境
  delete: (id: number) => request.delete(`/api/v1/storage-env/${id}`),

  // 设置默认存储环境
  setDefault: (id: number) =>
      request.post(`/api/v1/storage-env/default`, { id }),

  // 测试存储环境连接
  testConnection: (id: number, data?: any) =>
      request.post(`/api/v1/storage-env/${id}/test`, data),
};

// 附件 API
export const attachmentApi = {
  // 获取附件列表（分页）
  page: (params: { pageNum: number; pageSize: number; fileName?: string; fileType?: string; businessType?: string }) =>
        request.post<PageResponse<Attachment>>('/api/v1/attachment/page', params),

  // 获取附件详情
  detail: (attachmentId: number) =>
    request.get<Attachment>(`/api/v1/attachment/${attachmentId}`),

  // 第一步：上传文件
  uploadFile: (file: File, envCode?: string) => {
    const formData = new FormData();
    formData.append('file', file);
    if (envCode) {
      formData.append('envCode', envCode);
    }
    return request.post<Attachment>('/api/v1/attachment/upload-file', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
  },

  // 第二步：绑定业务信息
  bind: (
    attachmentId: number,
    data: {
      businessType: string;
      businessId: string;
      businessField?: string;
      isPublic?: boolean;
      expireTime?: string;
      metadata?: any;
    },
  ) => request.post(`/api/v1/attachment/${attachmentId}/bind`, data),

  // 根据业务获取附件列表
  getByBusiness: (businessType: string, businessId: string) =>
    request.get<Attachment[]>('/api/v1/attachment/business', { params: { businessType, businessId } }),

  // 删除附件
  delete: (attachmentId: number) =>
    request.delete(`/api/v1/attachment/${attachmentId}`),

  // 批量删除
  batchDelete: (attachmentIds: number[]) =>
    Promise.all(attachmentIds.map(id => attachmentApi.delete(id))),

  // 下载附件
  download: (attachmentId: number) => {
    window.open(`/api/v1/attachment/${attachmentId}/download`, '_blank');
  },

  // 获取附件 URL
  getUrl: (attachmentId: number, expires?: number) =>
    request.get<{ url: string; expires: number }>(`/api/v1/attachment/${attachmentId}/url`, { params: { expires } }),
};
