-- 001_init.up.sql - Initial database schema

-- Enable UUID extension if needed
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create admins table
CREATE TABLE IF NOT EXISTS admins (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Create api_keys table
CREATE TABLE IF NOT EXISTS api_keys (
    id BIGSERIAL PRIMARY KEY,
    key VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    expires_at TIMESTAMP,
    concurrent_requests_max INT NOT NULL DEFAULT 5,
    rate_limit_per_minute INT NOT NULL DEFAULT 60,
    rate_limit_per_hour INT NOT NULL DEFAULT 3600,
    rate_limit_per_day INT NOT NULL DEFAULT 86400,
    total_requests BIGINT NOT NULL DEFAULT 0,
    total_tokens BIGINT NOT NULL DEFAULT 0,
    total_cost DECIMAL(20,6) NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Create claude_accounts table
CREATE TABLE IF NOT EXISTS claude_accounts (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    access_token TEXT NOT NULL,
    refresh_token TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    is_schedulable BOOLEAN NOT NULL DEFAULT true,
    concurrent_requests INT NOT NULL DEFAULT 0,
    rate_limited_until TIMESTAMP,
    overload_until TIMESTAMP,
    total_requests BIGINT NOT NULL DEFAULT 0,
    total_tokens BIGINT NOT NULL DEFAULT 0,
    features JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Create codex_accounts table
CREATE TABLE IF NOT EXISTS codex_accounts (
    id BIGSERIAL PRIMARY KEY,
    api_key VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    is_schedulable BOOLEAN NOT NULL DEFAULT true,
    concurrent_requests INT NOT NULL DEFAULT 0,
    rate_limited_until TIMESTAMP,
    overload_until TIMESTAMP,
    total_requests BIGINT NOT NULL DEFAULT 0,
    total_tokens BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Create usage table
CREATE TABLE IF NOT EXISTS usage (
    id BIGSERIAL PRIMARY KEY,
    api_key_id BIGINT NOT NULL,
    type VARCHAR(50) NOT NULL,
    account_id BIGINT NOT NULL,
    model VARCHAR(100) NOT NULL,
    input_tokens BIGINT NOT NULL DEFAULT 0,
    output_tokens BIGINT NOT NULL DEFAULT 0,
    total_tokens BIGINT NOT NULL DEFAULT 0,
    cost DECIMAL(20,6) NOT NULL DEFAULT 0,
    request_duration BIGINT NOT NULL DEFAULT 0,
    status_code INT NOT NULL,
    error_message TEXT,
    request_metadata JSONB,
    response_metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Create trigger function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for all tables
CREATE TRIGGER update_admins_updated_at
    BEFORE UPDATE ON admins
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_api_keys_updated_at
    BEFORE UPDATE ON api_keys
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_claude_accounts_updated_at
    BEFORE UPDATE ON claude_accounts
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_codex_accounts_updated_at
    BEFORE UPDATE ON codex_accounts
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_usage_updated_at
    BEFORE UPDATE ON usage
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Create comments for tables
COMMENT ON TABLE admins IS 'Administrator users with access to the management interface';
COMMENT ON TABLE api_keys IS 'API keys for accessing the relay service';
COMMENT ON TABLE claude_accounts IS 'Claude API accounts with OAuth tokens';
COMMENT ON TABLE codex_accounts IS 'Codex API accounts';
COMMENT ON TABLE usage IS 'Usage records for API requests';
