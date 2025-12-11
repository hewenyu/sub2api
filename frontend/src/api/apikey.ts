import { request } from '@/utils/request';
import type {
  APIKey,
  CreateAPIKeyRequest,
  CreateAPIKeyResponse,
  APIKeyFilters,
  APIKeyListResponse,
} from '@/types/apikey';

export const apikeyApi = {
  /**
   * Create a new API key
   */
  create: async (data: CreateAPIKeyRequest): Promise<CreateAPIKeyResponse> => {
    const response = await request.post<CreateAPIKeyResponse>('/admin/api-keys', data);
    return response.data.data || { api_key: '', api_key_object: {} as APIKey };
  },

  /**
   * List all API keys with optional filters
   */
  list: async (params?: APIKeyFilters): Promise<APIKeyListResponse> => {
    const response = await request.get<APIKeyListResponse>('/admin/api-keys', params);
    return response.data.data || { items: [], pagination: { page: 1, page_size: 20 } };
  },

  /**
   * Get a single API key by ID
   */
  get: async (id: number): Promise<APIKey> => {
    const response = await request.get<APIKey>(`/admin/api-keys/${id}`);
    return response.data.data || ({} as APIKey);
  },

  /**
   * Update an API key
   */
  update: async (id: number, data: Partial<CreateAPIKeyRequest>): Promise<void> => {
    await request.put(`/admin/api-keys/${id}`, data);
  },

  /**
   * Delete an API key
   */
  delete: async (id: number): Promise<void> => {
    await request.delete(`/admin/api-keys/${id}`);
  },

  /**
   * Toggle API key active status
   */
  toggle: async (id: number, isActive: boolean): Promise<void> => {
    await request.patch(`/admin/api-keys/${id}/toggle`, { is_active: isActive });
  },
};
