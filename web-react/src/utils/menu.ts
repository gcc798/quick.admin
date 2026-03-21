import type { ItemType } from 'antd/es/menu/interface';
import type { MenuRecord, MenuRouteRecord } from '@/types/menu';
import { getMenuIconNode } from './icons';

export function joinMenuPath(parentPath: string, currentPath: string) {
  if (!currentPath) {
    return parentPath || '/';
  }

  if (currentPath.startsWith('/')) {
    return currentPath;
  }

  const normalizedParent = parentPath === '/' ? '' : parentPath.replace(/\/$/, '');
  return `${normalizedParent}/${currentPath}`.replace(/\/+/g, '/');
}

export function extractPermissions(menuTree: MenuRecord[]) {
  const permissions: string[] = [];

  const walk = (nodes: MenuRecord[]) => {
    nodes.forEach((node) => {
      if (node.perms) {
        permissions.push(node.perms);
      }

      if (node.children?.length) {
        walk(node.children);
      }
    });
  };

  walk(menuTree);
  return permissions;
}

export function flattenLeafMenus(
  menuTree: MenuRecord[],
  parentPath = '',
): MenuRouteRecord[] {
  const routes: MenuRouteRecord[] = [];

  menuTree.forEach((menu) => {
    if (menu.visible === 1 || menu.status === 1) {
      return;
    }

    const fullPath = joinMenuPath(parentPath, menu.path);

    if (menu.menuType === 1) {
      routes.push({ ...menu, fullPath });
    }

    if (menu.children?.length) {
      routes.push(...flattenLeafMenus(menu.children, fullPath));
    }
  });

  return routes;
}

export function buildSidebarMenus(
  menuTree: MenuRecord[],
  parentPath = '',
): ItemType[] {
  return menuTree
    .filter((menu) => menu.visible !== 1 && menu.status !== 1 && menu.menuType !== 2)
    .map((menu) => {
      const fullPath = joinMenuPath(parentPath, menu.path);
      const icon = getMenuIconNode(menu.icon);

      if (menu.children?.length) {
        return {
          key: fullPath,
          icon,
          label: menu.menuName,
          children: buildSidebarMenus(menu.children, fullPath),
        } satisfies ItemType;
      }

      return {
        key: fullPath,
        icon,
        label: menu.menuName,
      } satisfies ItemType;
    });
}
