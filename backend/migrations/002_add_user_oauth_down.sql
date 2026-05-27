-- Восстанавливаем колонки
ALTER TABLE users ADD COLUMN oauth_provider VARCHAR(50), ADD COLUMN oauth_id VARCHAR(255);

UPDATE users u SET 
    oauth_provider = sub.provider,
    oauth_id = sub.oauth_id
FROM (
    SELECT DISTINCT ON (user_id) user_id, provider, oauth_id 
    FROM user_oauth 
    ORDER BY user_id, created_at
) sub
WHERE u.id = sub.user_id;

DROP TABLE user_oauth;