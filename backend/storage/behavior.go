package storage

import (
    "context"
    "database/sql"
    "encoding/json"
    "time"
)

func (s *storage) UpdateFingerprint(ctx context.Context, id string, fingerprint string) error {
    query := `UPDATE sessions SET fingerprint = $1, updated_at = NOW() WHERE id = $2`
    _, err := s.db.ExecContext(ctx, query, fingerprint, id)
    return err
}

func (s *storage) GetFingerprint(ctx context.Context, id string) (string, error) {
    var fingerprint sql.NullString
    query := `SELECT fingerprint FROM sessions WHERE id = $1`
    err := s.db.QueryRowContext(ctx, query, id).Scan(&fingerprint)
    if err != nil {
        if err == sql.ErrNoRows {
            return "", ErrSessionNotFound
        }
        return "", err
    }
    if fingerprint.Valid {
        return fingerprint.String, nil
    }
    return "", nil
}

func (s *storage) RecordBehaviorEvent(ctx context.Context, event *BehaviorEvent) error {
    if event.ID == "" {
        event.ID = generateID()
    }
    if event.RecordedAt.IsZero() {
        event.RecordedAt = time.Now()
    }

    eventData, err := json.Marshal(event.EventData)
    if err != nil {
        return err
    }

    query := `
        INSERT INTO behavior_events (id, session_id, event_type, event_data, recorded_at)
        VALUES ($1, $2, $3, $4, $5)
    `
    _, err = s.db.ExecContext(ctx, query,
        event.ID, event.SessionID, event.EventType, eventData, event.RecordedAt,
    )
    return err
}