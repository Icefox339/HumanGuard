import { api } from '@/api/client';
import { ApiResponse, Site } from '@/api/types';

export const getSites = () => api.get<ApiResponse<Site[]>>('/sites').then(({ data }) => data.data);
