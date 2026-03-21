export interface MenuRecord {
  id: number;
  menuName: string;
  parentId: number;
  sort?: number;
  path: string;
  component?: string;
  query?: string;
  isFrame?: number;
  isCache?: number;
  menuType: number;
  visible?: number;
  status?: number;
  perms?: string;
  icon?: string;
  remark?: string;
  children?: MenuRecord[];
}

export interface MenuRouteRecord extends MenuRecord {
  fullPath: string;
}
