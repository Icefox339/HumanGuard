import { api } from '@/api/client';

export type SiteModuleSettings = {
  collector: Record<string, unknown>;
  analyzer: Record<string, unknown>;
  reaction: Record<string, unknown>;
};

export const getSiteSettings = (siteId: string) =>
  api.get<SiteModuleSettings>(`/sites/${siteId}/settings`).then(({ data }) => data);

export const updateSiteSettings = (siteId: string, payload: Record<string, unknown>) =>
  api.put(`/sites/${siteId}/settings`, payload).then(({ data }) => data);
