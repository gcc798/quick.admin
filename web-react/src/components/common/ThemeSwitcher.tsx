import { useEffect, useMemo, useRef, useState } from 'react';
import {
  CheckOutlined,
  DesktopOutlined,
  MoonOutlined,
  SunOutlined,
} from '@ant-design/icons';
import { Button } from 'antd';
import type { ThemeMode } from '@/store/theme';
import { useResolvedThemeMode, useThemeStore } from '@/store/theme';

const themeOptions: Array<{
  key: ThemeMode;
  label: string;
  description: string;
  icon: JSX.Element;
}> = [
  {
    key: 'light',
    label: '浅色模式',
    description: '明亮界面，适合日间浏览',
    icon: <SunOutlined />,
  },
  {
    key: 'dark',
    label: '深色模式',
    description: '降低眩光，适合夜间使用',
    icon: <MoonOutlined />,
  },
  {
    key: 'system',
    label: '跟随系统',
    description: '自动同步系统外观设置',
    icon: <DesktopOutlined />,
  },
];

interface ThemeSwitcherProps {
  entryClassName?: string;
  buttonClassName?: string;
  popoverClassName?: string;
}

export function ThemeSwitcher({
  entryClassName = 'header-theme-entry',
  buttonClassName = 'header-theme-switch',
  popoverClassName = 'theme-popover-floating',
}: ThemeSwitcherProps) {
  const mode = useThemeStore((state) => state.mode);
  const resolvedMode = useResolvedThemeMode(mode);
  const setTheme = useThemeStore((state) => state.setTheme);
  const [open, setOpen] = useState(false);
  const panelRef = useRef<HTMLDivElement | null>(null);
  const activeThemeOption = useMemo(
    () => themeOptions.find((item) => item.key === mode) ?? themeOptions[0],
    [mode],
  );

  useEffect(() => {
    if (!open) {
      return undefined;
    }

    const handlePointerDown = (event: PointerEvent) => {
      if (!panelRef.current?.contains(event.target as Node)) {
        setOpen(false);
      }
    };

    window.addEventListener('pointerdown', handlePointerDown);
    return () => window.removeEventListener('pointerdown', handlePointerDown);
  }, [open]);

  return (
    <div className={entryClassName} ref={panelRef}>
      <Button
        aria-label="主题切换"
        className={buttonClassName}
        icon={activeThemeOption.icon}
        title={activeThemeOption.label}
        type="text"
        onClick={() => setOpen((value) => !value)}
      />
      {open ? (
        <div className={`theme-popover ${popoverClassName}`} role="menu" aria-label="主题切换">
          <div className="theme-popover-header">
            <strong>界面外观</strong>
            <span>
              当前生效：
              {resolvedMode === 'dark' ? '深色模式' : '浅色模式'}
            </span>
          </div>
          <div className="theme-popover-list">
            {themeOptions.map((option) => {
              const selected = option.key === mode;

              return (
                <button
                  aria-pressed={selected}
                  className={`theme-popover-option${selected ? ' is-active' : ''}`}
                  key={option.key}
                  type="button"
                  onClick={() => {
                    setTheme(option.key);
                    setOpen(false);
                  }}
                >
                  <span className="theme-popover-option-icon">{option.icon}</span>
                  <span className="theme-popover-option-body">
                    <span className="theme-popover-option-label">{option.label}</span>
                    <span className="theme-popover-option-description">
                      {option.description}
                    </span>
                  </span>
                  <span className="theme-popover-option-check">
                    {selected ? <CheckOutlined /> : null}
                  </span>
                </button>
              );
            })}
          </div>
        </div>
      ) : null}
    </div>
  );
}
