import { api } from '@/api/client';

export const updateSiteSettings = (siteId: string, payload: Record<string, unknown>) =>
  api.put(`/sites/${siteId}/settings`, payload).then(({ data }) => data);
