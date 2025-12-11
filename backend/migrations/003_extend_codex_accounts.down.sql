-- 003_extend_codex_accounts.down.sql
-- Rollback extension of codex_accounts table

-- Drop indexes
DROP INDEX IF EXISTS idx_codex_accounts_type;
DROP INDEX IF EXISTS idx_codex_accounts_schedulable;
DROP INDEX IF EXISTS idx_codex_accounts_email;

-- Rename column back
ALTER TABLE codex_accounts
    RENAME COLUMN schedulable TO is_schedulable;

-- Revert api_key column changes
ALTER TABLE codex_accounts
    ALTER COLUMN api_key TYPE VARCHAR(255),
    ALTER COLUMN api_key SET NOT NULL;

-- Drop new columns
ALTER TABLE codex_accounts
    DROP COLUMN IF EXISTS last_used_at,
    DROP COLUMN IF EXISTS proxy_config,
    DROP COLUMN IF EXISTS priority,
    DROP COLUMN IF EXISTS rate_limit_reset_at,
    DROP COLUMN IF EXISTS rate_limit_status,
    DROP COLUMN IF EXISTS quota_reset_time,
    DROP COLUMN IF EXISTS last_reset_date,
    DROP COLUMN IF EXISTS daily_usage,
    DROP COLUMN IF EXISTS daily_quota,
    DROP COLUMN IF EXISTS subscription_expires_at,
    DROP COLUMN IF EXISTS subscription_level,
    DROP COLUMN IF EXISTS custom_user_agent,
    DROP COLUMN IF EXISTS base_api,
    DROP COLUMN IF EXISTS scopes,
    DROP COLUMN IF EXISTS expires_at,
    DROP COLUMN IF EXISTS refresh_token,
    DROP COLUMN IF EXISTS access_token,
    DROP COLUMN IF EXISTS email,
    DROP COLUMN IF EXISTS account_type;
