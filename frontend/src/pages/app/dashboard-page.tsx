import { useState } from 'react';
import { applyTheme, getStoredTheme, persistTheme, type ThemeMode } from '@/lib/theme';
import { useAuthStore } from '@/app/store/auth-store';
import { BehaviorDebugPanel } from '@/features/sessions/behavior-debug-panel';

const themeLabels: Record<ThemeMode, string> = {
  default: 'Стандарт',
  cyan: 'Cyan',
  dark: 'Тёмная'
};

export const DashboardPage = () => {
  const [theme, setTheme] = useState<ThemeMode>(() => getStoredTheme());
  const user = useAuthStore((s) => s.user);

  const selectTheme = (nextTheme: ThemeMode) => {
    setTheme(nextTheme);
    applyTheme(nextTheme);
    persistTheme(nextTheme);
  };

  return (
    <section className="space-y-5">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <h1 className="text-2xl font-semibold text-[rgb(var(--text-primary))]">Dashboard</h1>
        <div className="flex flex-wrap gap-2">
          {(['default', 'cyan', 'dark'] as ThemeMode[]).map((themeOption) => (
            <button
              key={themeOption}
              type="button"
              className={theme === themeOption ? 'interactive-chip theme-button' : 'interactive-chip theme-button theme-button--muted'}
              onClick={() => selectTheme(themeOption)}
            >
              {themeLabels[themeOption]}
            </button>
          ))}
        </div>
      </div>

      <div className="theme-card max-w-xl rounded-2xl p-5">
        <p className="text-sm font-medium text-[rgb(var(--text-secondary))]">Текущая тема: {themeLabels[theme]}.</p>
      </div>

      {user?.role === 'admin' ? <BehaviorDebugPanel /> : null}
    </section>
  );
};
