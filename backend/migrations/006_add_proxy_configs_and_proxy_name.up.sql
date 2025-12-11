-- 006_add_proxy_configs_and_proxy_name.up.sql
-- Add proxy_configs table and migrate codex_accounts.proxy_config to proxy_name

-- Create proxy_configs table
CREATE TABLE IF NOT EXISTS proxy_configs (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    protocol VARCHAR(10) NOT NULL,
    host VARCHAR(255) NOT NULL,
    port INT NOT NULL,
    username VARCHAR(255),
    password TEXT,
    is_default BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Create indexes for proxy_configs
CREATE UNIQUE INDEX IF NOT EXISTS idx_proxy_configs_name ON proxy_configs(name) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_proxy_configs_is_default ON proxy_configs(is_default) WHERE is_default = true AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_proxy_configs_enabled ON proxy_configs(enabled) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_proxy_configs_deleted_at ON proxy_configs(deleted_at);

-- Add comments
COMMENT ON TABLE proxy_configs IS 'Centralized proxy configuration for API requests';
COMMENT ON COLUMN proxy_configs.name IS 'Unique proxy configuration name';
COMMENT ON COLUMN proxy_configs.enabled IS 'Whether the proxy is enabled';
COMMENT ON COLUMN proxy_configs.protocol IS 'Proxy protocol: http, https, or socks5';
COMMENT ON COLUMN proxy_configs.host IS 'Proxy host (IP address or domain)';
COMMENT ON COLUMN proxy_configs.port IS 'Proxy port (1-65535)';
COMMENT ON COLUMN proxy_configs.username IS 'Optional proxy authentication username';
COMMENT ON COLUMN proxy_configs.password IS 'Encrypted proxy password (AES-256-CBC)';
COMMENT ON COLUMN proxy_configs.is_default IS 'Whether this is the default proxy';

-- Add trigger for auto-updating updated_at
CREATE OR REPLACE FUNCTION update_proxy_configs_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_proxy_configs_updated_at
BEFORE UPDATE ON proxy_configs
FOR EACH ROW
EXECUTE FUNCTION update_proxy_configs_updated_at();

-- Add proxy_name column to codex_accounts
ALTER TABLE codex_accounts
    ADD COLUMN IF NOT EXISTS proxy_name VARCHAR(100);

-- Add foreign key constraint (ON DELETE SET NULL)
ALTER TABLE codex_accounts
    ADD CONSTRAINT fk_codex_accounts_proxy_name
    FOREIGN KEY (proxy_name)
    REFERENCES proxy_configs(name)
    ON DELETE SET NULL;

-- Add index for proxy_name lookups
CREATE INDEX IF NOT EXISTS idx_codex_accounts_proxy_name ON codex_accounts(proxy_name) WHERE proxy_name IS NOT NULL;

-- Add comment
COMMENT ON COLUMN codex_accounts.proxy_name IS 'Reference to proxy_configs.name for proxy configuration';

-- Note: Existing proxy_config column (TEXT JSON) will be preserved for backward compatibility
-- Migration from proxy_config to proxy_name should be handled at application level
