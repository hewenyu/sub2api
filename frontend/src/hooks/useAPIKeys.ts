import { useState, useEffect, useCallback } from 'react';
import { apikeyApi } from '@/api/apikey';
import { message } from '@/utils/message';
import type {
  APIKey,
  CreateAPIKeyRequest,
  CreateAPIKeyResponse,
  APIKeyFilters,
} from '@/types/apikey';

interface UseAPIKeysResult {
  apiKeys: APIKey[];
  total: number;
  loading: boolean;
  filters: APIKeyFilters;
  setFilters: (filters: APIKeyFilters) => void;
  createAPIKey: (data: CreateAPIKeyRequest) => Promise<CreateAPIKeyResponse>;
  updateAPIKey: (id: number, data: Partial<CreateAPIKeyRequest>) => Promise<void>;
  deleteAPIKey: (id: number) => Promise<void>;
  toggleAPIKey: (id: number, isActive: boolean) => Promise<void>;
  refetch: () => Promise<void>;
}

export const useAPIKeys = (initialFilters?: APIKeyFilters): UseAPIKeysResult => {
  const [apiKeys, setApiKeys] = useState<APIKey[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [filters, setFilters] = useState<APIKeyFilters>(
    initialFilters || { page: 1, page_size: 20 }
  );

  const fetchAPIKeys = useCallback(async () => {
    try {
      setLoading(true);
      const response = await apikeyApi.list(filters);
      setApiKeys(response.items);
      setTotal(response.items.length);
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to fetch API keys';
      message.error(errorMessage);
      setApiKeys([]);
      setTotal(0);
    } finally {
      setLoading(false);
    }
  }, [filters]);

  useEffect(() => {
    fetchAPIKeys();
  }, [fetchAPIKeys]);

  const createAPIKey = async (data: CreateAPIKeyRequest): Promise<CreateAPIKeyResponse> => {
    try {
      const result = await apikeyApi.create(data);
      message.success('API Key created successfully');
      await fetchAPIKeys();
      return result;
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to create API key';
      message.error(errorMessage);
      throw error;
    }
  };

  const updateAPIKey = async (id: number, data: Partial<CreateAPIKeyRequest>): Promise<void> => {
    try {
      await apikeyApi.update(id, data);
      message.success('API Key updated successfully');
      await fetchAPIKeys();
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to update API key';
      message.error(errorMessage);
      throw error;
    }
  };

  const deleteAPIKey = async (id: number): Promise<void> => {
    try {
      await apikeyApi.delete(id);
      message.success('API Key deleted successfully');
      await fetchAPIKeys();
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to delete API key';
      message.error(errorMessage);
      throw error;
    }
  };

  const toggleAPIKey = async (id: number, isActive: boolean): Promise<void> => {
    try {
      await apikeyApi.toggle(id, isActive);
      message.success(isActive ? 'API Key activated' : 'API Key deactivated');
      await fetchAPIKeys();
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : 'Failed to toggle API key status';
      message.error(errorMessage);
      throw error;
    }
  };

  return {
    apiKeys,
    total,
    loading,
    filters,
    setFilters,
    createAPIKey,
    updateAPIKey,
    deleteAPIKey,
    toggleAPIKey,
    refetch: fetchAPIKeys,
  };
};
