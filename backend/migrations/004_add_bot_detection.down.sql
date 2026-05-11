DROP INDEX IF EXISTS idx_sessions_metrics;
ALTER TABLE sessions DROP COLUMN IF EXISTS metrics;
ALTER TABLE sessions DROP COLUMN IF EXISTS fingerprint;
