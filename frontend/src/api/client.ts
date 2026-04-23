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
