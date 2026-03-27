import { useEffect, useMemo } from 'react';
import { BrowserRouter } from 'react-router-dom';
import { App as AntApp, ConfigProvider } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import { AppRoutes } from '@/router/AppRoutes';
import { useThemeStore } from '@/store/theme';
import { getAntdTheme } from '@/theme/themeConfig';

export function App() {
  const mode = useThemeStore((state) => state.mode);
  const themeConfig = useMemo(() => getAntdTheme(mode), [mode]);

  useEffect(() => {
    document.documentElement.dataset.theme = mode;
  }, [mode]);

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
