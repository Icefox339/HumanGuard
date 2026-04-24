import { useAuthStore } from '@/app/store/auth-store';

export const Header = () => {
  const clearSession = useAuthStore((s) => s.clearSession);
  return (
    <header className="theme-surface flex items-center justify-end border-b theme-border px-6 py-3">
      <button className="interactive-chip theme-button" onClick={clearSession}>
        Logout
      </button>
    </header>
  );
};
