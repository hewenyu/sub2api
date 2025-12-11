import { request } from '@/utils/request';
import type {
  ClaudeAccount,
  ClaudeOAuthResponse,
  CreateClaudeAccountData,
  UpdateClaudeAccountRequest,
  AccountListResponse,
  AccountFilters,
} from '@/types/account';

export const claudeApi = {
  /**
   * Generate OAuth authorization URL
   */
  generateAuthUrl: async (callbackPort: number = 8888): Promise<ClaudeOAuthResponse> => {
    const response = await request.post<ClaudeOAuthResponse>('/admin/claude-accounts/auth-url', {
      callback_port: callbackPort,
    });
    return response.data.data || { auth_url: '', callback_url: '', state: '' };
  },

  /**
   * Verify OAuth authorization code
   */
  verifyAuth: async (
    code: string,
    state: string,
    accountData: CreateClaudeAccountData
  ): Promise<ClaudeAccount> => {
    const response = await request.post<ClaudeAccount>('/admin/claude-accounts/verify-auth', {
      code,
      state,
      account_data: accountData,
    });
    return response.data.data || ({} as ClaudeAccount);
  },

  /**
   * List all Claude accounts
   */
  list: async (params?: AccountFilters): Promise<AccountListResponse<ClaudeAccount>> => {
    const response = await request.get<AccountListResponse<ClaudeAccount>>(
      '/admin/claude-accounts',
      params
    );
    return response.data.data || { items: [], total: 0 };
  },

  /**
   * Get single Claude account
   */
  get: async (id: number): Promise<ClaudeAccount> => {
    const response = await request.get<ClaudeAccount>(`/admin/claude-accounts/${id}`);
    return response.data.data || ({} as ClaudeAccount);
  },

  /**
   * Update Claude account
   */
  update: async (id: number, data: UpdateClaudeAccountRequest): Promise<void> => {
    await request.put(`/admin/claude-accounts/${id}`, data);
  },

  /**
   * Delete Claude account
   */
  delete: async (id: number): Promise<void> => {
    await request.delete(`/admin/claude-accounts/${id}`);
  },

  /**
   * Toggle Claude account active status
   */
  toggle: async (id: number, isActive: boolean): Promise<void> => {
    await request.patch(`/admin/claude-accounts/${id}/toggle`, { is_active: isActive });
  },

  /**
   * Refresh Claude account token
   */
  refreshToken: async (id: number): Promise<void> => {
    await request.post(`/admin/claude-accounts/${id}/refresh`);
  },
};
