import { request } from '@/utils/request';
import type {
  CodexAccount,
  CreateCodexAccountRequest,
  UpdateCodexAccountRequest,
  AccountListResponse,
  AccountFilters,
  ProxyConfig,
} from '@/types/account';

export const codexApi = {
  /**
   * Create new Codex account
   */
  create: async (data: CreateCodexAccountRequest): Promise<CodexAccount> => {
    const response = await request.post<CodexAccount>('/admin/codex-accounts', data);
    return response.data.data || ({} as CodexAccount);
  },

  /**
   * List all Codex accounts
   */
  list: async (params?: AccountFilters): Promise<AccountListResponse<CodexAccount>> => {
    const response = await request.get<AccountListResponse<CodexAccount>>(
      '/admin/codex-accounts',
      params
    );
    return response.data.data || { items: [], total: 0 };
  },

  /**
   * Get single Codex account
   */
  get: async (id: number): Promise<CodexAccount> => {
    const response = await request.get<CodexAccount>(`/admin/codex-accounts/${id}`);
    return response.data.data || ({} as CodexAccount);
  },

  /**
   * Update Codex account
   */
  update: async (id: number, data: UpdateCodexAccountRequest): Promise<void> => {
    await request.put(`/admin/codex-accounts/${id}`, data);
  },

  /**
   * Delete Codex account
   */
  delete: async (id: number): Promise<void> => {
    await request.delete(`/admin/codex-accounts/${id}`);
  },

  /**
   * Toggle Codex account active status
   */
  toggle: async (id: number): Promise<{ is_active: boolean }> => {
    const response = await request.post<{ is_active: boolean }>(
      `/admin/codex-accounts/${id}/toggle`
    );
    return response.data.data || { is_active: false };
  },

  /**
   * Refresh Codex account token
   */
  refreshToken: async (id: number): Promise<void> => {
    await request.post(`/admin/codex-accounts/${id}/refresh-token`);
  },

  /**
   * Generate OAuth authorization URL
   */
  generateAuthURL: async (callbackPort: number = 1455): Promise<{
    auth_url: string;
    callback_url: string;
    state: string;
  }> => {
    const response = await request.post<{
      auth_url: string;
      callback_url: string;
      state: string;
    }>('/admin/codex-accounts/generate-auth-url', {
      callback_port: callbackPort,
    });
    return response.data.data || { auth_url: '', callback_url: '', state: '' };
  },

  /**
   * Verify OAuth authorization and create account
   */
  verifyAuth: async (
    code: string,
    state: string,
    accountData: {
      name: string;
      account_type: 'openai-oauth' | 'openai-responses';
      base_api?: string;
      priority: number;
      schedulable: boolean;
      proxy_config?: ProxyConfig;
      proxy_name?: string;
    }
  ): Promise<CodexAccount> => {
    const response = await request.post<CodexAccount>('/admin/codex-accounts/verify-auth', {
      code,
      state,
      account: accountData,
    });
    return response.data.data || ({} as CodexAccount);
  },
};
