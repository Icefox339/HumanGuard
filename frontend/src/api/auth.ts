import { api } from '@/api/client';
import { User } from '@/api/types';
import { UserDetails } from '@/api/users';

export type LoginPayload = {
  email: string;
  password: string;
  totp_code?: string;
};

export type RegisterPayload = {
  email: string;
  password: string;
  name?: string;
};

export type LoginResponse = {
  token: string;
  user: User;
};

export type RegisterResponse = {
  user: User;
  totp_secret: string;
  qr_code_url: string;
  message: string;
};

export const login = (payload: LoginPayload) =>
  api.post<LoginResponse | UserDetails>('/login', payload).then(({ data }) => data);

export const register = (payload: RegisterPayload) =>
  api.post<RegisterResponse>('/users', payload).then(({ data }) => data);

export const getCurrentUser = () => api.get<User>('/me').then(({ data }) => data);
