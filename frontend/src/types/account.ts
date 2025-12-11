export interface ProxyConfig {
  protocol: 'http' | 'https' | 'socks5';
  host: string;
  port: number;
  username?: string;
  password?: string;
}

export interface ClaudeAccount extends Record<string, unknown> {
  id: number;
  email: string;
  access_token: string;
  refresh_token: string;
  expires_at: string;
  subscription_level: 'free' | 'pro' | 'max';
  has_claude_pro: boolean;
  has_claude_max: boolean;
  is_active: boolean;
  schedulable: boolean;
  priority: number;
  last_used_at: string | null;
  rate_limited_until: string | null;
  overload_until: string | null;
  proxy_url?: string;
  created_at: string;
  updated_at: string;
}

export interface CodexAccount extends Record<string, unknown> {
  id: number;
  name: string;
  account_type: 'openai-oauth' | 'openai-responses';
  email?: string;
  chatgpt_account_id?: string;
  chatgpt_user_id?: string;
  organization_id?: string;
  organization_role?: string;
  organization_title?: string;
  access_token?: string;
  refresh_token?: string;
  expires_at?: string;
  session_token?: string;
  api_key?: string;
  base_api?: string;
  custom_user_agent?: string;
  is_active: boolean;
  schedulable: boolean;
  priority: number;
  last_used_at: string | null;
  rate_limited_until: string | null;
  proxy_url?: string;
  proxy_name?: string;
  created_at: string;
  updated_at: string;
}

export interface AccountStatus {
  claude_accounts: ClaudeAccount[];
  codex_accounts: CodexAccount[];
  claude_healthy: number;
  codex_healthy: number;
  claude_total: number;
  codex_total: number;
}

export interface ClaudeOAuthResponse {
  auth_url: string;
  callback_url: string;
  state: string;
}

export interface CreateClaudeAccountData {
  priority: number;
  schedulable: boolean;
  proxy_url?: string;
}

export interface VerifyClaudeAuthRequest {
  code: string;
  state: string;
  account_data: CreateClaudeAccountData;
}

export interface CreateCodexAccountRequest {
  name: string;
  account_type: 'openai-oauth' | 'openai-responses';

  // OAuth-specific (only for openai-oauth)
  email?: string;

  // OpenAI-Responses-specific (required for openai-responses)
  api_key?: string;
  base_api?: string;
  custom_user_agent?: string;
  daily_quota?: number;
  quota_reset_time?: string;

  // Common fields
  priority: number;
  schedulable: boolean;
  proxy_config?: ProxyConfig;
  proxy_name?: string;
}

export interface UpdateClaudeAccountRequest {
  priority?: number;
  schedulable?: boolean;
  proxy_url?: string;
}

export interface UpdateCodexAccountRequest {
  name?: string;
  base_api?: string;
  custom_user_agent?: string;
  daily_quota?: number;
  quota_reset_time?: string;
  priority?: number;
  schedulable?: boolean;
  proxy_config?: ProxyConfig;
  // proxy_url is kept for backward compatibility with existing UI;
  // it is ignored by the backend and only used on the client side.
  proxy_url?: string;
  proxy_name?: string;
}

export interface AccountListResponse<T> {
  items: T[];
  total: number;
  page?: number;
  page_size?: number;
}

export interface AccountFilters {
  page?: number;
  page_size?: number;
  is_active?: boolean;
  search?: string;
}
