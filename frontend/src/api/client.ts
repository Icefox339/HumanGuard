import axios from 'axios';
import { API_URL } from '@/lib/constants';
import { useAuthStore } from '@/app/store/auth-store';

let csrfToken: string | null = null;
let csrfRequest: Promise<string | null> | null = null;

function shouldSkipCSRF(url?: string): boolean {
  if (!url) {
    return false;
  }

  return url === '/csrf' || url === '/login' || url === '/users' || url.startsWith('/auth/');
}

async function getCSRFToken(): Promise<string | null> {
  if (csrfToken) {
    return csrfToken;
  }

  if (!csrfRequest) {
    csrfRequest = api
      .get<{ csrf_token: string }>('/csrf')
      .then((response) => {
        csrfToken = response.data.csrf_token;
        return csrfToken;
      })
      .catch(() => null)
      .finally(() => {
        csrfRequest = null;
      });
  }

  return csrfRequest;
}

export const api = axios.create({
  baseURL: `${API_URL}/api`,
  withCredentials: true
});

api.interceptors.request.use(async (config) => {
  const token = useAuthStore.getState().accessToken;
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }

  const method = config.method?.toUpperCase();
  if (method && !['GET', 'HEAD', 'OPTIONS'].includes(method) && !shouldSkipCSRF(config.url)) {
    const token = await getCSRFToken();
    if (token) {
      config.headers['X-CSRF-Token'] = token;
    }
  }

  return config;
});
