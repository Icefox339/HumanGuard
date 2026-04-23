import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import { AxiosError } from 'axios';
import { useAuth } from '@/features/auth/use-auth';

const schema = z.object({
  name: z.string().optional(),
  email: z.string().email('Введите корректный email'),
  password: z.string().min(8, 'Минимум 8 символов')
});

type FormValues = z.infer<typeof schema>;

export const RegisterForm = () => {
  const navigate = useNavigate();
  const [apiError, setApiError] = useState<string | null>(null);
  const { registerMutation } = useAuth();
  const { register, handleSubmit, formState: { errors } } = useForm<FormValues>({ resolver: zodResolver(schema) });

  return (
    <form
      className="w-full space-y-4 rounded border border-slate-200 bg-white p-6"
      onSubmit={handleSubmit((v) => {
        setApiError(null);
        registerMutation.mutate(v, {
          onSuccess: (result) => {
            navigate('/auth/2fa-setup', {
              state: {
                email: v.email,
                totp_secret: result.totp_secret,
                qr_code_url: result.qr_code_url
              }
            });
          },
          onError: (error) => {
            const err = error as AxiosError<{ error?: string }>;
            setApiError(err.response?.data?.error ?? 'Не удалось зарегистрироваться. Попробуйте снова.');
          }
        });
      })}
    >
      <h1 className="text-2xl font-semibold">Регистрация</h1>

      <input placeholder="Имя (опционально)" className="w-full rounded border px-3 py-2" {...register('name')} />

      <input placeholder="Email" className="w-full rounded border px-3 py-2" {...register('email')} />
      {errors.email && <p className="text-sm text-red-600">{errors.email.message}</p>}

      <input type="password" placeholder="Пароль" className="w-full rounded border px-3 py-2" {...register('password')} />
      {errors.password && <p className="text-sm text-red-600">{errors.password.message}</p>}

      {apiError && <p className="text-sm text-red-600">{apiError}</p>}

      <button
        type="submit"
        disabled={registerMutation.isPending}
        className="w-full rounded bg-slate-900 py-2 text-white disabled:opacity-60"
      >
        {registerMutation.isPending ? 'Создаём аккаунт...' : 'Создать аккаунт'}
      </button>

      <p className="text-sm text-slate-600">
        Уже есть аккаунт?{' '}
        <Link className="text-slate-900 underline" to="/auth/login">
          Войти
        </Link>
      </p>
    </form>
  );
};
