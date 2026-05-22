import { api } from '@/api/client';
import { ApiResponse, Session } from '@/api/types';

export const getSiteSessions = (siteId: string) =>
  api.get<ApiResponse<Session[]>>(`/sites/${siteId}/sessions`).then(({ data }) => data.data);

export type AdminUserSession = {
  id: string;
  user_id: string;
  email: string;
  role: string;
  created_at: string;
  last_seen: string;
  expires_at: string;
  ip: string;
  user_agent: string;
};

type AdminSessionsResponse = {
  total: number;
  sessions: AdminUserSession[];
};

export const getAdminUserSessions = () =>
  api.get<AdminSessionsResponse>('/admin/users/sessions').then(({ data }) => data.sessions);

export const deactivateUserSession = (sessionId: string) =>
  api.delete(`/admin/users/sessions/${sessionId}`).then(({ data }) => data);
