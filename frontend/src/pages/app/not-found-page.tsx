import { Link } from 'react-router-dom';
import { useAuthStore } from '@/app/store/auth-store';

export const NotFoundPage = () => {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);

  return (
    <section className="error-scene mx-auto flex min-h-[70vh] max-w-2xl flex-col items-center justify-center px-4 text-center">
      <div className="error-orb" aria-hidden="true" />
      <p className="error-code">404</p>
      <h1 className="text-3xl font-semibold text-[rgb(var(--text-primary))]">Кажется, эта страница потерялась</h1>
      <p className="mt-3 max-w-lg text-sm leading-6 text-[rgb(var(--text-secondary))]">
        {isAuthenticated
          ? 'Ничего страшного — вернитесь в панель управления.'
          : 'Эта страница доступна только после авторизации.'}
      </p>

      <div className="mt-8 flex flex-wrap items-center justify-center gap-3">
        {isAuthenticated ? (
          <>
            <Link to="/dashboard" className="interactive-chip theme-button px-5 py-2.5 text-sm font-medium">
              Открыть дашборд
            </Link>
            <Link to="/sites" className="interactive-chip rounded-lg border border-[rgb(var(--border))] bg-[rgb(var(--bg-surface))] px-5 py-2.5 text-sm font-medium text-[rgb(var(--text-primary))]">
              К сайтам
            </Link>
          </>
        ) : (
          <Link to="/auth/login" className="interactive-chip theme-button px-5 py-2.5 text-sm font-medium">
            Залогиниться
          </Link>
        )}
      </div>
    </section>
  );
};
