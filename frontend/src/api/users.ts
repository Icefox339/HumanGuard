import { api } from '@/api/client';
import { User } from '@/api/types';

export const getUsers = () => api.get<User[]>('/users').then(({ data }) => data);
