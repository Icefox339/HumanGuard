import { api } from '@/api/client';
import { ApiResponse, Session } from '@/api/types';

export const getSiteSessions = (siteId: string) =>
  api.get<ApiResponse<Session[]>>(`/sites/${siteId}/sessions`).then(({ data }) => data.data);
