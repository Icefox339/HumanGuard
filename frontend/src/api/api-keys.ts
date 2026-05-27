import { api } from '@/api/client';

export type ApiKey = {
  id: string;
  name: string;
  prefix: string;
  created_at: string;
  expires_at?: string;
  last_used_at?: string;
  revoked: boolean;
};

export type CreatedApiKey = ApiKey & {
  key: string;
};

export type AdminApiKey = ApiKey & {
  user_id: string;
  user_email: string;
  user_name: string;
};

export const listApiKeys = async () => {
  const response = await api.get<ApiKey[]>('/keys');
  return response.data;
};

export const createApiKey = async (payload: { name: string; expires_in_days?: number }) => {
  const response = await api.post<CreatedApiKey>('/keys', payload);
  return response.data;
};

export const revokeApiKey = async (id: string) => {
  await api.delete(`/keys/${id}`);
};

export const listAllApiKeys = async () => {
  const response = await api.get<AdminApiKey[]>('/admin/keys');
  return response.data;
};
