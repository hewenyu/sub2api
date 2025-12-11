export interface CreateAdminRequest {
  username: string;
  email: string;
  password: string;
  role?: string;
}

export interface UpdateAdminRequest {
  username?: string;
  email?: string;
  password?: string;
  role?: string;
  status?: number;
}

export interface CreateAPIKeyRequest {
  key_name: string;
  api_key: string;
  provider: string;
}

export interface UpdateAPIKeyRequest {
  key_name?: string;
  api_key?: string;
  provider?: string;
  status?: number;
}

export interface CreateCodexAccountRequest {
  email: string;
  password: string;
}

export interface UpdateCodexAccountRequest {
  email?: string;
  password?: string;
  status?: number;
}

export interface UsageStatsFilters {
  api_key_id?: number;
  codex_account_id?: number;
  model?: string;
  start_time?: string;
  end_time?: string;
}
