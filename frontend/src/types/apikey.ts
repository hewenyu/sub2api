export interface APIKey extends Record<string, unknown> {
  id: number;
  key_prefix: string;
  name: string;
  is_active: boolean;
  expires_at?: string;
  max_concurrent_requests: number;
  bound_codex_account_id?: number;
  rate_limit_per_minute: number;
  rate_limit_per_hour: number;
  rate_limit_per_day: number;
  daily_cost_limit: number;
  weekly_cost_limit: number;
  monthly_cost_limit: number;
  total_cost_limit: number;
  enable_model_restriction: boolean;
  restricted_models: string[];
  enable_client_restriction: boolean;
  allowed_clients: string[];
  total_requests: number;
  total_tokens: number;
  total_cost: number;
  created_at: number;
  updated_at: number;
}

export interface CreateAPIKeyRequest {
  name: string;
  permissions?: string[];
  max_concurrent_requests: number;
  rate_limit_per_minute: number;
  rate_limit_per_hour: number;
  rate_limit_per_day: number;
  daily_cost_limit: number;
  weekly_cost_limit: number;
  monthly_cost_limit: number;
  total_cost_limit: number;
  enable_model_restriction: boolean;
  restricted_models?: string[];
  enable_client_restriction: boolean;
  allowed_clients?: string[];
}

export interface CreateAPIKeyResponse {
  api_key: string;
  api_key_object: APIKey;
}

export interface APIKeyFilters {
  is_active?: boolean;
  search?: string;
  page?: number;
  page_size?: number;
}

export interface APIKeyListResponse {
  items: APIKey[];
  pagination: {
    page: number;
    page_size: number;
  };
}
