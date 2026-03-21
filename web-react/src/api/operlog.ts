import { request } from '@/utils/request';
import type { PageData, PageQuery } from '@/types/api';
import type { OperLogRecord } from '@/types/system';

export const operLogApi = {
  page: (
    data: PageQuery & {
      title?: string;
      operName?: string;
      businessType?: string;
      status?: number | string;
      startTime?: string;
      endTime?: string;
    },
  ) => request.post<PageData<OperLogRecord>>('/api/v1/operLog/page', data),
  clean: (days = 30) => request.post<string>('/api/v1/operLog/clean', { days }),
};
