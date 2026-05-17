import { request } from '@/utils/request';
import type { PageData, PageQuery, SnowflakeId } from '@/types/api';
import type { AttachmentRecord } from '@/types/system';

export const attachmentApi = {
  page: (
    data: PageQuery & { fileName?: string; fileType?: string; businessType?: string },
  ) => request.post<PageData<AttachmentRecord>>('/api/v1/attachment/page', data),
  detail: (attachmentId: SnowflakeId) =>
    request.get<AttachmentRecord>(`/api/v1/attachment/${attachmentId}`),
  getUrl: (attachmentId: SnowflakeId, expires?: number) =>
    request.get<{ url: string; expires?: number }>(`/api/v1/attachment/${attachmentId}/url`, {
      params: { expires },
    }),
  uploadFile: (file: File, envCode?: string) => {
    const formData = new FormData();
    formData.append('file', file);
    if (envCode) {
      formData.append('envCode', envCode);
    }
    return request.post<string>('/api/v1/attachment/upload-file', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    });
  },
  getByBusiness: (businessType: string, businessId: string) =>
    request.get<AttachmentRecord[]>('/api/v1/attachment/business', {
      params: { businessType, businessId },
    }),
  delete: (attachmentId: SnowflakeId) =>
    request.delete<string>(`/api/v1/attachment/${attachmentId}`),
  batchDelete: (attachmentIds: SnowflakeId[]) =>
    Promise.all(attachmentIds.map((attachmentId) => attachmentApi.delete(attachmentId))),
  download: (attachmentId: SnowflakeId) => {
    window.open(
      `${import.meta.env.VITE_API_BASE_URL}/api/v1/attachment/${attachmentId}/download`,
      '_blank',
      'noopener,noreferrer',
    );
  },
};
