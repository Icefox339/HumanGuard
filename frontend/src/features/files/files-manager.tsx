import { ChangeEvent, useEffect, useState } from 'react';
import { AxiosError } from 'axios';
import { API_URL } from '@/lib/constants';
import { createFileShare, getFiles, ManagedFile, uploadFile } from '@/api/files';

const formatBytes = (bytes: number) => {
  if (bytes === 0) {
    return '0 B';
  }
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(1024));
  return `${(bytes / 1024 ** i).toFixed(i > 1 ? 2 : 0)} ${units[i]}`;
};

type ErrorDetails = {
  operation: 'upload' | 'list' | 'share';
  message: string;
  status?: number;
  code?: string;
  method?: string;
  url?: string;
  backendError?: string;
  fileName?: string;
  fileSize?: number;
};

const buildErrorDetails = (
  operation: 'upload' | 'list' | 'share',
  error: unknown,
  file?: File | null
): ErrorDetails => {
  const err = error as AxiosError<{ error?: string }>;
  const response = err.response;
  return {
    operation,
    message: response?.data?.error ?? err.message ?? 'Unknown error',
    status: response?.status,
    code: err.code,
    method: err.config?.method?.toUpperCase(),
    url: err.config?.url,
    backendError: response?.data?.error,
    fileName: file?.name,
    fileSize: file?.size
  };
};

const getDownloadUrl = (file: ManagedFile) => `${API_URL}/api/files/${file.id}`;

export const FilesManager = () => {
  const [files, setFiles] = useState<ManagedFile[]>([]);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [loading, setLoading] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [uploadProgress, setUploadProgress] = useState(0);
  const [error, setError] = useState<string | null>(null);
  const [errorDetails, setErrorDetails] = useState<ErrorDetails | null>(null);
  const [sharingId, setSharingId] = useState<string | null>(null);
  const [shareLinks, setShareLinks] = useState<Record<string, string>>({});
  const [copiedShareId, setCopiedShareId] = useState<string | null>(null);

  const loadFiles = async () => {
    setLoading(true);
    setError(null);
    setErrorDetails(null);
    try {
      const data = await getFiles();
      setFiles(data);
    } catch (e) {
      const details = buildErrorDetails('list', e);
      setError(details.message);
      setErrorDetails(details);
      console.error('[FilesManager] list failed', details, e);
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
    setErrorDetails(null);
  };

  const onUpload = async () => {
    if (!selectedFile) {
      setError('Сначала выберите файл.');
      return;
    }

    setError(null);
    setErrorDetails(null);
    setUploading(true);
    setUploadProgress(0);

    try {
      const uploaded = await uploadFile(selectedFile, setUploadProgress);
      setFiles((prev) => [uploaded, ...prev]);
      setSelectedFile(null);
    } catch (e) {
      const details = buildErrorDetails('upload', e, selectedFile);
      setError(details.message);
      setErrorDetails(details);
      console.error('[FilesManager] upload failed', details, e);
    } finally {
      setUploading(false);
    }
  };

  const normalizeShareUrl = (shareUrl: string) => {
    try {
      const url = new URL(shareUrl);
      const apiBase = new URL(API_URL);
      if (url.hostname === 'localhost') {
        url.protocol = apiBase.protocol;
        url.host = apiBase.host;
      }
      return url.toString();
    } catch {
      return shareUrl;
    }
  };

  const onShareFile = async (fileId: string) => {
    setError(null);
    setErrorDetails(null);
    setSharingId(fileId);
    try {
      const share = await createFileShare({ file_id: fileId });
      const shareUrl = normalizeShareUrl(share.share_url);
      setShareLinks((prev) => ({ ...prev, [fileId]: shareUrl }));
      await navigator.clipboard.writeText(shareUrl);
      setCopiedShareId(fileId);
      setTimeout(() => setCopiedShareId((current) => (current === fileId ? null : current)), 1800);
    } catch (e) {
      const details = buildErrorDetails('share', e);
      setError(details.message);
      setErrorDetails(details);
      console.error('[FilesManager] share failed', details, e);
    } finally {
      setSharingId(null);
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
        {errorDetails && (
          <details className="mt-3 rounded-lg border border-[rgb(var(--border))] p-3 text-xs text-[rgb(var(--text-secondary))]">
            <summary className="cursor-pointer select-none text-sm font-medium text-[rgb(var(--text-primary))]">
              Подробности ошибки
            </summary>
            <ul className="mt-2 space-y-1">
              <li>Операция: {errorDetails.operation}</li>
              <li>Status: {errorDetails.status ?? 'n/a'}</li>
              <li>Code: {errorDetails.code ?? 'n/a'}</li>
              <li>Method: {errorDetails.method ?? 'n/a'}</li>
              <li>URL: {errorDetails.url ?? 'n/a'}</li>
              <li>Backend message: {errorDetails.backendError ?? 'n/a'}</li>
              <li>File: {errorDetails.fileName ?? 'n/a'}</li>
              <li>File size: {typeof errorDetails.fileSize === 'number' ? formatBytes(errorDetails.fileSize) : 'n/a'}</li>
            </ul>
          </details>
        )}
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
                <div className="flex flex-wrap items-center gap-2">
                  <button
                    className="interactive-chip rounded border border-[rgb(var(--border))] px-3 py-1 text-sm text-[rgb(var(--text-primary))] disabled:opacity-60"
                    disabled={sharingId === file.id}
                    onClick={() => void onShareFile(file.id)}
                    type="button"
                  >
                    {sharingId === file.id ? 'Генерация...' : 'Поделиться URL'}
                  </button>
                  <a className="interactive-chip rounded border border-[rgb(var(--border))] px-3 py-1 text-sm text-[rgb(var(--text-primary))]" download href={getDownloadUrl(file)} rel="noreferrer" target="_blank">
                    Скачать
                  </a>
                </div>
              </div>
              {shareLinks[file.id] && (
                <div className="mt-3 rounded-lg border border-[rgb(var(--border))] bg-[rgb(var(--bg-main))] p-2">
                  <p className="mb-1 text-xs text-[rgb(var(--text-secondary))]">Публичная ссылка для скачивания:</p>
                  <div className="flex flex-wrap items-center gap-2">
                    <input className="form-input min-w-[16rem] flex-1 rounded px-2 py-1 text-xs" readOnly value={shareLinks[file.id]} />
                    <button
                      className="interactive-chip rounded border border-[rgb(var(--border))] px-2 py-1 text-xs text-[rgb(var(--text-primary))]"
                      onClick={() => void navigator.clipboard.writeText(shareLinks[file.id])}
                      type="button"
                    >
                      Копировать
                    </button>
                    {copiedShareId === file.id && <span className="text-xs text-emerald-700">Скопировано</span>}
                  </div>
                </div>
              )}
            </article>
          ))}
        </div>
      </section>
    </section>
  );
};
