import { ChangeEvent, FormEvent, useEffect, useState } from 'react';
import { AxiosError } from 'axios';
import { uploadAvatarFile } from '@/api/files';
import { api } from '@/api/client';
import { useAuthStore } from '@/app/store/auth-store';

const parseError = (error: unknown) => {
  const err = error as AxiosError<{ error?: string }>;
  return err.response?.data?.error ?? err.message ?? 'Unknown error';
};

type UpdateAvatarFormProps = {
  onUpdated?: () => void;
};

export const UpdateAvatarForm = ({ onUpdated }: UpdateAvatarFormProps) => {
  const user = useAuthStore((s) => s.user);
  const [avatarFile, setAvatarFile] = useState<File | null>(null);
  const [previewUrl, setPreviewUrl] = useState('');
  const [fileName, setFileName] = useState<string | null>(null);
  const [status, setStatus] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [processing, setProcessing] = useState(false);

  const updatePreviewUrl = (nextPreviewUrl: string) => {
    setPreviewUrl((previousPreviewUrl) => {
      if (previousPreviewUrl) {
        URL.revokeObjectURL(previousPreviewUrl);
      }

      return nextPreviewUrl;
    });
  };

  const clearSelectedAvatar = () => {
    setAvatarFile(null);
    updatePreviewUrl('');
    setFileName(null);
  };

  useEffect(
    () => () => {
      if (previewUrl) {
        URL.revokeObjectURL(previewUrl);
      }
    },
    [previewUrl],
  );

  const onFileChange = (event: ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    setStatus(null);
    setError(null);

    if (!file) {
      clearSelectedAvatar();
      return;
    }

    if (!file.type.startsWith('image/')) {
      clearSelectedAvatar();
      setError('Можно загрузить только изображение.');
      event.target.value = '';
      return;
    }

    if (file.size > 15 * 1024 * 1024) {
      clearSelectedAvatar();
      setError('Максимальный размер аватарки — 15MB.');
      event.target.value = '';
      return;
    }

    setAvatarFile(file);
    updatePreviewUrl(URL.createObjectURL(file));
    setFileName(file.name);
  };

  const onSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setStatus(null);
    setError(null);

    if (!user?.id) {
      setError('Не найден user id в сессии. Перелогинься.');
      return;
    }

    if (!avatarFile) {
      setError('Сначала выбери изображение с компьютера.');
      return;
    }

    setProcessing(true);
    try {
      const avatarUrl = await uploadAvatarFile(avatarFile);
      await api.post(`/users/${user.id}/avatar`, { avatar_url: avatarUrl });
      setStatus('Аватар загружен, в профиле сохранён URL.');
      onUpdated?.();
    } catch (e) {
      setError(parseError(e));
    } finally {
      setProcessing(false);
    }
  };

  return (
    <form
      className="theme-card space-y-3 rounded-2xl border border-[rgb(var(--border))] p-5 shadow-sm"
      onSubmit={onSubmit}
    >
      <h2 className="text-lg font-semibold text-[rgb(var(--text-primary))]">
        Обновить аватар
      </h2>
      <label
        className="block text-sm text-[rgb(var(--text-primary))]"
        htmlFor="avatar-file"
      >
        Загрузить с компьютера
      </label>
      <input
        id="avatar-file"
        type="file"
        accept="image/*"
        className="form-input w-full rounded-lg px-3 py-2"
        onChange={onFileChange}
      />

      {fileName && (
        <p className="text-xs text-[rgb(var(--text-secondary))]">
          Выбран файл: {fileName}
        </p>
      )}
      {processing && (
        <p className="text-sm text-[rgb(var(--text-secondary))]">
          Загружаем изображение и сохраняем URL...
        </p>
      )}

      {previewUrl && (
        <div className="space-y-2">
          <p className="text-xs text-[rgb(var(--text-secondary))]">
            Предпросмотр
          </p>
          <img
            src={previewUrl}
            alt="Предпросмотр аватарки"
            className="h-20 w-20 rounded-full border border-[rgb(var(--border))] object-cover"
          />
        </div>
      )}

      {status && <p className="text-sm text-emerald-700">{status}</p>}
      {error && <p className="field-error">{error}</p>}
      <button
        className="interactive-chip theme-button px-4 py-2"
        disabled={processing}
      >
        {processing ? 'Сохраняем...' : 'Сохранить'}
      </button>
    </form>
  );
};
