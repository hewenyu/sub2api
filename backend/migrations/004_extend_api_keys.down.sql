-- 004_extend_api_keys.down.sql
-- Rollback api_keys table extensions

-- Drop indexes
DROP INDEX IF EXISTS idx_api_keys_active;
DROP INDEX IF EXISTS idx_api_keys_bound_account;
DROP INDEX IF EXISTS idx_api_keys_key_hash;

-- Drop foreign key constraint
ALTER TABLE api_keys DROP CONSTRAINT IF EXISTS fk_api_keys_bound_codex_account;

-- Drop new columns
ALTER TABLE api_keys DROP COLUMN IF EXISTS allowed_clients;
ALTER TABLE api_keys DROP COLUMN IF EXISTS enable_client_restriction;
ALTER TABLE api_keys DROP COLUMN IF EXISTS restricted_models;
ALTER TABLE api_keys DROP COLUMN IF EXISTS enable_model_restriction;
ALTER TABLE api_keys DROP COLUMN IF EXISTS total_cost_limit;
ALTER TABLE api_keys DROP COLUMN IF EXISTS monthly_cost_limit;
ALTER TABLE api_keys DROP COLUMN IF EXISTS weekly_cost_limit;
ALTER TABLE api_keys DROP COLUMN IF EXISTS daily_cost_limit;
ALTER TABLE api_keys DROP COLUMN IF EXISTS bound_codex_account_id;
ALTER TABLE api_keys DROP COLUMN IF EXISTS key_prefix;

-- Rename columns back to original names
ALTER TABLE api_keys RENAME COLUMN max_concurrent_requests TO concurrent_requests_max;
ALTER TABLE api_keys RENAME COLUMN key_hash TO key;

-- Restore original unique constraint on key column
ALTER TABLE api_keys ADD CONSTRAINT api_keys_key_key UNIQUE (key);
