import axios, { AxiosError } from 'axios';
import type {
  AxiosInstance,
  InternalAxiosRequestConfig,
  AxiosResponse,
} from 'axios';
import { API_BASE_URL, TOKEN_KEY } from './constants';
import type { ApiResponse } from '@/types';

const instance: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
});

instance.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = localStorage.getItem(TOKEN_KEY);
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error: AxiosError) => {
    return Promise.reject(error);
  }
);

instance.interceptors.response.use(
  (response: AxiosResponse<ApiResponse>) => {
    return response;
  },
  (error: AxiosError<ApiResponse>) => {
    if (error.response) {
      const { status, data } = error.response;

      if (status === 401) {
        localStorage.removeItem(TOKEN_KEY);
        window.location.href = '/login';
        return Promise.reject(new Error('Unauthorized. Please login again.'));
      }

      if (data && data.message) {
        return Promise.reject(new Error(data.message));
      }

      switch (status) {
        case 400:
          return Promise.reject(new Error('Bad request'));
        case 403:
          return Promise.reject(new Error('Forbidden'));
        case 404:
          return Promise.reject(new Error('Resource not found'));
        case 500:
          return Promise.reject(new Error('Internal server error'));
        default:
          return Promise.reject(new Error('An error occurred'));
      }
    }

    if (error.request) {
      return Promise.reject(new Error('No response from server'));
    }

    return Promise.reject(error);
  }
);

export default instance;

export const request = {
  get: <T = unknown>(url: string, params?: object) =>
    instance.get<ApiResponse<T>>(url, { params }),

  post: <T = unknown>(url: string, data?: object) =>
    instance.post<ApiResponse<T>>(url, data),

  put: <T = unknown>(url: string, data?: object) =>
    instance.put<ApiResponse<T>>(url, data),

  delete: <T = unknown>(url: string) =>
    instance.delete<ApiResponse<T>>(url),

  patch: <T = unknown>(url: string, data?: object) =>
    instance.patch<ApiResponse<T>>(url, data),
};
