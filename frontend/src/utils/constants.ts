export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api/v1';

export const TOKEN_KEY = 'auth_token';

export const ROUTES = {
  HOME: '/',
  LOGIN: '/login',
  DASHBOARD: '/dashboard',
  API_KEYS: '/api-keys',
  CLAUDE_ACCOUNTS: '/claude-accounts',
  CODEX_ACCOUNTS: '/codex-accounts',
  USAGE_STATS: '/usage-stats',
  PROXIES: '/proxies',
  ADMIN: '/admin',
} as const;

export const STATUS = {
  ACTIVE: 1,
  INACTIVE: 0,
} as const;

export const PROVIDERS = {
  ANTHROPIC: 'anthropic',
  OPENAI: 'openai',
  GOOGLE: 'google',
} as const;

export const ROLES = {
  ADMIN: 'admin',
  USER: 'user',
} as const;

export const PAGINATION_DEFAULTS = {
  PAGE: 1,
  LIMIT: 10,
  PAGE_SIZE_OPTIONS: [10, 20, 50, 100],
} as const;
