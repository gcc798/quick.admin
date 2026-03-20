import { request } from '@/utils/request';
import type { PageParams, PageResponse } from '@/types/api';

export interface Attachment {
  attachmentId: number;
  fileName: string;
  fileSize: number;
  fileType: string;
  filePath: string;
  storageEnvId: number;
  businessType?: string;
  businessId?: string;
  uploadBy: number;
  uploadTime: string;
  remark?: string;
}

export interface UploadResponse {
  attachmentId: number;
  fileName: string;
  fileSize: number;
  fileType: string;
  filePath: string;
}

export const attachmentApi = {
  // 上传文件
  uploadFile: (file: File, envCode?: string) => {
    const formData = new FormData();
    formData.append('file', file);
    if (envCode) {
      formData.append('envCode', envCode);
    }
    return request.post<UploadResponse>('/api/v1/attachment/upload-file', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
  },

  // 绑定业务
  bind: (attachmentId: number, businessType: string, businessId: string) =>
    request.post(`/api/v1/attachment/${attachmentId}/bind`, { businessType, businessId }),

  // 根据业务获取附件列表
  getByBusiness: (businessType: string, businessId: string) =>
    request.get<Attachment[]>('/api/v1/attachment/business', { params: { businessType, businessId } }),

  // 获取附件列表（分页）（注意：此接口在swagger中不存在，可能需要后端添加）
  page: (params: PageParams & { fileName?: string; fileType?: string; businessType?: string }) =>
      request.post<PageResponse<Attachment>>('/api/v1/attachment/page', params),

  // 获取附件详情
  detail: (attachmentId: number) =>
    request.get<Attachment>(`/api/v1/attachment/${attachmentId}`),

  // 获取文件访问URL
  getUrl: (attachmentId: number, expires?: number) =>
    request.get<{ url: string }>(`/api/v1/attachment/${attachmentId}/url`, { params: { expires } }),

  // 下载文件
  download: (attachmentId: number) =>
    request.get(`/api/v1/attachment/${attachmentId}/download`, { responseType: 'blob' }),

  // 删除附件
  delete: (attachmentId: number) =>
    request.delete(`/api/v1/attachment/${attachmentId}`),

  // 批量删除附件
  batchDelete: (attachmentIds: number[]) =>
    Promise.all(attachmentIds.map(id => attachmentApi.delete(id))),
};
