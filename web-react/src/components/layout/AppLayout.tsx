import { useEffect, useMemo, useState } from 'react';
import { Link, Outlet, useLocation, useNavigate } from 'react-router-dom';
import {
  AppstoreOutlined,
  BulbFilled,
  BulbOutlined,
  LogoutOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  UserOutlined,
} from '@ant-design/icons';
import { Avatar, Button, Dropdown, Layout, Menu, Space, Typography } from 'antd';
import { buildSidebarMenus, joinMenuPath } from '@/utils/menu';
import { usePermissionStore } from '@/store/permission';
import { useAppStore } from '@/store/app';
import { useAuthStore } from '@/store/auth';
import { useThemeStore } from '@/store/theme';

const { Header, Sider, Content } = Layout;

function findOpenKeys(pathname: string) {
  const parts = pathname.split('/').filter(Boolean);
  if (parts.length <= 1) {
    return [];
  }

  return parts.slice(0, -1).map((_, index) => joinMenuPath('', parts.slice(0, index + 1).join('/')));
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

  const menuItems = useMemo(() => buildSidebarMenus(menuTree), [menuTree]);
  const currentOpenKeys = useMemo(() => findOpenKeys(location.pathname), [location.pathname]);
  const [openKeys, setOpenKeys] = useState<string[]>(currentOpenKeys);

  useEffect(() => {
    setOpenKeys(currentOpenKeys);
  }, [currentOpenKeys]);

  return (
    <Layout className="app-shell">
      <Sider
        breakpoint="lg"
        collapsed={collapsed}
        collapsible
        onCollapse={toggleCollapsed}
        trigger={null}
        width={220}
      >
        <div className="app-logo">
          <AppstoreOutlined />
          {!collapsed ? <span>{import.meta.env.VITE_APP_TITLE}</span> : null}
        </div>

        <Menu
          mode="inline"
          openKeys={openKeys}
          selectedKeys={[location.pathname]}
          items={menuItems}
          onOpenChange={(keys) => setOpenKeys(keys as string[])}
          onClick={({ key }) => navigate(String(key))}
        />
      </Sider>

      <Layout>
        <Header className="app-header">
          <Space align="center">
            <Button
              icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
              type="text"
              onClick={toggleCollapsed}
            />
          </Space>

          <Space align="center" size="middle">
            <Button
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
                <Avatar icon={<UserOutlined />} />
                <Typography.Text>
                  {authStore.userInfo?.nickname || authStore.userInfo?.username || '管理员'}
                </Typography.Text>
              </Space>
            </Dropdown>
          </Space>
        </Header>

        <Content className="app-content">
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  );
}
