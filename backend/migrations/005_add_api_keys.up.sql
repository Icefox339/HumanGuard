CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS api_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    key_hash VARCHAR(64) NOT NULL UNIQUE,
    prefix VARCHAR(20) NOT NULL,
    last_used_at TIMESTAMP,
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    revoked BOOLEAN DEFAULT FALSE,
    created_by UUID REFERENCES users(id),
    permissions JSONB DEFAULT '[]'::jsonb
);

CREATE INDEX idx_api_keys_user_id ON api_keys(user_id);
CREATE INDEX idx_api_keys_key_hash ON api_keys(key_hash);
CREATE INDEX idx_api_keys_expires_at ON api_keys(expires_at);
CREATE INDEX idx_api_keys_revoked ON api_keys(revoked);

COMMENT ON TABLE api_keys IS 'API keys for external application access';
COMMENT ON COLUMN api_keys.prefix IS 'Key prefix for identification (hg_v1_)';
COMMENT ON COLUMN api_keys.permissions IS 'Array of allowed permissions (read, write, admin)';