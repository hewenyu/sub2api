-- 001_init.down.sql - Rollback initial database schema

-- Drop triggers
DROP TRIGGER IF EXISTS update_usage_updated_at ON usage;
DROP TRIGGER IF EXISTS update_codex_accounts_updated_at ON codex_accounts;
DROP TRIGGER IF EXISTS update_claude_accounts_updated_at ON claude_accounts;
DROP TRIGGER IF EXISTS update_api_keys_updated_at ON api_keys;
DROP TRIGGER IF EXISTS update_admins_updated_at ON admins;

-- Drop trigger function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables (in reverse order of creation to handle dependencies)
DROP TABLE IF EXISTS usage;
DROP TABLE IF EXISTS codex_accounts;
DROP TABLE IF EXISTS claude_accounts;
DROP TABLE IF EXISTS api_keys;
DROP TABLE IF EXISTS admins;

-- Drop extensions if needed (optional)
-- DROP EXTENSION IF EXISTS "uuid-ossp";
