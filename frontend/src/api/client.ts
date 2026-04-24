import axios from 'axios';
import { API_URL } from '@/lib/constants';
import { useAuthStore } from '@/app/store/auth-store';

export const api = axios.create({
  baseURL: `${API_URL}/api`
});

api.interceptors.request.use((config) => {
  const token = useAuthStore.getState().accessToken;
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

api.interceptors.response.use(
  (response) => response,
  (error) => {
    const status = error?.response?.status as number | undefined;
    if (status === 401) {
      useAuthStore.getState().clearSession();
      if (typeof window !== 'undefined' && !window.location.pathname.startsWith('/auth/')) {
        window.location.assign('/auth/login');
      }
    }

    return Promise.reject(error);
  }
);
