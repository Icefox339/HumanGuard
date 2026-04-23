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

CREATE TABLE blacklist (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    ip VARCHAR(45) NOT NULL,
    reason TEXT DEFAULT 'Auto-blocked by high risk score',
    created_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP,
    UNIQUE(site_id, ip)
);

CREATE TABLE access_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID REFERENCES sessions(id) ON DELETE CASCADE,
    site_id UUID REFERENCES sites(id) ON DELETE CASCADE,
    ip VARCHAR(45) NOT NULL,
    path TEXT NOT NULL,
    method VARCHAR(10) NOT NULL,
    user_agent TEXT,
    referer TEXT,
    status_code INTEGER,
    risk_score INTEGER DEFAULT 0,
    action VARCHAR(20) DEFAULT 'allowed',
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_sessions_site_id ON sessions(site_id);
CREATE INDEX idx_sessions_ip ON sessions(ip);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);
CREATE INDEX idx_sites_user_id ON sites(user_id);
CREATE INDEX idx_sites_status ON sites(status);
CREATE INDEX idx_blacklist_site_ip ON blacklist(site_id, ip);
CREATE INDEX idx_blacklist_expires ON blacklist(expires_at);
CREATE INDEX idx_access_logs_session_id ON access_logs(session_id);
CREATE INDEX idx_access_logs_site_id ON access_logs(site_id);
CREATE INDEX idx_access_logs_created_at ON access_logs(created_at);

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