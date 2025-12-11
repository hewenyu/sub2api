-- 004_extend_api_keys.up.sql
-- Extend api_keys table to match GORM model with restrictions and cost limits

-- Rename columns to match model naming
ALTER TABLE api_keys RENAME COLUMN key TO key_hash;
ALTER TABLE api_keys RENAME COLUMN concurrent_requests_max TO max_concurrent_requests;

-- Add new columns for key management
ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS key_prefix VARCHAR(20) NOT NULL DEFAULT '';

-- Add bound account support
ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS bound_codex_account_id BIGINT;

-- Add cost limit columns
ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS daily_cost_limit DECIMAL(20,6) NOT NULL DEFAULT 0;
ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS weekly_cost_limit DECIMAL(20,6) NOT NULL DEFAULT 0;
ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS monthly_cost_limit DECIMAL(20,6) NOT NULL DEFAULT 0;
ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS total_cost_limit DECIMAL(20,6) NOT NULL DEFAULT 0;

-- Add model restriction columns
ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS enable_model_restriction BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS restricted_models JSONB DEFAULT '[]';

-- Add client restriction columns
ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS enable_client_restriction BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS allowed_clients JSONB DEFAULT '[]';

-- Update key_hash column to use unique index instead of unique constraint on the column
ALTER TABLE api_keys DROP CONSTRAINT IF EXISTS api_keys_key_key;
CREATE UNIQUE INDEX IF NOT EXISTS idx_api_keys_key_hash ON api_keys(key_hash);

-- Add foreign key constraint for bound_codex_account_id
ALTER TABLE api_keys ADD CONSTRAINT fk_api_keys_bound_codex_account
    FOREIGN KEY (bound_codex_account_id) REFERENCES codex_accounts(id) ON DELETE SET NULL;

-- Add index for bound account lookups
CREATE INDEX IF NOT EXISTS idx_api_keys_bound_account ON api_keys(bound_codex_account_id) WHERE bound_codex_account_id IS NOT NULL;

-- Add index for active keys
CREATE INDEX IF NOT EXISTS idx_api_keys_active ON api_keys(is_active) WHERE is_active = true;

-- Add comments for new columns
COMMENT ON COLUMN api_keys.key_hash IS 'Hashed API key value';
COMMENT ON COLUMN api_keys.key_prefix IS 'Visible prefix of the API key for identification';
COMMENT ON COLUMN api_keys.bound_codex_account_id IS 'Optional bound Codex account ID';
COMMENT ON COLUMN api_keys.daily_cost_limit IS 'Daily cost limit (0 = unlimited)';
COMMENT ON COLUMN api_keys.weekly_cost_limit IS 'Weekly cost limit (0 = unlimited)';
COMMENT ON COLUMN api_keys.monthly_cost_limit IS 'Monthly cost limit (0 = unlimited)';
COMMENT ON COLUMN api_keys.total_cost_limit IS 'Total cost limit (0 = unlimited)';
COMMENT ON COLUMN api_keys.enable_model_restriction IS 'Enable model restriction for this key';
COMMENT ON COLUMN api_keys.restricted_models IS 'List of restricted model names';
COMMENT ON COLUMN api_keys.enable_client_restriction IS 'Enable client restriction for this key';
COMMENT ON COLUMN api_keys.allowed_clients IS 'List of allowed client identifiers';
