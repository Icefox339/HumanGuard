import { useEffect, useState } from 'react';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import { AxiosError } from 'axios';
import { useAuth } from '@/features/auth/use-auth';

const schema = z.object({
  email: z.string().email('Введите корректный email'),
  password: z.string().min(8, 'Минимум 8 символов'),
  totp_code: z.string().length(6, 'Введите 6 цифр из приложения-аутентификатора')
});

type FormValues = z.infer<typeof schema>;

export const LoginForm = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const [apiError, setApiError] = useState<string | null>(null);
  const { loginMutation } = useAuth();
  const { register, handleSubmit, formState: { errors } } = useForm<FormValues>({ resolver: zodResolver(schema) });

  useEffect(() => {
    if (loginMutation.isSuccess) {
      navigate('/dashboard', { replace: true });
    }
  }, [loginMutation.isSuccess, navigate]);

  const registrationHint = location.state && typeof location.state === 'object' && 'message' in location.state
    ? String(location.state.message)
    : null;

  return (
    <form
      className="w-full space-y-4 rounded border border-slate-200 bg-white p-6"
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
      <h1 className="text-2xl font-semibold">Вход</h1>
      {registrationHint && <p className="rounded bg-emerald-50 p-2 text-sm text-emerald-700">{registrationHint}</p>}
      <input placeholder="Email" className="w-full rounded border px-3 py-2" {...register('email')} />
      {errors.email && <p className="text-sm text-red-600">{errors.email.message}</p>}

      <input type="password" placeholder="Пароль" className="w-full rounded border px-3 py-2" {...register('password')} />
      {errors.password && <p className="text-sm text-red-600">{errors.password.message}</p>}

      <input placeholder="TOTP код (6 цифр)" className="w-full rounded border px-3 py-2" {...register('totp_code')} />
      {errors.totp_code && <p className="text-sm text-red-600">{errors.totp_code.message}</p>}

      {apiError && <p className="text-sm text-red-600">{apiError}</p>}

      <button disabled={loginMutation.isPending} className="w-full rounded bg-slate-900 py-2 text-white disabled:opacity-60">
        {loginMutation.isPending ? 'Входим...' : 'Войти'}
      </button>

      <p className="text-sm text-slate-600">
        Нет аккаунта?{' '}
        <Link className="text-slate-900 underline" to="/auth/register">
          Зарегистрироваться
        </Link>
      </p>
    </form>
  );
};
