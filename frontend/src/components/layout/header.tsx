import { useAuthStore } from '@/app/store/auth-store';

type HeaderProps = {
  onToggleSidebar: () => void;
};

export const Header = ({ onToggleSidebar }: HeaderProps) => {
  const clearSession = useAuthStore((s) => s.clearSession);
  return (
    <header className="theme-surface sticky top-0 z-20 flex items-center justify-between gap-2 border-b theme-border px-3 py-3 sm:gap-3 sm:px-6">
      <button
        className="interactive-chip rounded-md border theme-border px-3 py-2 text-sm font-medium text-[rgb(var(--text-primary))] sm:hidden"
        onClick={onToggleSidebar}
        type="button"
      >
        Меню
      </button>
      <button className="interactive-chip theme-button px-3 py-2 text-sm sm:px-4" onClick={clearSession}>
        Выйти
      </button>
    </header>
  );
};
