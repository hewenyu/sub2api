-- 002_add_indexes.down.sql - Drop all indexes

-- Drop Usage table indexes
DROP INDEX IF EXISTS idx_usage_aggregate;
DROP INDEX IF EXISTS idx_usage_composite;
DROP INDEX IF EXISTS idx_usage_deleted_at;
DROP INDEX IF EXISTS idx_usage_created_at;
DROP INDEX IF EXISTS idx_usage_status_code;
DROP INDEX IF EXISTS idx_usage_model;
DROP INDEX IF EXISTS idx_usage_account_id;
DROP INDEX IF EXISTS idx_usage_type;
DROP INDEX IF EXISTS idx_usage_api_key_id;

-- Drop Codex Accounts table indexes
DROP INDEX IF EXISTS idx_codex_accounts_schedulable;
DROP INDEX IF EXISTS idx_codex_accounts_deleted_at;
DROP INDEX IF EXISTS idx_codex_accounts_overload_until;
DROP INDEX IF EXISTS idx_codex_accounts_rate_limited_until;
DROP INDEX IF EXISTS idx_codex_accounts_is_schedulable;
DROP INDEX IF EXISTS idx_codex_accounts_is_active;
DROP INDEX IF EXISTS idx_codex_accounts_api_key;

-- Drop Claude Accounts table indexes
DROP INDEX IF EXISTS idx_claude_accounts_schedulable;
DROP INDEX IF EXISTS idx_claude_accounts_deleted_at;
DROP INDEX IF EXISTS idx_claude_accounts_overload_until;
DROP INDEX IF EXISTS idx_claude_accounts_rate_limited_until;
DROP INDEX IF EXISTS idx_claude_accounts_expires_at;
DROP INDEX IF EXISTS idx_claude_accounts_is_schedulable;
DROP INDEX IF EXISTS idx_claude_accounts_is_active;
DROP INDEX IF EXISTS idx_claude_accounts_email;

-- Drop API Keys table indexes
DROP INDEX IF EXISTS idx_api_keys_deleted_at;
DROP INDEX IF EXISTS idx_api_keys_expires_at;
DROP INDEX IF EXISTS idx_api_keys_is_active;
DROP INDEX IF EXISTS idx_api_keys_key;

-- Drop Admins table indexes
DROP INDEX IF EXISTS idx_admins_deleted_at;
DROP INDEX IF EXISTS idx_admins_is_active;
DROP INDEX IF EXISTS idx_admins_email;
DROP INDEX IF EXISTS idx_admins_username;
