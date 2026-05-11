// backend/storage/behavior.go
package storage

import (
	"fmt"
	"errors"
    "context"
    "database/sql"
    "encoding/json"
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

func (s *storage) UpdateSessionMetrics(ctx context.Context, sessionID string, metrics map[string]interface{}) error {
    if len(metrics) == 0 {
        return nil
    }
    
    metricsJSON, err := json.Marshal(metrics)
    if err != nil {
        return fmt.Errorf("failed to marshal metrics: %w", err)
    }
    
    query := `
        UPDATE sessions 
        SET metrics = COALESCE(metrics, '{}'::jsonb) || $1::jsonb,
            updated_at = NOW()
        WHERE id = $2
    `
    
    result, err := s.db.ExecContext(ctx, query, string(metricsJSON), sessionID)
    if err != nil {
        return fmt.Errorf("failed to update session metrics: %w", err)
    }
    
    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        return ErrSessionNotFound
    }
    
    return nil
}

func (s *storage) GetSessionMetrics(ctx context.Context, sessionID string) (map[string]interface{}, error) {
    var metricsJSON []byte
    query := `SELECT COALESCE(metrics, '{}'::jsonb) FROM sessions WHERE id = $1`
    
    err := s.db.QueryRowContext(ctx, query, sessionID).Scan(&metricsJSON)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, ErrSessionNotFound
        }
        return nil, fmt.Errorf("failed to get session metrics: %w", err)
    }
    
    var metrics map[string]interface{}
    if len(metricsJSON) > 0 {
        if err := json.Unmarshal(metricsJSON, &metrics); err != nil {
            return make(map[string]interface{}), nil
        }
    }
    
    if metrics == nil {
        return make(map[string]interface{}), nil
    }
    return metrics, nil
}
