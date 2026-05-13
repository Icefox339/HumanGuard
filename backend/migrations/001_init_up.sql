CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    avatar_url TEXT,
    role VARCHAR(20) DEFAULT 'user' CHECK (role IN ('user', 'admin')),
    password_hash VARCHAR(255) NOT NULL,
    is_verified BOOLEAN DEFAULT false,
    totp_secret VARCHAR(255),
    oauth_provider VARCHAR(50),
    oauth_id VARCHAR(255),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    last_login TIMESTAMP
);

CREATE TABLE sites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    domain VARCHAR(255) UNIQUE NOT NULL,
    origin_server TEXT NOT NULL,
    status VARCHAR(20) DEFAULT 'verifying' CHECK (status IN ('verifying', 'active', 'suspended')),
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE blacklist (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    ip VARCHAR(45) NOT NULL,
    reason TEXT DEFAULT 'Auto-blocked by high risk score',
    created_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP,
    UNIQUE(site_id, ip)
);

CREATE TABLE files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    original_name VARCHAR(255) NOT NULL,
    size BIGINT NOT NULL,
    mime_type VARCHAR(100),
    hash VARCHAR(64),
    path TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE shares (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    token VARCHAR(64) UNIQUE NOT NULL,
    shared_by UUID NOT NULL REFERENCES users(id),
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
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

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_oauth ON users(oauth_provider, oauth_id);
CREATE INDEX idx_sites_user_id ON sites(user_id);
CREATE INDEX idx_sites_status ON sites(status);
CREATE INDEX idx_blacklist_site_ip ON blacklist(site_id, ip);
CREATE INDEX idx_blacklist_expires ON blacklist(expires_at);
CREATE INDEX idx_api_keys_user_id ON api_keys(user_id);
CREATE INDEX idx_api_keys_key_hash ON api_keys(key_hash);
CREATE INDEX idx_api_keys_expires_at ON api_keys(expires_at);
CREATE INDEX idx_api_keys_revoked ON api_keys(revoked);

UPDATE sites SET settings = '{
    "collector": {
        "enabled": true,
        "mouse_tracking": true,
        "click_tracking": true,
        "scroll_tracking": true,
        "keystroke_tracking": true,
        "fingerprint_enabled": true,
        "sample_rate": 100
    },
    "analyzer": {
        "enabled": true,
        "rate_limiting": true,
        "pattern_analysis": true,
        "headless_detection": true,
        "thresholds": {
            "low": 30,
            "medium": 60,
            "high": 80
        }
    },
    "reaction": {
        "enabled": true,
        "low_risk_action": "allow",
        "medium_risk_action": "captcha",
        "high_risk_action": "block",
        "block_duration_minutes": 60,
        "block_duration_permanent": false,
        "block_message": "Access denied. Your activity appears to be automated.",
        "add_to_blacklist": true,
        "blacklist_duration_minutes": 1440,
        "captcha_provider": "hcaptcha"
    },
    "blacklist": {
        "enabled": true,
        "ips": [],
        "cidrs": [],
        "user_agents": [],
        "auto_block_threshold": 90,
        "auto_block_duration_minutes": 1440
    }
}'::jsonb WHERE settings = '{}'::jsonb;