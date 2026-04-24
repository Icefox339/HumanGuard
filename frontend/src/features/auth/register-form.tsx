import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import { AxiosError } from 'axios';
import { useAuth } from '@/features/auth/use-auth';
import { ErrorAlert } from '@/components/common/error-alert';

const schema = z.object({
  name: z.string().optional(),
  email: z.string().email('Введите корректный email'),
  password: z
    .string()
    .min(8, 'Минимум 8 символов')
    .regex(/[A-ZА-Я]/, 'Добавьте хотя бы одну заглавную букву')
    .regex(/[a-zа-я]/, 'Добавьте хотя бы одну строчную букву')
    .regex(/\d/, 'Добавьте хотя бы одну цифру')
});

type FormValues = z.infer<typeof schema>;

export const RegisterForm = () => {
  const navigate = useNavigate();
  const [apiError, setApiError] = useState<string | null>(null);
  const { registerMutation } = useAuth();
  const { register, handleSubmit, formState: { errors } } = useForm<FormValues>({ resolver: zodResolver(schema) });

  return (
    <form
      className="auth-card w-full space-y-4 rounded-2xl p-6"
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
      <h1 className="text-2xl font-semibold text-[rgb(var(--text-primary))]">Регистрация</h1>

      <div className="space-y-1.5">
        <input placeholder="Имя (опционально)" className="auth-input w-full rounded-lg px-3 py-2" {...register('name')} />
      </div>

      <div className="space-y-1.5">
        <input placeholder="Email" className="auth-input w-full rounded-lg px-3 py-2" {...register('email')} />
        <p className="auth-hint">Используйте действующий email, например name@example.com.</p>
        {errors.email && <p className="field-error">{errors.email.message}</p>}
      </div>

      <div className="space-y-1.5">
        <input type="password" placeholder="Пароль" className="auth-input w-full rounded-lg px-3 py-2" {...register('password')} />
        <div className="theme-info rounded-lg p-3">
          <p className="text-sm font-semibold">Требования к паролю:</p>
          <ul className="theme-text-muted mt-1 list-disc space-y-1 pl-5 text-xs">
            <li>Минимум 8 символов</li>
            <li>Хотя бы одна заглавная буква</li>
            <li>Хотя бы одна строчная буква</li>
            <li>Хотя бы одна цифра</li>
          </ul>
        </div>
        {errors.password && <p className="field-error">{errors.password.message}</p>}
      </div>

      {apiError && <ErrorAlert title="Ошибка регистрации" message={apiError} />}

      <button
        type="submit"
        disabled={registerMutation.isPending}
        className="interactive-chip theme-button w-full py-2 disabled:opacity-60"
      >
        {registerMutation.isPending ? 'Создаём аккаунт...' : 'Создать аккаунт'}
      </button>

      <p className="theme-text-muted text-sm">
        Уже есть аккаунт?{' '}
        <Link className="theme-link font-semibold underline" to="/auth/login">
          Войти
        </Link>
      </p>
    </form>
  );
};
