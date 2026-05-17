import { request } from '@/utils/request';
import type { PageData, PageQuery, SnowflakeId } from '@/types/api';
import type { RoleRecord, UserRecord } from '@/types/system';

type RawRoleRecord = Partial<RoleRecord> & { id?: SnowflakeId; roleId?: SnowflakeId };
type RawUserRecord = Partial<UserRecord> & {
  id?: SnowflakeId;
  userId?: SnowflakeId;
  userName?: string;
  username?: string;
  nickName?: string;
  nickname?: string;
};

function normalizeRoleRecord(record: RawRoleRecord): RoleRecord {
  const id = record.id ?? record.roleId ?? '';

  return {
    ...record,
    id,
    roleName: record.roleName ?? '',
    roleKey: record.roleKey ?? '',
  };
}

function normalizeUserRecord(record: RawUserRecord): UserRecord {
  return {
    ...record,
    id: record.id ?? record.userId ?? '',
    userName: record.userName ?? record.username ?? '',
    nickName: record.nickName ?? record.nickname ?? '',
  };
}

export const roleApi = {
  page: async (data: PageQuery & { roleName?: string; status?: number }) => {
    const result = await request.post<PageData<RawRoleRecord>>('/api/v1/role/page', data);
    return {
      ...result,
      records: (result.records ?? []).map(normalizeRoleRecord),
    };
  },
  detail: async (roleId: SnowflakeId) =>
    normalizeRoleRecord(await request.get<RawRoleRecord>(`/api/v1/role/${roleId}`)),
  create: (data: Partial<RoleRecord>) => request.post<string>('/api/v1/role', data),
  update: (roleId: SnowflakeId, data: Partial<RoleRecord>) =>
    request.put<string>(`/api/v1/role/${roleId}`, data),
  delete: (roleId: SnowflakeId) => request.delete<string>(`/api/v1/role/${roleId}`),
  assignRole: (userId: SnowflakeId, roleId: SnowflakeId) =>
    request.post<string>('/api/v1/role/assign', { userId, roleId }),
  removeRole: (userId: SnowflakeId, roleId: SnowflakeId) =>
    request.delete<string>('/api/v1/role/remove', { params: { userId, roleId } }),
  getUserRoles: async (userId: SnowflakeId) =>
    (await request.get<RawRoleRecord[]>('/api/v1/role/user', { params: { userId } })).map(
      normalizeRoleRecord,
    ),
  getRoleUsers: async (roleId: SnowflakeId) =>
    (await request.get<RawUserRecord[]>(`/api/v1/role/${roleId}/users`)).map(
      normalizeUserRecord,
    ),
  assignUsers: (roleId: SnowflakeId, userIds: SnowflakeId[]) =>
    request.post<string>(`/api/v1/role/${roleId}/users`, { userIds }),
  removeUsers: (roleId: SnowflakeId, userIds: SnowflakeId[]) =>
    request.delete<string>(`/api/v1/role/${roleId}/users`, { data: { userIds } }),
  getPermissions: (roleKey: string) =>
    request.get<string[]>('/api/v1/role/permissions', { params: { roleKey } }),
  getMenus: (roleId: SnowflakeId) => request.get<SnowflakeId[]>(`/api/v1/role/${roleId}/menus`),
  assignMenus: (roleId: SnowflakeId, menuIds: SnowflakeId[]) =>
    request.post<string>(`/api/v1/role/${roleId}/menus`, { menuIds }),
};
