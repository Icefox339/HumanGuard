CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    avatar_url TEXT,
    role VARCHAR(20) DEFAULT 'user' CHECK (role IN ('user', 'admin')),
    password_hash VARCHAR(255) NOT NULL,
    is_verified BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    last_login TIMESTAMP
);

CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL,
    ip VARCHAR(45) NOT NULL,
    user_agent TEXT,
    device VARCHAR(100),
    location VARCHAR(100),
    is_active BOOLEAN DEFAULT true,
    risk_score INTEGER DEFAULT 0 CHECK (risk_score >= 0 AND risk_score <= 100),
    is_blocked BOOLEAN DEFAULT false,
    captcha_shown BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW(),
    last_activity TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL
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

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_sessions_site_id ON sessions(site_id);
CREATE INDEX idx_sessions_ip ON sessions(ip);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);
CREATE INDEX idx_sites_user_id ON sites(user_id);
CREATE INDEX idx_sites_status ON sites(status);

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
        "block_duration": 60,
        "captcha_provider": "hcaptcha"
    }
}'::jsonb WHERE settings = '{}'::jsonb;