import { FormEvent, useState } from 'react';
import { AxiosError } from 'axios';
import { api } from '@/api/client';
import { useAuthStore } from '@/app/store/auth-store';

const parseError = (error: unknown) => {
  const err = error as AxiosError<{ error?: string }>;
  return err.response?.data?.error ?? err.message ?? 'Unknown error';
};

export const UpdateAvatarForm = () => {
  const user = useAuthStore((s) => s.user);
  const [avatarUrl, setAvatarUrl] = useState('');
  const [status, setStatus] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const onSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setStatus(null);
    setError(null);

    if (!user?.id) {
      setError('Не найден user id в сессии. Перелогинься.');
      return;
    }

    try {
      await api.post(`/users/${user.id}/avatar`, { avatar_url: avatarUrl });
      setStatus('Аватар обновлён.');
      setAvatarUrl('');
    } catch (e) {
      setError(parseError(e));
    }
  };

  return (
    <form className="space-y-3 rounded border border-slate-200 bg-white p-4" onSubmit={onSubmit}>
      <h2 className="text-lg font-semibold">Обновить аватар</h2>
      <input className="w-full rounded border px-3 py-2" placeholder="https://..." value={avatarUrl} onChange={(e) => setAvatarUrl(e.target.value)} required />
      {status && <p className="text-sm text-emerald-700">{status}</p>}
      {error && <p className="text-sm text-red-600">{error}</p>}
      <button className="rounded bg-slate-900 px-4 py-2 text-white">Сохранить</button>
    </form>
  );
};
