-- 005_extend_usage_with_cache_tokens.up.sql
-- Add cache-related token columns to the usage table so that
-- prompt caching usage can be stored explicitly, in line with the
-- PRD design for usage_records.

ALTER TABLE usage
    ADD COLUMN IF NOT EXISTS cache_creation_input_tokens BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS cache_read_input_tokens BIGINT NOT NULL DEFAULT 0;

