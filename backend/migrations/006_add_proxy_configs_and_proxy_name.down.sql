-- 006_add_proxy_configs_and_proxy_name.down.sql
-- Rollback proxy_configs table and proxy_name column changes

-- Remove proxy_name column from codex_accounts
ALTER TABLE codex_accounts
    DROP CONSTRAINT IF EXISTS fk_codex_accounts_proxy_name;

DROP INDEX IF EXISTS idx_codex_accounts_proxy_name;

ALTER TABLE codex_accounts
    DROP COLUMN IF EXISTS proxy_name;

-- Drop proxy_configs table and related objects
DROP TRIGGER IF EXISTS trigger_update_proxy_configs_updated_at ON proxy_configs;
DROP FUNCTION IF EXISTS update_proxy_configs_updated_at();

DROP INDEX IF EXISTS idx_proxy_configs_deleted_at;
DROP INDEX IF EXISTS idx_proxy_configs_enabled;
DROP INDEX IF EXISTS idx_proxy_configs_is_default;
DROP INDEX IF EXISTS idx_proxy_configs_name;

DROP TABLE IF EXISTS proxy_configs;
