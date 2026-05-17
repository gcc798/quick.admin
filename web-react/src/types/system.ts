import type { SnowflakeId } from './api';
import type { MenuRecord } from './menu';

export interface UserRecord {
  id: SnowflakeId;
  orgId?: SnowflakeId;
  userName: string;
  nickName: string;
  email?: string;
  phonenumber?: string;
  sex?: number;
  avatar?: string;
  userType?: number;
  status?: number;
  remark?: string;
  createdTime?: string;
  updatedTime?: string;
}

export interface UserFormData {
  orgId?: SnowflakeId;
  userName: string;
  nickName: string;
  password?: string;
  userType?: number;
  email?: string;
  phonenumber?: string;
  sex?: number;
  avatar?: string;
  status?: number;
  remark?: string;
}

export interface RoleRecord {
  id: SnowflakeId;
  roleName: string;
  roleKey: string;
  sort?: number;
  status?: number;
  dataScope?: number;
  remark?: string;
  createdTime?: string;
  updatedTime?: string;
}

export interface ApiPermissionRecord {
  id: SnowflakeId;
  parentId: SnowflakeId;
  module: string;
  code: string;
  name: string;
  nodeType: number;
  action: string;
  method?: string;
  path?: string;
  sort?: number;
  status?: number;
  remark?: string;
  createdTime?: string;
  updatedTime?: string;
  children?: ApiPermissionRecord[];
}

export interface OrgRecord {
  id: SnowflakeId;
  orgName: string;
  orgCode?: string;
  parentId?: SnowflakeId;
  orgType?: string;
  sort?: number;
  leader?: string;
  phone?: string;
  email?: string;
  status?: number;
  remark?: string;
  createdTime?: string;
  children?: OrgRecord[];
}

export interface DictRecord {
  id: SnowflakeId;
  parentId?: SnowflakeId;
  dictType: string;
  dictLabel: string;
  dictValue: string;
  sort?: number;
  isDefault?: boolean;
  status?: number;
  remark?: string;
  createBy?: SnowflakeId;
  updateBy?: SnowflakeId;
  createdTime?: string;
}

export interface ConfigRecord {
  id: SnowflakeId;
  name: string;
  code: string;
  data?: unknown;
  remark?: string;
  createBy?: SnowflakeId;
  updateBy?: SnowflakeId;
  createdTime?: string;
  updatedTime?: string;
}

export interface StorageEnvRecord {
  id: SnowflakeId;
  name: string;
  code: string;
  storageType: string;
  isDefault?: boolean;
  status?: number;
  config?: unknown;
  remark?: string;
  createdTime?: string;
  updatedTime?: string;
}

export interface AttachmentRecord {
  attachmentId: SnowflakeId;
  fileName: string;
  fileSize: number;
  fileType: string;
  filePath?: string;
  businessType?: string;
  businessId?: string;
  uploadBy?: SnowflakeId;
  uploadTime?: string;
}

export interface LoginLogRecord {
  id: SnowflakeId;
  userName: string;
  ipaddr: string;
  loginLocation?: string;
  browser?: string;
  os?: string;
  status?: number;
  msg?: string;
  loginTime?: string;
  clientId?: string;
}

export interface OperLogRecord {
  id: SnowflakeId;
  title: string;
  businessType?: string;
  method?: string;
  requestMethod?: string;
  deviceType?: string;
  operName?: string;
  operUrl?: string;
  operIp?: string;
  operLocation?: string;
  operParam?: string;
  jsonResult?: string;
  status?: string | number;
  errorMsg?: string;
  operTime?: string;
  costTime?: number;
  userAgent?: string;
}

export interface DashboardMetric {
  title: string;
  value: string;
  description: string;
}

export type PermissionCode = string;
export type MenuTree = MenuRecord[];
