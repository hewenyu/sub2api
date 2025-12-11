import { create } from 'zustand';
import type { Admin } from '@/types';
import { TOKEN_KEY } from '@/utils/constants';

interface AuthState {
  token: string | null;
  admin: Admin | null;
  isAuthenticated: boolean;
  setAuth: (token: string, admin: Admin) => void;
  setToken: (token: string) => void;
  setAdmin: (admin: Admin) => void;
  logout: () => void;
  initAuth: () => void;
  getToken: () => string | null;
}

export const useAuthStore = create<AuthState>((set, get) => ({
  token: null,
  admin: null,
  isAuthenticated: false,

  setAuth: (token: string, admin: Admin) => {
    localStorage.setItem(TOKEN_KEY, token);
    set({ token, admin, isAuthenticated: true });
  },

  setToken: (token: string) => {
    localStorage.setItem(TOKEN_KEY, token);
    set({ token, isAuthenticated: true });
  },

  setAdmin: (admin: Admin) => {
    set({ admin });
  },

  logout: () => {
    localStorage.removeItem(TOKEN_KEY);
    set({ token: null, admin: null, isAuthenticated: false });
  },

  initAuth: () => {
    const token = localStorage.getItem(TOKEN_KEY);
    if (token) {
      set({ token, isAuthenticated: true });
    }
  },

  getToken: () => get().token,
}));
