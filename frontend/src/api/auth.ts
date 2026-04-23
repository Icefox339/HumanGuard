import { api } from '@/api/client';
import { User } from '@/api/types';

type LoginResponse = {
  token: string;
  user: User;
};

type RegisterResponse = {
  user: User;
  totp_secret?: string;
  qr_code_url?: string;
};

export const login = (payload: { email: string; password: string; otp?: string }) =>
  api
    .post<LoginResponse>('/login', {
      email: payload.email,
      password: payload.password,
      totp_code: payload.otp ?? ''
    })
    .then(({ data }) => data);

export const register = (payload: { email: string; password: string; name?: string }) =>
  api
    .post<RegisterResponse>('/users', {
      email: payload.email,
      password: payload.password,
      name: payload.name ?? ''
    })
    .then(({ data }) => data);
