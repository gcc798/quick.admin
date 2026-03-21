import type { MenuRecord } from './menu';

export interface UserRecord {
  userId: number;
  orgId?: number;
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
  orgId?: number;
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
  roleId: number;
  roleName: string;
  roleKey: string;
  sort?: number;
  status?: number;
  dataScope?: number;
  remark?: string;
  createdTime?: string;
  updatedTime?: string;
}

export interface OrgRecord {
  id: number;
  orgName: string;
  orgCode?: string;
  parentId?: number;
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
  id: number;
  parentId?: number;
  dictType: string;
  dictLabel: string;
  dictValue: string;
  sort?: number;
  isDefault?: boolean;
  status?: number;
  remark?: string;
  createBy?: number;
  updateBy?: number;
  createdTime?: string;
}

export interface ConfigRecord {
  id: number;
  name: string;
  code: string;
  data?: unknown;
  remark?: string;
  createBy?: number;
  updateBy?: number;
  createdTime?: string;
  updatedTime?: string;
}

export interface StorageEnvRecord {
  id: number;
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
  attachmentId: number;
  fileName: string;
  fileSize: number;
  fileType: string;
  filePath?: string;
  businessType?: string;
  businessId?: string;
  uploadBy?: number;
  uploadTime?: string;
}

export interface LoginLogRecord {
  id: number;
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
  id: number;
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
