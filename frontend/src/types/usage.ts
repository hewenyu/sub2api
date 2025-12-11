export interface UsageSummary {
  total_requests: number;
  total_cost: number;
  total_input_tokens: number;
  total_output_tokens: number;
  claude_account_count: number;
  codex_account_count: number;
  api_key_count: number;
  daily_usage: DailyUsage[];
}

export interface DailyUsage {
  date: string;
  requests: number;
  cost: number;
  input_tokens: number;
  output_tokens: number;
}

export interface RequestRecord {
  id: number;
  api_key_name: string;
  model: string;
  input_tokens: number;
  output_tokens: number;
  cost: number;
  created_at: string;
}

export interface UsageFilters {
  start_date?: string;
  end_date?: string;
  api_key_id?: number;
  account_id?: number;
  account_type?: 'claude' | 'codex';
  model?: string;
}

export interface UsageByModel {
  model: string;
  cost: number;
  requests: number;
  percentage: number;
}

export interface UsageByKey {
  api_key_id: number;
  api_key_name: string;
  cost: number;
  requests: number;
  percentage: number;
}

export interface UsageRecord {
  id: number;
  timestamp: string;
  api_key_name: string;
  account_email: string;
  model: string;
  input_tokens: number;
  output_tokens: number;
  cost: number;
}

export interface UsageStats {
  total_requests: number;
  total_cost: number;
  total_input_tokens: number;
  total_output_tokens: number;
  avg_cost_per_request: number;
  most_used_model: string;
}

export interface DateRange {
  start_date: string;
  end_date: string;
}
