import { useAuthStore } from '@/app/store/auth-store';

export const Header = () => {
  const clearSession = useAuthStore((s) => s.clearSession);
  return (
    <header className="flex items-center justify-end border-b border-slate-200 bg-white px-6 py-3">
      <button className="rounded bg-slate-900 px-4 py-2 text-sm text-white" onClick={clearSession}>
        Logout
      </button>
    </header>
  );
};
