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
  setUser: (user: User) => void;
  clearSession: () => void;
};

const tokenKey = 'hg_access_token';
const userKey = 'hg_user';

export const useAuthStore = create<AuthState>((set) => ({
  accessToken: storage.get<string>(tokenKey) ?? null,
  user: storage.get<User>(userKey) ?? null,
  isAuthenticated: Boolean(storage.get<string>(tokenKey)),
  setSession: (accessToken, user) => {
    storage.set(tokenKey, accessToken);
    storage.set(userKey, user);
    set({ accessToken, user, isAuthenticated: true });
  },
  setUser: (user) => {
    storage.set(userKey, user);
    set({ user });
  },
  clearSession: () => {
    storage.remove(tokenKey);
    storage.remove(userKey);
    set({ accessToken: null, user: null, isAuthenticated: false });
  }
}));
