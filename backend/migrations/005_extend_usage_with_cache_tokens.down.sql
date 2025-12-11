-- 005_extend_usage_with_cache_tokens.down.sql
-- Remove cache-related token columns from the usage table.

ALTER TABLE usage
    DROP COLUMN IF EXISTS cache_creation_input_tokens,
    DROP COLUMN IF EXISTS cache_read_input_tokens;

