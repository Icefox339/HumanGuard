import { api } from '@/api/client';
import { Session } from '@/api/types';

export const getSiteSessions = (siteId: string) => api.get<Session[]>(`/sites/${siteId}/sessions`).then(({ data }) => data);
