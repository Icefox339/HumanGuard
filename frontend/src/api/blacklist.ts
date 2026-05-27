import { api } from '@/api/client';

export type BlacklistEntry = {
  id: string;
  site_id: string;
  ip: string;
  reason: string;
  created_at: string;
  expires_at?: string | null;
};

export const getSiteBlacklist = (siteId: string) =>
  api.get<BlacklistEntry[]>(`/sites/${siteId}/blacklist`).then(({ data }) => data);

export const addSiteBlacklistIP = (siteId: string, payload: { ip: string; reason?: string }) =>
  api.post<BlacklistEntry>(`/sites/${siteId}/blacklist`, payload).then(({ data }) => data);

export const removeSiteBlacklistIP = (siteId: string, ip: string) =>
  api.delete<void>(`/sites/${siteId}/blacklist/${encodeURIComponent(ip)}`).then(({ data }) => data);
