import { api } from '@/api/client';

export type ManagedFile = {
  id: string;
  name: string;
  original_name?: string;
  size: number;
  mime_type?: string;
  created_at?: string;
};

export type CreateFileSharePayload = {
  file_id: string;
  expires_in_hours?: number;
};

export type CreateFileShareResponse = {
  token: string;
  share_url: string;
};

export const getFiles = () => api.get<ManagedFile[]>('/files').then(({ data }) => data);

export const uploadFile = (
  file: File,
  onProgress?: (progress: number) => void
) => {
  const formData = new FormData();
  formData.append('file', file);

  return api
    .post<ManagedFile>('/files/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
      onUploadProgress: (event) => {
        if (!event.total || !onProgress) {
          return;
        }
        onProgress(Math.round((event.loaded / event.total) * 100));
      },
      maxBodyLength: Infinity,
      maxContentLength: Infinity
    })
    .then(({ data }) => data);
};

export const createFileShare = (payload: CreateFileSharePayload) =>
  api.post<CreateFileShareResponse>('/files/share', payload).then(({ data }) => data);
