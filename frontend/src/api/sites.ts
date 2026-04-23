import { api } from '@/api/client';
import { Site } from '@/api/types';

export const getSites = () => api.get<Site[]>('/sites').then(({ data }) => data);
