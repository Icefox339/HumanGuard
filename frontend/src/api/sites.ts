import { api } from '@/api/client';

export type SiteStatus = 'active' | 'suspended' | 'verifying' | string;

export type Site = {
  id: string;
  user_id: string;
  name: string;
  domain: string;
  origin_server: string;
  status: SiteStatus;
  created_at?: string;
  updated_at?: string;
};

export type CreateSitePayload = {
  user_id: string;
  name: string;
  domain: string;
  origin_server: string;
};

export const getSites = () => api.get<Site[]>('/sites').then(({ data }) => data);

export const createSite = (payload: CreateSitePayload) =>
  api.post<Site>('/sites', payload).then(({ data }) => data);

export const deleteSite = (id: string) =>
  api.delete<void>(`/sites/${id}`).then(({ data }) => data);

export const activateSite = (id: string) =>
  api.post<void>(`/sites/${id}/activate`).then(({ data }) => data);

export const suspendSite = (id: string) =>
  api.post<void>(`/sites/${id}/suspend`).then(({ data }) => data);
