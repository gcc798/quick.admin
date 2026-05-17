import { useEffect, useState } from 'react';
import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';

export type ThemeMode = 'light' | 'dark' | 'system';
export type ResolvedThemeMode = 'light' | 'dark';

function readSystemThemeMode(): ResolvedThemeMode {
  if (typeof window !== 'undefined' && window.matchMedia('(prefers-color-scheme: dark)').matches) {
    return 'dark';
  }

  return 'light';
}

export function resolveThemeMode(
  mode: ThemeMode,
  systemMode: ResolvedThemeMode = readSystemThemeMode(),
): ResolvedThemeMode {
  return mode === 'system' ? systemMode : mode;
}

export function useResolvedThemeMode(mode: ThemeMode): ResolvedThemeMode {
  const [systemMode, setSystemMode] = useState<ResolvedThemeMode>(() => readSystemThemeMode());

  useEffect(() => {
    if (typeof window === 'undefined' || typeof window.matchMedia !== 'function') {
      return undefined;
    }

    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
    const syncSystemMode = () => {
      setSystemMode(mediaQuery.matches ? 'dark' : 'light');
    };

    syncSystemMode();

    if (typeof mediaQuery.addEventListener === 'function') {
      mediaQuery.addEventListener('change', syncSystemMode);
      return () => mediaQuery.removeEventListener('change', syncSystemMode);
    }

    mediaQuery.addListener(syncSystemMode);
    return () => mediaQuery.removeListener(syncSystemMode);
  }, []);

  return resolveThemeMode(mode, systemMode);
}

interface ThemeState {
  mode: ThemeMode;
  toggleTheme: () => void;
  setTheme: (mode: ThemeMode) => void;
}

export const useThemeStore = create<ThemeState>()(
  persist(
    (set) => ({
      mode: 'light',
      toggleTheme: () =>
        set((state) => ({
          mode: resolveThemeMode(state.mode) === 'light' ? 'dark' : 'light',
        })),
      setTheme: (mode) => set({ mode }),
    }),
    {
      name: 'web-react-theme',
      storage: createJSONStorage(() => localStorage),
    },
  ),
);
