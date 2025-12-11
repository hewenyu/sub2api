export interface Admin {
  id: number;
  username: string;
  email: string;
  role: string;
  status: number;
  created_at: string;
  updated_at: string;
}

export interface APIKey {
  id: number;
  key_name: string;
  api_key: string;
  provider: string;
  status: number;
  created_at: string;
  updated_at: string;
}

export interface CodexAccount {
  id: number;
  email: string;
  password: string;
  status: number;
  last_used_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface UsageRecord {
  id: number;
  api_key_id: number;
  codex_account_id: number | null;
  model: string;
  input_tokens: number;
  output_tokens: number;
  total_tokens: number;
  request_time: string;
  created_at: string;
}

export interface PaginationParams {
  page?: number;
  limit?: number;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  limit: number;
  total_pages: number;
}

export interface ApiResponse<T = unknown> {
  code: number;
  message: string;
  data?: T;
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  token: string;
  admin: Admin;
}
