import { useEffect, useMemo, useState, type PointerEvent as ReactPointerEvent } from 'react';
import { Outlet, useLocation, useNavigate } from 'react-router-dom';
import {
  AppstoreOutlined,
  BulbFilled,
  BulbOutlined,
  LogoutOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  UserOutlined,
} from '@ant-design/icons';
import type { ItemType } from 'antd/es/menu/interface';
import { Avatar, Button, Dropdown, Layout, Menu, Space } from 'antd';
import type { MenuRecord } from '@/types/menu';
import { usePermissionStore } from '@/store/permission';
import { useAppStore } from '@/store/app';
import { useAuthStore } from '@/store/auth';
import { useThemeStore } from '@/store/theme';
import { getMenuIconNode } from '@/utils/icons';
import { findFirstNavigablePath, isMenuHidden, joinMenuPath } from '@/utils/menu';

const { Header, Sider, Content } = Layout;

function findOpenKeys(pathname: string) {
  const parts = pathname.split('/').filter(Boolean);
  if (parts.length <= 1) {
    return [];
  }

  return parts.slice(0, -1).map((_, index) => joinMenuPath('', parts.slice(0, index + 1).join('/')));
}

function buildMenuItems(
  menuTree: MenuRecord[],
  navigate: (path: string) => void,
  parentPath = '',
): ItemType[] {
  return menuTree
    .filter((menu) => !isMenuHidden(menu) && menu.menuType !== 2)
    .map((menu) => {
      const fullPath = joinMenuPath(parentPath, menu.path);
      const icon = getMenuIconNode(menu.icon);
      const targetPath =
        menu.menuType === 1
          ? fullPath
          : findFirstNavigablePath(menu.children ?? [], fullPath) ?? fullPath;
      const childItems = buildMenuItems(menu.children ?? [], navigate, fullPath);

      // 这里不能直接看原始 children.length。
      // 因为很多“页面菜单”下面挂的是按钮权限节点（menuType=2），它们不应该把页面菜单渲染成可展开子菜单。
      if (childItems.length > 0) {
        return {
          key: fullPath,
          icon,
          label: menu.menuName,
          // 目录标题点击时直接进入第一个真实页面，避免只展开不跳转造成“没反应”的错觉。
          onTitleClick: () => {
            if (targetPath && targetPath !== fullPath) {
              navigate(targetPath);
            }
          },
          children: childItems,
        } satisfies ItemType;
      }

      return {
        key: fullPath,
        icon,
        label: menu.menuName,
        onClick: () => navigate(targetPath),
      } satisfies ItemType;
    });
}

export function AppLayout() {
  const navigate = useNavigate();
  const location = useLocation();
  const menuTree = usePermissionStore((state) => state.menuTree);
  const collapsed = useAppStore((state) => state.collapsed);
  const toggleCollapsed = useAppStore((state) => state.toggleCollapsed);
  const authStore = useAuthStore();
  const mode = useThemeStore((state) => state.mode);
  const toggleTheme = useThemeStore((state) => state.toggleTheme);

  const menuItems = useMemo(
    () => buildMenuItems(menuTree, (path) => navigate(path)),
    [menuTree, navigate],
  );
  const currentOpenKeys = useMemo(() => findOpenKeys(location.pathname), [location.pathname]);
  const [openKeys, setOpenKeys] = useState<string[]>(currentOpenKeys);
  const [isPointerPressing, setIsPointerPressing] = useState(false);

  useEffect(() => {
    setOpenKeys(currentOpenKeys);
  }, [currentOpenKeys]);

  useEffect(() => {
    if (!isPointerPressing) {
      return undefined;
    }

    const resetPointerState = () => setIsPointerPressing(false);

    window.addEventListener('pointerup', resetPointerState);
    window.addEventListener('blur', resetPointerState);

    return () => {
      window.removeEventListener('pointerup', resetPointerState);
      window.removeEventListener('blur', resetPointerState);
    };
  }, [isPointerPressing]);

  const handlePointerMove = (event: ReactPointerEvent<HTMLDivElement>) => {
    const rect = event.currentTarget.getBoundingClientRect();
    const x = ((event.clientX - rect.left) / rect.width) * 100;
    const y = ((event.clientY - rect.top) / rect.height) * 100;

    // 用 CSS 变量驱动鼠标追踪光效，避免每次移动都触发 React 重渲染。
    event.currentTarget.style.setProperty('--pointer-x', `${x}%`);
    event.currentTarget.style.setProperty('--pointer-y', `${y}%`);
  };

  return (
    <div
      className={`app-shell-wrap${isPointerPressing ? ' is-pressing' : ''}`}
      onPointerDown={() => setIsPointerPressing(true)}
      onPointerMove={handlePointerMove}
    >
      <Layout className="app-shell">
        <Sider
          breakpoint="lg"
          className="app-sider"
          collapsed={collapsed}
          collapsible
          onCollapse={toggleCollapsed}
          theme="light"
          trigger={null}
          width={224}
        >
          <div className="app-logo">
            <span className="app-logo-mark">
              <AppstoreOutlined />
            </span>
            {!collapsed ? (
              <span className="app-logo-text">{import.meta.env.VITE_APP_TITLE}</span>
            ) : null}
          </div>

          <Menu
            className="app-menu"
            inlineIndent={14}
            mode="inline"
            openKeys={openKeys}
            selectedKeys={[location.pathname]}
            items={menuItems}
            onOpenChange={(keys) => setOpenKeys(keys as string[])}
            theme="light"
          />
        </Sider>

        <Layout>
          <Header className="app-header">
            <Space align="center" className="app-header-group" size={12}>
              <Button
                className="header-action-btn"
                icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
                type="text"
                onClick={toggleCollapsed}
              />
            </Space>

            <Space align="center" className="app-header-group" size={12}>
              <Button
                className="header-action-btn header-theme-btn"
                icon={mode === 'dark' ? <BulbFilled /> : <BulbOutlined />}
                type="text"
                onClick={toggleTheme}
              />
              <Dropdown
                menu={{
                  items: [
                    {
                      key: 'logout',
                      icon: <LogoutOutlined />,
                      label: '退出登录',
                      onClick: async () => {
                        await authStore.logout();
                        navigate('/login', { replace: true });
                      },
                    },
                  ],
                }}
              >
                <Space className="header-user">
                  <Avatar icon={<UserOutlined />} shape="square" size={30} />
                  <span className="header-user-name">
                    {authStore.userInfo?.nickname || authStore.userInfo?.username || '管理员'}
                  </span>
                </Space>
              </Dropdown>
            </Space>
          </Header>

          <Content className="app-content">
            <div className="app-view" key={location.pathname}>
              <Outlet />
            </div>
          </Content>
        </Layout>
      </Layout>
    </div>
  );
}
