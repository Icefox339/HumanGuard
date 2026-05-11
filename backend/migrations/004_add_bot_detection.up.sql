ALTER TABLE sessions ADD COLUMN IF NOT EXISTS fingerprint VARCHAR(255);
ALTER TABLE sessions ADD COLUMN IF NOT EXISTS metrics JSONB DEFAULT '{}';

CREATE INDEX IF NOT EXISTS idx_sessions_metrics ON sessions USING gin(metrics);

COMMENT ON COLUMN sessions.metrics IS 'Aggregated behavior metrics: mouse_moves, clicks, scrolls, keystrokes, duration, fingerprint, etc.';