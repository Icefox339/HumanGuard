import { api } from '@/api/client';

export type UserRole = 'user' | 'admin' | string;

export type UserDetails = {
  id: string;
  email: string;
  name: string;
  avatar_url?: string | null;
  role: UserRole;
  oauth_provider?: string | null;
  created_at?: string;
  updated_at?: string;
  last_login?: string | null;
};

export type CreateUserPayload = {
  email: string;
  name: string;
  password_hash: string;
  role: UserRole;
};

export type UpdateUserPayload = {
  name?: string;
  role?: UserRole;
};

export const createUser = (payload: CreateUserPayload) =>
  api.post<UserDetails>('/users', payload).then(({ data }) => data);

export const getUserById = (id: string) =>
  api.get<UserDetails>(`/users/${id}`).then(({ data }) => data);

export const getUserByEmail = (email: string) =>
  api.get<UserDetails>(`/users/email/${encodeURIComponent(email)}`).then(({ data }) => data);

export const checkEmailExists = (email: string) =>
  api.get<{ exists: boolean }>(`/users/exists?email=${encodeURIComponent(email)}`).then(({ data }) => data);

export const changeUserPassword = (id: string, payload: { old_password: string; new_password: string }) =>
  api.post<void>(`/users/${id}/password`, payload).then(({ data }) => data);

export const updateUser = (id: string, payload: UpdateUserPayload) =>
  api.put<UserDetails>(`/users/${id}`, payload).then(({ data }) => data);

export const deleteUser = (id: string) =>
  api.delete<void>(`/users/${id}`).then(({ data }) => data);
