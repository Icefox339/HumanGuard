CREATE TABLE user_oauth (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,
    oauth_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(provider, oauth_id)
);

CREATE INDEX idx_user_oauth_user_id ON user_oauth(user_id);
CREATE INDEX idx_user_oauth_provider_oauth ON user_oauth(provider, oauth_id);

INSERT INTO user_oauth (user_id, provider, oauth_id)
SELECT id, oauth_provider, oauth_id 
FROM users 
WHERE oauth_provider IS NOT NULL AND oauth_id IS NOT NULL;

ALTER TABLE users DROP COLUMN oauth_provider, DROP COLUMN oauth_id;