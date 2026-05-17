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
      primary: '#14b8a6',
      success: '#52c41a',
      warning: '#faad14',
      error: '#ff4d4f',
      text: '#f7fafc',
      textSecondary: '#cbd5e1',
      textTertiary: '#94a3b8',
      bgBase: '#111315',
      panel: 'rgba(25, 28, 31, 0.88)',
      panelStrong: 'rgba(30, 34, 38, 0.96)',
      border: 'rgba(203, 213, 225, 0.15)',
      borderSecondary: 'rgba(203, 213, 225, 0.08)',
      fill: 'rgba(20, 184, 166, 0.16)',
      fillSoft: 'rgba(20, 184, 166, 0.08)',
      hover: 'rgba(255, 255, 255, 0.06)',
      selected: 'rgba(20, 184, 166, 0.18)',
      shadow: '0 22px 54px rgba(0, 0, 0, 0.34)',
    };
  }

  return {
    primary: '#0f766e',
    success: '#52c41a',
    warning: '#faad14',
    error: '#ff4d4f',
    text: '#17202a',
    textSecondary: '#475569',
    textTertiary: '#7b8794',
    bgBase: '#f4f7fb',
    panel: 'rgba(255, 255, 255, 0.9)',
    panelStrong: 'rgba(255, 255, 255, 0.98)',
    border: 'rgba(71, 85, 105, 0.14)',
    borderSecondary: 'rgba(71, 85, 105, 0.09)',
    fill: 'rgba(15, 118, 110, 0.1)',
    fillSoft: 'rgba(15, 118, 110, 0.05)',
    hover: 'rgba(15, 118, 110, 0.08)',
    selected: 'rgba(15, 118, 110, 0.12)',
    shadow: '0 16px 34px rgba(15, 23, 42, 0.08)',
  };
}

export function getAntdTheme(mode: ThemeMode): ThemeConfig {
  const palette = getThemePalette(mode);

  // 主题配置拆到单独文件，方便后续扩展浅色、深色或更多品牌主题，
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
      fontFamily: '"Manrope", "Avenir Next", "PingFang SC", "Microsoft YaHei", sans-serif',
      fontFamilyCode:
        '"SFMono-Regular", "JetBrains Mono", Consolas, monospace',
      borderRadius: 8,
      borderRadiusLG: 8,
      borderRadiusSM: 6,
      controlHeight: 36,
      controlHeightLG: 42,
      wireframe: false,
    },
    components: {
      Card: {
        headerFontSize: 16,
        headerFontSizeSM: 14,
        paddingLG: 18,
      },
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
        borderRadius: 8,
        fontWeight: 700,
      },
      Table: {
        borderColor: palette.borderSecondary,
        headerBg: palette.fillSoft,
        headerColor: palette.textSecondary,
        rowHoverBg: palette.hover,
      },
      Modal: {
        borderRadiusLG: 8,
        titleFontSize: 18,
      },
      Input: {
        activeBorderColor: palette.primary,
        hoverBorderColor: palette.primary,
      },
      Select: {
        optionSelectedBg: palette.selected,
      },
      Tag: {
        borderRadiusSM: 6,
      },
    },
  };
}
