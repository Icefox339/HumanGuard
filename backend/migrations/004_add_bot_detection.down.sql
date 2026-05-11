DROP INDEX IF EXISTS idx_sessions_fingerprint;
DROP INDEX IF EXISTS idx_be_type;
DROP INDEX IF EXISTS idx_be_recorded;
DROP INDEX IF EXISTS idx_be_session;
ALTER TABLE sessions DROP COLUMN IF EXISTS fingerprint;
DROP TABLE IF EXISTS behavior_events;
