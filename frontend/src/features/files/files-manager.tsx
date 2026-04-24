import { ChangeEvent, useEffect, useState } from 'react';
import { AxiosError } from 'axios';
import { API_URL } from '@/lib/constants';
import { getFiles, ManagedFile, uploadFile } from '@/api/files';

const formatBytes = (bytes: number) => {
  if (bytes === 0) {
    return '0 B';
  }
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(1024));
  return `${(bytes / 1024 ** i).toFixed(i > 1 ? 2 : 0)} ${units[i]}`;
};

const parseError = (error: unknown) => {
  const err = error as AxiosError<{ error?: string }>;
  return err.response?.data?.error ?? err.message ?? 'Unknown error';
};

const getDownloadUrl = (file: ManagedFile) => `${API_URL}/api/files/${file.id}`;

export const FilesManager = () => {
  const [files, setFiles] = useState<ManagedFile[]>([]);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [loading, setLoading] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [uploadProgress, setUploadProgress] = useState(0);
  const [error, setError] = useState<string | null>(null);

  const loadFiles = async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await getFiles();
      setFiles(data);
    } catch (e) {
      setError(parseError(e));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void loadFiles();
  }, []);

  const onChooseFile = (event: ChangeEvent<HTMLInputElement>) => {
    setSelectedFile(event.target.files?.[0] ?? null);
    setUploadProgress(0);
    setError(null);
  };

  const onUpload = async () => {
    if (!selectedFile) {
      setError('Сначала выберите файл.');
      return;
    }

    setError(null);
    setUploading(true);
    setUploadProgress(0);

    try {
      const uploaded = await uploadFile(selectedFile, setUploadProgress);
      setFiles((prev) => [uploaded, ...prev]);
      setSelectedFile(null);
    } catch (e) {
      setError(parseError(e));
    } finally {
      setUploading(false);
    }
  };

  return (
    <section className="space-y-4">
      <header className="theme-card rounded-2xl border border-[rgb(var(--border))] p-5 shadow-sm">
        <h2 className="text-lg font-semibold text-[rgb(var(--text-primary))]">Files</h2>
        <p className="mt-1 text-sm text-[rgb(var(--text-secondary))]">
          Здесь отображаются только ваши файлы. Поддерживается загрузка больших файлов (&gt; 1 GB).
        </p>

        <div className="mt-4 flex flex-wrap items-center gap-3">
          <input className="form-input max-w-md rounded-lg px-3 py-2" onChange={onChooseFile} type="file" />
          <button className="interactive-chip theme-button px-4 py-2 disabled:opacity-60" disabled={!selectedFile || uploading} onClick={() => void onUpload()} type="button">
            {uploading ? 'Загрузка…' : 'Загрузить файл'}
          </button>
        </div>

        {selectedFile && (
          <p className="mt-2 text-sm text-[rgb(var(--text-secondary))]">
            Выбран: {selectedFile.name} ({formatBytes(selectedFile.size)})
          </p>
        )}

        {(uploading || uploadProgress > 0) && (
          <div className="mt-3">
            <div className="h-2 w-full overflow-hidden rounded-full bg-[rgb(var(--bg-muted))]">
              <div className="h-full bg-blue-500 transition-all" style={{ width: `${uploadProgress}%` }} />
            </div>
            <p className="mt-1 text-xs text-[rgb(var(--text-secondary))]">{uploadProgress}%</p>
          </div>
        )}

        {error && <p className="mt-3 field-error">{error}</p>}
      </header>

      <section className="theme-card rounded-2xl border border-[rgb(var(--border))] p-5 shadow-sm">
        <div className="mb-3 flex items-center justify-between">
          <h3 className="text-base font-semibold text-[rgb(var(--text-primary))]">
            Мои файлы
          </h3>
          <button className="interactive-chip rounded border border-[rgb(var(--border))] px-3 py-1 text-sm text-[rgb(var(--text-primary))]" onClick={() => void loadFiles()} type="button">
            Обновить
          </button>
        </div>

        {loading && <p className="text-sm text-[rgb(var(--text-secondary))]">Загрузка...</p>}
        {!loading && files.length === 0 && (
          <p className="text-sm text-[rgb(var(--text-secondary))]">Файлов пока нет.</p>
        )}

        <div className="space-y-3">
          {files.map((file) => (
            <article className="rounded border border-[rgb(var(--border))] p-3" key={file.id}>
              <div className="flex flex-wrap items-center justify-between gap-3">
                <div>
                  <p className="font-medium text-[rgb(var(--text-primary))]">{file.original_name ?? file.name}</p>
                  <p className="text-sm text-[rgb(var(--text-secondary))]">{formatBytes(file.size)}</p>
                  <p className="text-xs text-[rgb(var(--text-secondary))]">
                    {file.created_at ? new Date(file.created_at).toLocaleString() : 'Дата неизвестна'}
                  </p>
                </div>
                <a className="interactive-chip rounded border border-[rgb(var(--border))] px-3 py-1 text-sm text-[rgb(var(--text-primary))]" download href={getDownloadUrl(file)} rel="noreferrer" target="_blank">
                  Скачать
                </a>
              </div>
            </article>
          ))}
        </div>
      </section>
    </section>
  );
};
