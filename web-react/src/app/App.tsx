import { BrowserRouter } from 'react-router-dom';
import { App as AntApp, ConfigProvider, theme } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import { AppRoutes } from '@/router/AppRoutes';
import { useThemeStore } from '@/store/theme';

export function App() {
  const mode = useThemeStore((state) => state.mode);

  return (
    <ConfigProvider
      locale={zhCN}
      theme={{
        algorithm:
          mode === 'dark' ? theme.darkAlgorithm : theme.defaultAlgorithm,
        token: {
          colorPrimary: '#1677ff',
        },
      }}
    >
      <AntApp>
        <BrowserRouter>
          <AppRoutes />
        </BrowserRouter>
      </AntApp>
    </ConfigProvider>
  );
}
