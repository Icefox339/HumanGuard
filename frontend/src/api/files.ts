import { api } from '@/api/client';
import { API_URL } from '@/lib/constants';

export type ManagedFile = {
  id: string;
  name: string;
  original_name?: string;
  size: number;
  mime_type?: string;
  created_at?: string;
};

export type FileShareResponse = {
  token: string;
};

export const getFiles = () => api.get<ManagedFile[]>('/files').then(({ data }) => data);

export const createFileShare = (fileId: string, expiresInHours?: number) =>
  api.post<FileShareResponse>('/files/share', {
    file_id: fileId,
    ...(expiresInHours && expiresInHours > 0 ? { expires_in_hours: expiresInHours } : {})
  }).then(({ data }) => data);

export const uploadFile = (
  file: File,
  uploadId: string,
  onProgress?: (progress: number) => void
) => {
  const formData = new FormData();
  formData.append('file', file);

  return api
    .post<ManagedFile>(`/files/upload?upload_id=${encodeURIComponent(uploadId)}`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
      onUploadProgress: onProgress
        ? (event) => {
            if (!event.total) {
              return;
            }
            onProgress(Math.round((event.loaded / event.total) * 100));
          }
        : undefined,
      maxBodyLength: Infinity,
      maxContentLength: Infinity
    })
    .then(({ data }) => data);
};

export type UploadProgressMessage = {
  upload_id: string;
  bytes_done: number;
  total_bytes: number;
  percentage: number;
  completed: boolean;
};

export const openUploadProgressSocket = (
  uploadId: string,
  onMessage: (progress: UploadProgressMessage) => void
) => {
  const baseUrl = API_URL.replace(/\/$/, '');
  const wsBaseUrl = baseUrl.replace(/^http/, 'ws');
  const ws = new WebSocket(`${wsBaseUrl}/api/files/upload/progress?upload_id=${encodeURIComponent(uploadId)}`);

  ws.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data) as UploadProgressMessage;
      onMessage(data);
    } catch (error) {
      console.error('[FilesAPI] invalid websocket payload', error);
    }
  };

  return ws;
};
