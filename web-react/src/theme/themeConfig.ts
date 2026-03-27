import type { ThemeConfig } from 'antd';
import { theme } from 'antd';
import type { ThemeMode } from '@/store/theme';

interface ThemePalette {
  primary: string;
  success: string;
  warning: string;
  error: string;
  text: string;
  textSecondary: string;
  textTertiary: string;
  bgBase: string;
  panel: string;
  panelStrong: string;
  border: string;
  borderSecondary: string;
  fill: string;
  fillSoft: string;
  hover: string;
  selected: string;
  shadow: string;
}

function getThemePalette(mode: ThemeMode): ThemePalette {
  if (mode === 'dark') {
    return {
      primary: '#58c3db',
      success: '#22c55e',
      warning: '#f59e0b',
      error: '#f87171',
      text: '#f6f7f9',
      textSecondary: '#cbd5df',
      textTertiary: '#94a3b8',
      bgBase: '#23272f',
      panel: 'rgba(39, 44, 53, 0.86)',
      panelStrong: 'rgba(42, 48, 58, 0.96)',
      border: 'rgba(148, 163, 184, 0.16)',
      borderSecondary: 'rgba(148, 163, 184, 0.08)',
      fill: 'rgba(88, 195, 219, 0.12)',
      fillSoft: 'rgba(88, 195, 219, 0.06)',
      hover: 'rgba(255, 255, 255, 0.05)',
      selected: 'rgba(88, 195, 219, 0.12)',
      shadow: '0 16px 40px rgba(15, 23, 42, 0.24)',
    };
  }

  return {
    primary: '#087ea4',
    success: '#16a34a',
    warning: '#f59e0b',
    error: '#ef4444',
    text: '#0f172a',
    textSecondary: '#475569',
    textTertiary: '#64748b',
    bgBase: '#f7f8fa',
    panel: 'rgba(255, 255, 255, 0.92)',
    panelStrong: 'rgba(255, 255, 255, 0.98)',
    border: 'rgba(15, 23, 42, 0.08)',
    borderSecondary: 'rgba(15, 23, 42, 0.05)',
    fill: 'rgba(8, 126, 164, 0.1)',
    fillSoft: 'rgba(8, 126, 164, 0.05)',
    hover: 'rgba(15, 23, 42, 0.05)',
    selected: 'rgba(8, 126, 164, 0.1)',
    shadow: '0 12px 28px rgba(15, 23, 42, 0.08)',
  };
}

export function getAntdTheme(mode: ThemeMode): ThemeConfig {
  const palette = getThemePalette(mode);

  // 主题配置拆到单独文件，方便后续统一调整日间/夜间色板，
  // 也避免把布局代码和视觉 token 混在一起。
  return {
    algorithm:
      mode === 'dark' ? theme.darkAlgorithm : theme.defaultAlgorithm,
    token: {
      colorPrimary: palette.primary,
      colorInfo: palette.primary,
      colorSuccess: palette.success,
      colorWarning: palette.warning,
      colorError: palette.error,
      colorText: palette.text,
      colorTextSecondary: palette.textSecondary,
      colorTextTertiary: palette.textTertiary,
      colorBgBase: palette.bgBase,
      colorBgLayout: palette.bgBase,
      colorBgContainer: palette.panel,
      colorBgElevated: palette.panelStrong,
      colorBorder: palette.border,
      colorBorderSecondary: palette.borderSecondary,
      colorSplit: palette.borderSecondary,
      colorFillSecondary: palette.fill,
      colorFillTertiary: palette.fillSoft,
      colorLink: palette.primary,
      colorLinkHover: palette.primary,
      boxShadowSecondary: palette.shadow,
      fontFamily:
        '"Avenir Next", "PingFang SC", "Microsoft YaHei", sans-serif',
      fontFamilyCode:
        '"SFMono-Regular", "JetBrains Mono", Consolas, monospace',
      borderRadius: 10,
      borderRadiusLG: 12,
      borderRadiusSM: 8,
      controlHeight: 38,
      controlHeightLG: 42,
    },
    components: {
      Menu: {
        itemBg: 'transparent',
        itemColor: palette.textSecondary,
        itemHoverColor: palette.text,
        itemHoverBg: palette.hover,
        itemSelectedColor: palette.text,
        itemSelectedBg: palette.selected,
        subMenuItemBg: 'transparent',
        activeBarHeight: 0,
      },
      Button: {
        primaryShadow: 'none',
        borderRadius: 10,
      },
    },
  };
}
