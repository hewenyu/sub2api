-- 002_add_indexes.up.sql - Create indexes for better query performance

-- Admins table indexes
CREATE INDEX IF NOT EXISTS idx_admins_username ON admins(username) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_admins_email ON admins(email) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_admins_is_active ON admins(is_active) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_admins_deleted_at ON admins(deleted_at);

-- API Keys table indexes
CREATE INDEX IF NOT EXISTS idx_api_keys_key ON api_keys(key) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_api_keys_is_active ON api_keys(is_active) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_api_keys_expires_at ON api_keys(expires_at) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_api_keys_deleted_at ON api_keys(deleted_at);

-- Claude Accounts table indexes
CREATE INDEX IF NOT EXISTS idx_claude_accounts_email ON claude_accounts(email) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_claude_accounts_is_active ON claude_accounts(is_active) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_claude_accounts_is_schedulable ON claude_accounts(is_schedulable) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_claude_accounts_expires_at ON claude_accounts(expires_at) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_claude_accounts_rate_limited_until ON claude_accounts(rate_limited_until) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_claude_accounts_overload_until ON claude_accounts(overload_until) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_claude_accounts_deleted_at ON claude_accounts(deleted_at);

-- Composite index for GetSchedulable query optimization
CREATE INDEX IF NOT EXISTS idx_claude_accounts_schedulable ON claude_accounts(is_active, is_schedulable, expires_at, rate_limited_until, overload_until)
    WHERE deleted_at IS NULL AND is_active = true AND is_schedulable = true;

-- Codex Accounts table indexes
CREATE INDEX IF NOT EXISTS idx_codex_accounts_api_key ON codex_accounts(api_key) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_codex_accounts_is_active ON codex_accounts(is_active) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_codex_accounts_is_schedulable ON codex_accounts(is_schedulable) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_codex_accounts_rate_limited_until ON codex_accounts(rate_limited_until) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_codex_accounts_overload_until ON codex_accounts(overload_until) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_codex_accounts_deleted_at ON codex_accounts(deleted_at);

-- Composite index for GetSchedulable query optimization
CREATE INDEX IF NOT EXISTS idx_codex_accounts_schedulable ON codex_accounts(is_active, is_schedulable, rate_limited_until, overload_until)
    WHERE deleted_at IS NULL AND is_active = true AND is_schedulable = true;

-- Usage table indexes
CREATE INDEX IF NOT EXISTS idx_usage_api_key_id ON usage(api_key_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_usage_type ON usage(type) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_usage_account_id ON usage(account_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_usage_model ON usage(model) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_usage_status_code ON usage(status_code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_usage_created_at ON usage(created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_usage_deleted_at ON usage(deleted_at);

-- Composite index for usage queries (api_key_id, type, created_at)
CREATE INDEX IF NOT EXISTS idx_usage_composite ON usage(api_key_id, type, created_at DESC) WHERE deleted_at IS NULL;

-- Index for aggregate queries
CREATE INDEX IF NOT EXISTS idx_usage_aggregate ON usage(api_key_id, created_at) WHERE deleted_at IS NULL;

-- Add comments
COMMENT ON INDEX idx_claude_accounts_schedulable IS 'Optimizes GetSchedulable queries for Claude accounts';
COMMENT ON INDEX idx_codex_accounts_schedulable IS 'Optimizes GetSchedulable queries for Codex accounts';
COMMENT ON INDEX idx_usage_composite IS 'Optimizes usage queries by API key, type, and time range';
COMMENT ON INDEX idx_usage_aggregate IS 'Optimizes aggregate statistics queries';
