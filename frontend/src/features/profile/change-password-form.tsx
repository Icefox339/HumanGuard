import { useState } from 'react';
import { AxiosError } from 'axios';
import { useForm } from 'react-hook-form';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import { changeUserPassword } from '@/api/users';
import { useAuthStore } from '@/app/store/auth-store';

const parseError = (error: unknown) => {
  const err = error as AxiosError<{ error?: string }>;
  return err.response?.data?.error ?? err.message ?? 'Unknown error';
};

const schema = z.object({
  oldPassword: z.string().min(1, 'Введите текущий пароль'),
  newPassword: z.string().min(8, 'Минимум 8 символов')
});

type FormValues = z.infer<typeof schema>;

export const ChangePasswordForm = () => {
  const user = useAuthStore((s) => s.user);
  const [status, setStatus] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors }
  } = useForm<FormValues>({ resolver: zodResolver(schema) });

  const onSubmit = async (values: FormValues) => {
    setStatus(null);
    setError(null);

    if (!user?.id) {
      setError('Не найден user id в сессии. Перелогинься.');
      return;
    }

    try {
      setSubmitting(true);
      await changeUserPassword(user.id, { old_password: values.oldPassword, new_password: values.newPassword });
      setStatus('Пароль успешно изменён.');
      reset();
    } catch (e) {
      setError(parseError(e));
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <form className="theme-card space-y-3 rounded-2xl border border-[rgb(var(--border))] p-5 shadow-sm" onSubmit={handleSubmit((values) => void onSubmit(values))}>
      <h2 className="text-lg font-semibold text-[rgb(var(--text-primary))]">Сменить пароль</h2>

      <div className="space-y-1.5">
        <input className="form-input w-full rounded-lg px-3 py-2" type="password" placeholder="Старый пароль" {...register('oldPassword')} />
        {errors.oldPassword && <p className="field-error">{errors.oldPassword.message}</p>}
      </div>

      <div className="space-y-1.5">
        <input className="form-input w-full rounded-lg px-3 py-2" type="password" placeholder="Новый пароль" {...register('newPassword')} />
        <p className="text-xs text-[rgb(var(--text-secondary))]">Новый пароль должен быть минимум 8 символов (как при регистрации/входе).</p>
        {errors.newPassword && <p className="field-error">{errors.newPassword.message}</p>}
      </div>

      {status && <p className="text-sm text-emerald-700">{status}</p>}
      {error && <p className="field-error">{error}</p>}
      <button className="interactive-chip theme-button px-4 py-2 disabled:opacity-60" disabled={submitting}>
        {submitting ? 'Обновляем...' : 'Обновить пароль'}
      </button>
    </form>
  );
};
