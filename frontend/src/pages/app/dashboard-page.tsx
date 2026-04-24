import { useState } from 'react';
import { applyTheme, getStoredTheme, persistTheme, type ThemeMode } from '@/lib/theme';

export const DashboardPage = () => {
  const [theme, setTheme] = useState<ThemeMode>(() => getStoredTheme());

  const toggleTheme = () => {
    const nextTheme: ThemeMode = theme === 'default' ? 'cyan' : 'default';
    setTheme(nextTheme);
    applyTheme(nextTheme);
    persistTheme(nextTheme);
  };

  return (
    <section className="space-y-5">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <h1 className="text-2xl font-semibold text-[rgb(var(--text-primary))]">Dashboard</h1>
        <button type="button" className="interactive-chip theme-button" onClick={toggleTheme}>
          {theme === 'cyan' ? 'Вернуть стандартную тему' : 'Включить cyan-тему'}
        </button>
      </div>

      <div className="theme-card max-w-xl rounded-2xl p-5">
        <p className="text-sm text-[rgb(var(--text-secondary))]">
          Переключатель темы влияет на весь интерфейс: фон, карточки, меню, кнопки и акцентные элементы.
        </p>
      </div>
    </section>
  );
};
