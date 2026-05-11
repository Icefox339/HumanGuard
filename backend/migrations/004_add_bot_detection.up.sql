CREATE TABLE IF NOT EXISTS behavior_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL,
    event_data JSONB,
    recorded_at TIMESTAMP DEFAULT NOW()
);

ALTER TABLE sessions ADD COLUMN IF NOT EXISTS fingerprint VARCHAR(255);

CREATE INDEX IF NOT EXISTS idx_be_session ON behavior_events(session_id);
CREATE INDEX IF NOT EXISTS idx_be_recorded ON behavior_events(recorded_at);
CREATE INDEX IF NOT EXISTS idx_be_type ON behavior_events(event_type);
CREATE INDEX IF NOT EXISTS idx_sessions_fingerprint ON sessions(fingerprint);

COMMENT ON TABLE behavior_events IS 'Поведенческие события для детекции ботов';
COMMENT ON COLUMN sessions.fingerprint IS 'Browser fingerprint для отслеживания подмены';
