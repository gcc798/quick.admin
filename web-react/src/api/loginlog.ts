import { request } from '@/utils/request';
import type { PageData, PageQuery } from '@/types/api';
import type { LoginLogRecord } from '@/types/system';

export const loginLogApi = {
  page: (
    data: PageQuery & {
      userName?: string;
      ipaddr?: string;
      status?: number;
      startTime?: string;
      endTime?: string;
    },
  ) => request.post<PageData<LoginLogRecord>>('/api/v1/loginLog/page', data),
  clean: (days = 30) => request.post<string>('/api/v1/loginLog/clean', { days }),
};
