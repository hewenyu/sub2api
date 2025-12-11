export interface ProxyConfig {
  id: number;
  name: string;
  enabled: boolean;
  protocol: string;
  host: string;
  port: number;
  username?: string;
  has_password: boolean;
  is_default: boolean;
  created_at: number;
  updated_at: number;
}

export interface CreateProxyRequest {
  name: string;
  enabled: boolean;
  protocol: string;
  host: string;
  port: number;
  username?: string;
  password?: string;
}

export interface UpdateProxyRequest {
  name?: string;
  enabled?: boolean;
  protocol?: string;
  host?: string;
  port?: number;
  username?: string;
  password?: string;
}

export interface TestProxyResult {
  success: boolean;
  message: string;
  ip?: string;
  country?: string;
  region?: string;
  city?: string;
  isp?: string;
  response_ms?: number;
  error?: string;
  geo_provider?: string;
}

export interface ProxyListResponse {
  items: ProxyConfig[];
  pagination: {
    page: number;
    page_size: number;
    total: number;
    total_page: number;
  };
}

export interface ProxyNameInfo {
  name: string;
  enabled: boolean;
  protocol: string;
  is_default: boolean;
}
