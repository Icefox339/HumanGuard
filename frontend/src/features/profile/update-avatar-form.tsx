import { ChangeEvent, FormEvent, useState } from 'react';
import { AxiosError } from 'axios';
import { api } from '@/api/client';
import { useAuthStore } from '@/app/store/auth-store';

const parseError = (error: unknown) => {
  const err = error as AxiosError<{ error?: string }>;
  return err.response?.data?.error ?? err.message ?? 'Unknown error';
};

type UpdateAvatarFormProps = {
  onUpdated?: () => void;
};

const fileToDataUrl = (file: File) => new Promise<string>((resolve, reject) => {
  const reader = new FileReader();
  reader.onload = () => resolve(String(reader.result ?? ''));
  reader.onerror = () => reject(new Error('Не удалось прочитать файл'));
  reader.readAsDataURL(file);
});

export const UpdateAvatarForm = ({ onUpdated }: UpdateAvatarFormProps) => {
  const user = useAuthStore((s) => s.user);
  const [avatarUrl, setAvatarUrl] = useState('');
  const [fileName, setFileName] = useState<string | null>(null);
  const [status, setStatus] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [processing, setProcessing] = useState(false);

  const onFileChange = async (event: ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    setStatus(null);
    setError(null);

    if (!file) {
      setAvatarUrl('');
      setFileName(null);
      return;
    }

    if (!file.type.startsWith('image/')) {
      setError('Можно загрузить только изображение.');
      event.target.value = '';
      return;
    }

    if (file.size > 2 * 1024 * 1024) {
      setError('Максимальный размер аватарки — 2MB.');
      event.target.value = '';
      return;
    }

    setProcessing(true);
    try {
      const dataUrl = await fileToDataUrl(file);
      setAvatarUrl(dataUrl);
      setFileName(file.name);
    } catch (e) {
      setError(parseError(e));
    } finally {
      setProcessing(false);
    }
  };

  const onSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setStatus(null);
    setError(null);

    if (!user?.id) {
      setError('Не найден user id в сессии. Перелогинься.');
      return;
    }

    if (!avatarUrl) {
      setError('Сначала выбери изображение с компьютера.');
      return;
    }

    try {
      await api.post(`/users/${user.id}/avatar`, { avatar_url: avatarUrl });
      setStatus('Аватар обновлён.');
      onUpdated?.();
    } catch (e) {
      setError(parseError(e));
    }
  };

  return (
    <form className="theme-card space-y-3 rounded-2xl border border-[rgb(var(--border))] p-5 shadow-sm" onSubmit={onSubmit}>
      <h2 className="text-lg font-semibold text-[rgb(var(--text-primary))]">Обновить аватар</h2>
      <label className="block text-sm text-[rgb(var(--text-primary))]" htmlFor="avatar-file">Загрузить с компьютера</label>
      <input id="avatar-file" type="file" accept="image/*" className="form-input w-full rounded-lg px-3 py-2" onChange={onFileChange} />

      {fileName && <p className="text-xs text-[rgb(var(--text-secondary))]">Выбран файл: {fileName}</p>}
      {processing && <p className="text-sm text-[rgb(var(--text-secondary))]">Обрабатываем изображение...</p>}

      {avatarUrl && (
        <div className="space-y-2">
          <p className="text-xs text-[rgb(var(--text-secondary))]">Предпросмотр</p>
          <img src={avatarUrl} alt="Предпросмотр аватарки" className="h-20 w-20 rounded-full border border-[rgb(var(--border))] object-cover" />
        </div>
      )}

      {status && <p className="text-sm text-emerald-700">{status}</p>}
      {error && <p className="field-error">{error}</p>}
      <button className="interactive-chip theme-button px-4 py-2" disabled={processing}>Сохранить</button>
    </form>
  );
};
