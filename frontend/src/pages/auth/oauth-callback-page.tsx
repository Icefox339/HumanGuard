import { useEffect, useState } from 'react';
import { Link, useNavigate, useSearchParams } from 'react-router-dom';
import { useAuthStore } from '@/app/store/auth-store';

type OAuthUser = {
  id: string;
  email: string;
  role: 'user' | 'admin';
};

export const OAuthCallbackPage = () => {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const setSession = useAuthStore((s) => s.setSession);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const providerError = searchParams.get('error');
    if (providerError) {
      setError(`OAuth провайдер вернул ошибку: ${providerError}`);
      return;
    }

    const token = searchParams.get('token');
    const userRaw = searchParams.get('user');

    if (!token || !userRaw) {
      setError('OAuth вход не удался: отсутствует токен или данные пользователя.');
      return;
    }

    try {
      const user = JSON.parse(userRaw) as OAuthUser;
      if (!user?.id || !user?.email || !user?.role) {
        setError('OAuth вход не удался: неверный формат данных пользователя.');
        return;
      }
      setSession(token, user);
      window.history.replaceState({}, document.title, '/auth/oauth/callback');
      navigate('/dashboard', { replace: true });
    } catch {
      setError('OAuth вход не удался: не удалось разобрать ответ провайдера.');
    }
  }, [navigate, searchParams, setSession]);

  return (
    <section className="auth-card w-full space-y-3 rounded-2xl p-6 text-sm text-[rgb(var(--text-secondary))]">
      <h1 className="text-xl font-semibold text-[rgb(var(--text-primary))]">Вход через OAuth</h1>
      {error ? (
        <div className="space-y-3">
          <p className="text-red-500">{error}</p>
          <Link className="font-medium text-[rgb(var(--accent))] underline" to="/auth/login">
            Вернуться на страницу входа
          </Link>
        </div>
      ) : (
        <p>Завершаем вход, подождите...</p>
      )}
    </section>
  );
};
