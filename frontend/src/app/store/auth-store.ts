import { create } from 'zustand';
import { storage } from '@/lib/storage';

type User = {
  id: string;
  email: string;
  role: 'user' | 'admin';
};

type AuthState = {
  accessToken: string | null;
  user: User | null;
  isAuthenticated: boolean;
  setSession: (token: string, user: User) => void;
  clearSession: () => void;
};

const tokenKey = 'hg_access_token';

export const useAuthStore = create<AuthState>((set) => ({
  accessToken: storage.get<string>(tokenKey) ?? null,
  user: null,
  isAuthenticated: Boolean(storage.get<string>(tokenKey)),
  setSession: (accessToken, user) => {
    storage.set(tokenKey, accessToken);
    set({ accessToken, user, isAuthenticated: true });
  },
  clearSession: () => {
    storage.remove(tokenKey);
    set({ accessToken: null, user: null, isAuthenticated: false });
  }
}));
