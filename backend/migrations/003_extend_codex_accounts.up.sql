-- 003_extend_codex_accounts.up.sql
-- Extend codex_accounts table to support OAuth and enhanced account management

-- Add new columns for account types and OAuth support
ALTER TABLE codex_accounts
    ADD COLUMN IF NOT EXISTS account_type VARCHAR(50) DEFAULT 'openai-responses',
    ADD COLUMN IF NOT EXISTS email VARCHAR(255),
    ADD COLUMN IF NOT EXISTS access_token TEXT,
    ADD COLUMN IF NOT EXISTS refresh_token TEXT,
    ADD COLUMN IF NOT EXISTS expires_at TIMESTAMP,
    ADD COLUMN IF NOT EXISTS scopes TEXT,
    ADD COLUMN IF NOT EXISTS base_api VARCHAR(255) DEFAULT 'https://api.openai.com/v1',
    ADD COLUMN IF NOT EXISTS custom_user_agent VARCHAR(255),
    ADD COLUMN IF NOT EXISTS subscription_level VARCHAR(50),
    ADD COLUMN IF NOT EXISTS subscription_expires_at TIMESTAMP,
    ADD COLUMN IF NOT EXISTS daily_quota DECIMAL(10,2) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS daily_usage DECIMAL(10,2) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS last_reset_date TIMESTAMP,
    ADD COLUMN IF NOT EXISTS quota_reset_time VARCHAR(10) DEFAULT '00:00',
    ADD COLUMN IF NOT EXISTS rate_limit_status VARCHAR(50),
    ADD COLUMN IF NOT EXISTS rate_limit_reset_at TIMESTAMP,
    ADD COLUMN IF NOT EXISTS chatgpt_account_id VARCHAR(255),
    ADD COLUMN IF NOT EXISTS chatgpt_user_id VARCHAR(255),
    ADD COLUMN IF NOT EXISTS organization_id VARCHAR(255),
    ADD COLUMN IF NOT EXISTS organization_role VARCHAR(100),
    ADD COLUMN IF NOT EXISTS organization_title VARCHAR(255),
    ADD COLUMN IF NOT EXISTS priority INT DEFAULT 100,
    ADD COLUMN IF NOT EXISTS proxy_config TEXT,
    ADD COLUMN IF NOT EXISTS last_used_at TIMESTAMP;

-- Update existing api_key column to be nullable and TEXT type for encrypted values
ALTER TABLE codex_accounts
    ALTER COLUMN api_key DROP NOT NULL,
    ALTER COLUMN api_key TYPE TEXT;

-- Rename columns to match new naming convention
ALTER TABLE codex_accounts
    RENAME COLUMN is_schedulable TO schedulable;

-- Add comments for new columns
COMMENT ON COLUMN codex_accounts.account_type IS 'Account type: openai-oauth or openai-responses';
COMMENT ON COLUMN codex_accounts.email IS 'Email address associated with the account';
COMMENT ON COLUMN codex_accounts.access_token IS 'Encrypted OAuth access token';
COMMENT ON COLUMN codex_accounts.refresh_token IS 'Encrypted OAuth refresh token';
COMMENT ON COLUMN codex_accounts.expires_at IS 'OAuth token expiration timestamp';
COMMENT ON COLUMN codex_accounts.scopes IS 'OAuth scopes granted';
COMMENT ON COLUMN codex_accounts.base_api IS 'Base API URL for requests';
COMMENT ON COLUMN codex_accounts.custom_user_agent IS 'Custom User-Agent header';
COMMENT ON COLUMN codex_accounts.subscription_level IS 'OpenAI subscription level';
COMMENT ON COLUMN codex_accounts.subscription_expires_at IS 'Subscription expiration date';
COMMENT ON COLUMN codex_accounts.daily_quota IS 'Daily usage quota';
COMMENT ON COLUMN codex_accounts.daily_usage IS 'Current daily usage';
COMMENT ON COLUMN codex_accounts.last_reset_date IS 'Last quota reset date';
COMMENT ON COLUMN codex_accounts.quota_reset_time IS 'Daily quota reset time (HH:MM)';
COMMENT ON COLUMN codex_accounts.rate_limit_status IS 'Rate limit status';
COMMENT ON COLUMN codex_accounts.rate_limit_reset_at IS 'Rate limit reset timestamp';
COMMENT ON COLUMN codex_accounts.chatgpt_account_id IS 'OpenAI ChatGPT account identifier from ID token';
COMMENT ON COLUMN codex_accounts.chatgpt_user_id IS 'OpenAI ChatGPT user identifier from ID token';
COMMENT ON COLUMN codex_accounts.organization_id IS 'Default organization ID from OpenAI ID token';
COMMENT ON COLUMN codex_accounts.organization_role IS 'User role in default organization';
COMMENT ON COLUMN codex_accounts.organization_title IS 'User title in default organization';
COMMENT ON COLUMN codex_accounts.priority IS 'Account priority for scheduling';
COMMENT ON COLUMN codex_accounts.proxy_config IS 'JSON proxy configuration';
COMMENT ON COLUMN codex_accounts.last_used_at IS 'Last time account was used';

-- Add index for email lookups (nullable, so no unique constraint)
CREATE INDEX IF NOT EXISTS idx_codex_accounts_email ON codex_accounts(email) WHERE email IS NOT NULL;

-- Add index for schedulable accounts
CREATE INDEX IF NOT EXISTS idx_codex_accounts_schedulable ON codex_accounts(is_active, schedulable, priority DESC) WHERE is_active = true AND schedulable = true;

-- Add index for account type filtering
CREATE INDEX IF NOT EXISTS idx_codex_accounts_type ON codex_accounts(account_type);
