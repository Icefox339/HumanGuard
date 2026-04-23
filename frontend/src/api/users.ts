import { api } from '@/api/client';
import { ApiResponse, User } from '@/api/types';

export const getUsers = () => api.get<ApiResponse<User[]>>('/users').then(({ data }) => data.data);
