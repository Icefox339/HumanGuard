import { api } from '@/api/client';
import { ApiResponse, User } from '@/api/types';

export const login = (payload: { email: string; password: string; otp?: string }) =>
  api.post<ApiResponse<{ token: string; user: User }>>('/login', payload).then(({ data }) => data.data);

export const register = (payload: { email: string; password: string }) =>
  api.post<ApiResponse<User>>('/users', payload).then(({ data }) => data.data);
