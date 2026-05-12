ALTER TABLE sessions ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT NOW();

CREATE INDEX IF NOT EXISTS idx_sessions_updated_at ON sessions(updated_at);

UPDATE sessions SET updated_at = created_at WHERE updated_at IS NULL;
