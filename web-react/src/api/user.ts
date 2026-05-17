import { request } from '@/utils/request';
import type { PageData, PageQuery, SnowflakeId } from '@/types/api';
import type { UserFormData, UserRecord } from '@/types/system';

type RawUserRecord = Partial<UserRecord> & {
  id?: SnowflakeId;
  userId?: SnowflakeId;
  userName?: string;
  username?: string;
  nickName?: string;
  nickname?: string;
};

function normalizeUserRecord(record: RawUserRecord): UserRecord {
  return {
    ...record,
    id: record.id ?? record.userId ?? '',
    userName: record.userName ?? record.username ?? '',
    nickName: record.nickName ?? record.nickname ?? '',
  };
}

export const userApi = {
  page: async (data: PageQuery & { username?: string; phonenumber?: string; status?: number }) => {
    const result = await request.post<PageData<RawUserRecord>>('/api/v1/user/page', data);
    return {
      ...result,
      records: (result.records ?? []).map(normalizeUserRecord),
    };
  },
  detail: async (id: SnowflakeId) =>
    normalizeUserRecord(await request.get<RawUserRecord>(`/api/v1/user/${id}`)),
  create: (data: UserFormData) => request.post<string>('/api/v1/user', data),
  update: (id: SnowflakeId, data: Partial<UserFormData>) =>
    request.put<string>(`/api/v1/user/${id}`, data),
  delete: (id: SnowflakeId) => request.delete<string>(`/api/v1/user/${id}`),
  batchDelete: (ids: SnowflakeId[]) =>
    request.delete<string>('/api/v1/user/batch', { data: { ids } }),
  resetPassword: (id: SnowflakeId, newPassword: string) =>
    request.put<string>(`/api/v1/user/${id}/password`, { newPassword }),
};
