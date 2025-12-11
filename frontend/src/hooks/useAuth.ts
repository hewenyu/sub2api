import { useNavigate } from 'react-router-dom';
import { useAuthStore } from '@/store/authStore';
import { authApi } from '@/api';
import { ROUTES } from '@/utils/constants';
import type { LoginRequest, LoginResponse } from '@/types';

export const useAuth = () => {
  const navigate = useNavigate();
  const { token, admin, isAuthenticated, setAuth, logout: clearAuth } = useAuthStore();

  const login = async (credentials: LoginRequest) => {
    const response = await authApi.login(credentials);
    const loginData = response.data as unknown as LoginResponse;
    setAuth(loginData.token, loginData.admin);
    navigate(ROUTES.DASHBOARD);
  };

  const logout = async () => {
    try {
      authApi.logout();
    } catch (error) {
      console.error('Logout error:', error);
    } finally {
      clearAuth();
      navigate(ROUTES.LOGIN);
    }
  };

  return {
    token,
    admin,
    isAuthenticated,
    login,
    logout,
  };
};
