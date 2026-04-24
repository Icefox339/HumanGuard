export type ThemeMode = 'default' | 'cyan';

const THEME_STORAGE_KEY = 'hg-theme';

export const getStoredTheme = (): ThemeMode => {
  if (typeof window === 'undefined') {
    return 'default';
  }

  const saved = window.localStorage.getItem(THEME_STORAGE_KEY);
  return saved === 'cyan' ? 'cyan' : 'default';
};

export const applyTheme = (theme: ThemeMode) => {
  if (typeof document === 'undefined') {
    return;
  }

  document.documentElement.dataset.theme = theme;
};

export const persistTheme = (theme: ThemeMode) => {
  if (typeof window === 'undefined') {
    return;
  }

  window.localStorage.setItem(THEME_STORAGE_KEY, theme);
};

export const initializeTheme = () => {
  applyTheme(getStoredTheme());
};
