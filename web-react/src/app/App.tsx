import { useEffect, useMemo } from 'react';
import { BrowserRouter } from 'react-router-dom';
import { App as AntApp, ConfigProvider } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import { AppRoutes } from '@/router/AppRoutes';
import { useResolvedThemeMode, useThemeStore } from '@/store/theme';
import { getAntdTheme } from '@/theme/themeConfig';

export function App() {
  const themeMode = useThemeStore((state) => state.mode);
  const resolvedThemeMode = useResolvedThemeMode(themeMode);
  const themeConfig = useMemo(() => getAntdTheme(resolvedThemeMode), [resolvedThemeMode]);

  useEffect(() => {
    document.documentElement.dataset.theme = resolvedThemeMode;
    document.documentElement.dataset.themePreference = themeMode;
  }, [resolvedThemeMode, themeMode]);

  return (
    <ConfigProvider locale={zhCN} theme={themeConfig}>
      <AntApp>
        <div className="app-root">
          <BrowserRouter>
            <AppRoutes />
          </BrowserRouter>
        </div>
      </AntApp>
    </ConfigProvider>
  );
}
