import { request } from '@/utils/request';
import type {
  Admin,
  APIKey,
  CodexAccount,
  UsageRecord,
  LoginRequest,
  LoginResponse,
  PaginatedResponse,
  PaginationParams,
} from '@/types';
import type {
  CreateAdminRequest,
  UpdateAdminRequest,
  CreateAPIKeyRequest,
  UpdateAPIKeyRequest,
  CreateCodexAccountRequest,
  UpdateCodexAccountRequest,
  UsageStatsFilters,
} from '@/types/admin';

export const authApi = {
  login: async (data: LoginRequest) => {
    const response = await request.post<LoginResponse>('/admin/login', data);
    return response;
  },

  getInfo: async () => {
    const response = await request.get<Admin>('/admin/info');
    return response;
  },

  logout: () => {
    localStorage.removeItem('auth_token');
    window.location.href = '/login';
  },
};

export const apiKeyApi = {
  list: (params?: PaginationParams) => request.get<PaginatedResponse<APIKey>>('/api-keys', params),

  get: (id: number) => request.get<APIKey>(`/api-keys/${id}`),

  create: (data: CreateAPIKeyRequest) => request.post<APIKey>('/api-keys', data),

  update: (id: number, data: UpdateAPIKeyRequest) => request.put<APIKey>(`/api-keys/${id}`, data),

  delete: (id: number) => request.delete(`/api-keys/${id}`),
};

export const codexAccountApi = {
  list: (params?: PaginationParams) =>
    request.get<PaginatedResponse<CodexAccount>>('/codex-accounts', params),

  get: (id: number) => request.get<CodexAccount>(`/codex-accounts/${id}`),

  create: (data: CreateCodexAccountRequest) => request.post<CodexAccount>('/codex-accounts', data),

  update: (id: number, data: UpdateCodexAccountRequest) =>
    request.put<CodexAccount>(`/codex-accounts/${id}`, data),

  delete: (id: number) => request.delete(`/codex-accounts/${id}`),
};

export const usageRecordApi = {
  list: (params?: PaginationParams & UsageStatsFilters) =>
    request.get<PaginatedResponse<UsageRecord>>('/usage-records', params),

  get: (id: number) => request.get<UsageRecord>(`/usage-records/${id}`),
};

export const adminApi = {
  list: (params?: PaginationParams) => request.get<PaginatedResponse<Admin>>('/admins', params),

  get: (id: number) => request.get<Admin>(`/admins/${id}`),

  create: (data: CreateAdminRequest) => request.post<Admin>('/admins', data),

  update: (id: number, data: UpdateAdminRequest) => request.put<Admin>(`/admins/${id}`, data),

  delete: (id: number) => request.delete(`/admins/${id}`),
};
