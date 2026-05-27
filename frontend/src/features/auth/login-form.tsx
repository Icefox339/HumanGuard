import { useEffect, useState } from 'react';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import { AxiosError } from 'axios';
import { useAuth } from '@/features/auth/use-auth';
import { ErrorAlert } from '@/components/common/error-alert';
import { API_URL } from '@/lib/constants';
import { useAuthStore } from '@/app/store/auth-store';

const schema = z.object({
  email: z.string().email('Введите корректный email'),
  password: z.string().min(8, 'Минимум 8 символов'),
  totp_code: z.string().optional().refine((value) => !value || value.length === 6, 'Если код указан, он должен быть 6 цифр')
});

type FormValues = z.infer<typeof schema>;
type OAuthProvider = 'keycloak' | 'google' | 'github';

export const LoginForm = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  const [apiError, setApiError] = useState<string | null>(null);
  const { loginMutation } = useAuth();
  const {
    register,
    handleSubmit,
    formState: { errors }
  } = useForm<FormValues>({ resolver: zodResolver(schema) });

  useEffect(() => {
    if (isAuthenticated) {
      navigate('/dashboard', { replace: true });
    }
  }, [isAuthenticated, navigate]);



  const [oauthPendingProvider, setOauthPendingProvider] = useState<OAuthProvider | null>(null);
  const oauthProviders: Array<{ id: OAuthProvider; label: string }> = [
    { id: 'keycloak', label: 'Keycloak' },
    { id: 'google', label: 'Google' },
    { id: 'github', label: 'GitHub' }
  ];

  const startOAuthLogin = (provider: OAuthProvider) => {
    setOauthPendingProvider(provider);
    window.location.href = `${API_URL}/api/auth/${provider}/login`;
  };

  const registrationHint = location.state && typeof location.state === 'object' && 'message' in location.state
    ? String(location.state.message)
    : null;

  return (
    <form
      className="auth-card w-full space-y-4 rounded-2xl p-6"
      onSubmit={handleSubmit((v) => {
        setApiError(null);
        loginMutation.mutate(v, {
          onError: (error) => {
            const err = error as AxiosError<{ error?: string }>;
            setApiError(err.response?.data?.error ?? 'Не удалось войти. Проверьте данные и попробуйте снова.');
          }
        });
      })}
    >
      <h1 className="text-2xl font-semibold text-[rgb(var(--text-primary))]">Вход</h1>
      {registrationHint && <p className="rounded-lg bg-emerald-100/80 p-2 text-sm text-emerald-700">{registrationHint}</p>}
      <div className="space-y-1.5">
        <input placeholder="Email" className="auth-input w-full rounded-lg px-3 py-2" {...register('email')} />
        <p className="auth-hint">Введите корректный email (например, name@example.com).</p>
        {errors.email && <p className="field-error">{errors.email.message}</p>}
      </div>

      <div className="space-y-1.5">
        <input type="password" placeholder="Пароль" className="auth-input w-full rounded-lg px-3 py-2" {...register('password')} />
        <p className="auth-hint">Пароль должен быть минимум 8 символов.</p>
        {errors.password && <p className="field-error">{errors.password.message}</p>}
      </div>

      <div className="space-y-1.5">
        <input placeholder="TOTP код (если включён 2FA)" className="auth-input w-full rounded-lg px-3 py-2" {...register('totp_code')} />
        {errors.totp_code && <p className="field-error">{errors.totp_code.message}</p>}
      </div>

      {apiError && <ErrorAlert message={apiError} />}

      <button disabled={loginMutation.isPending} className="interactive-chip theme-button w-full py-2 disabled:opacity-60">
        {loginMutation.isPending ? 'Входим...' : 'Войти'}
      </button>


      <div className="space-y-2 pt-2">
        <p className="text-xs uppercase tracking-wide text-[rgb(var(--text-secondary))]">Или войдите через SSO OAuth</p>
        <div className="grid gap-2 sm:grid-cols-3">
          {oauthProviders.map((provider) => (
            <button
              key={provider.id}
              type="button"
              disabled={oauthPendingProvider !== null}
              onClick={() => startOAuthLogin(provider.id)}
              className="interactive-chip rounded-lg border border-[rgb(var(--border))] px-3 py-2 text-sm disabled:cursor-not-allowed disabled:opacity-60"
            >
              {oauthPendingProvider === provider.id ? 'Переходим...' : provider.label}
            </button>
          ))}
        </div>
      </div>

      <p className="text-sm text-[rgb(var(--text-secondary))]">
        Нет аккаунта?{' '}
        <Link className="font-medium text-[rgb(var(--accent))] underline" to="/auth/register">
          Зарегистрироваться
        </Link>
      </p>
    </form>
  );
};
