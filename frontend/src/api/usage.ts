import { request } from '@/utils/request';
import type {
  UsageSummary,
  UsageFilters,
  RequestRecord,
  UsageByModel,
  UsageByKey,
  UsageStats,
} from '@/types/usage';

export const usageApi = {
  /**
   * Get usage statistics summary
   */
  getSummary: async (filters?: UsageFilters): Promise<UsageSummary> => {
    const response = await request.get<UsageSummary>('/usage/summary', filters);
    return (
      response.data.data || {
        total_requests: 0,
        total_cost: 0,
        total_input_tokens: 0,
        total_output_tokens: 0,
        claude_account_count: 0,
        codex_account_count: 0,
        api_key_count: 0,
        daily_usage: [],
      }
    );
  },

  /**
   * Get recent request records
   */
  getRecent: async (limit: number = 10): Promise<RequestRecord[]> => {
    const response = await request.get<RequestRecord[]>('/usage/recent', { limit });
    return response.data.data || [];
  },

  /**
   * Get usage statistics with filters
   */
  getStats: async (filters?: UsageFilters): Promise<UsageStats> => {
    const response = await request.get<UsageStats>('/admin/usage/stats', filters);
    return (
      response.data.data || {
        total_requests: 0,
        total_cost: 0,
        total_input_tokens: 0,
        total_output_tokens: 0,
        avg_cost_per_request: 0,
        most_used_model: '',
      }
    );
  },

  /**
   * Get usage breakdown by model
   */
  getByModel: async (filters?: UsageFilters): Promise<UsageByModel[]> => {
    const response = await request.get<UsageByModel[]>('/admin/usage/by-model', filters);
    return response.data.data || [];
  },

  /**
   * Get usage breakdown by API key
   */
  getByKey: async (filters?: UsageFilters): Promise<UsageByKey[]> => {
    const response = await request.get<UsageByKey[]>('/admin/usage/by-key', filters);
    return response.data.data || [];
  },

  /**
   * Get detailed usage records with pagination
   */
  getRecords: async (
    filters?: UsageFilters,
    page: number = 1,
    pageSize: number = 50
  ): Promise<{ items: RequestRecord[]; total: number }> => {
    const response = await request.get<{ items: RequestRecord[]; total: number }>(
      '/admin/usage/records',
      { ...filters, page, page_size: pageSize }
    );
    return response.data.data || { items: [], total: 0 };
  },

  /**
   * Export usage data as CSV
   */
  exportCSV: async (filters?: UsageFilters): Promise<Blob> => {
    const params = new URLSearchParams({ ...(filters as Record<string, string>), format: 'csv' });
    const response = await fetch('/admin/usage/export?' + params);
    return response.blob();
  },

  /**
   * Export usage data as Excel
   */
  exportExcel: async (filters?: UsageFilters): Promise<Blob> => {
    const params = new URLSearchParams({ ...(filters as Record<string, string>), format: 'xlsx' });
    const response = await fetch('/admin/usage/export?' + params);
    return response.blob();
  },
};
