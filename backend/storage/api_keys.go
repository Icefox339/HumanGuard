// backend/storage/api_keys.go
package storage

import (
    "context"
    "database/sql"
    "fmt"
    "time"
)

func (s *storage) CreateAPIKey(ctx context.Context, key *APIKey) error {
    if key.ID == "" {
        key.ID = generateID()
    }
    if key.CreatedAt.IsZero() {
        key.CreatedAt = time.Now()
    }
    
    query := `
        INSERT INTO api_keys (id, user_id, name, key_hash, prefix, 
                              expires_at, created_at, created_by)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `
    
    _, err := s.db.ExecContext(ctx, query,
        key.ID, key.UserID, key.Name, key.KeyHash, key.Prefix,
        key.ExpiresAt, key.CreatedAt, key.CreatedBy,
    )
    
    if err != nil {
        if isUniqueViolation(err) {
            return fmt.Errorf("api key hash already exists")
        }
        return fmt.Errorf("failed to create api key: %w", err)
    }
    
    return nil
}

func (s *storage) GetAPIKeyByHash(ctx context.Context, keyHash string) (*APIKey, error) {
    query := `
        SELECT id, user_id, name, key_hash, prefix, last_used_at, 
               expires_at, created_at, revoked, created_by
        FROM api_keys 
        WHERE key_hash = $1
    `
    
    var key APIKey
    var lastUsedAt, expiresAt sql.NullTime
    var createdBy sql.NullString
    
    err := s.db.QueryRowContext(ctx, query, keyHash).Scan(
        &key.ID, &key.UserID, &key.Name, &key.KeyHash, &key.Prefix,
        &lastUsedAt, &expiresAt, &key.CreatedAt, &key.Revoked,
        &createdBy,
    )
    
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }
        return nil, fmt.Errorf("failed to get api key: %w", err)
    }
    
    if lastUsedAt.Valid {
        key.LastUsedAt = &lastUsedAt.Time
    }
    if expiresAt.Valid {
        key.ExpiresAt = &expiresAt.Time
    }
    if createdBy.Valid {
        key.CreatedBy = &createdBy.String
    }
    
    return &key, nil
}

func (s *storage) GetAPIKeyByID(ctx context.Context, id string) (*APIKey, error) {
    query := `
        SELECT id, user_id, name, key_hash, prefix, last_used_at, 
               expires_at, created_at, revoked, created_by
        FROM api_keys 
        WHERE id = $1
    `
    
    var key APIKey
    var lastUsedAt, expiresAt sql.NullTime
    var createdBy sql.NullString
    
    err := s.db.QueryRowContext(ctx, query, id).Scan(
        &key.ID, &key.UserID, &key.Name, &key.KeyHash, &key.Prefix,
        &lastUsedAt, &expiresAt, &key.CreatedAt, &key.Revoked,
        &createdBy,
    )
    
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }
        return nil, fmt.Errorf("failed to get api key: %w", err)
    }
    
    if lastUsedAt.Valid {
        key.LastUsedAt = &lastUsedAt.Time
    }
    if expiresAt.Valid {
        key.ExpiresAt = &expiresAt.Time
    }
    if createdBy.Valid {
        key.CreatedBy = &createdBy.String
    }
    
    return &key, nil
}

func (s *storage) ListAPIKeys(ctx context.Context, userID string) ([]*APIKey, error) {
    query := `
        SELECT id, user_id, name, key_hash, prefix, last_used_at, 
               expires_at, created_at, revoked, created_by
        FROM api_keys 
        WHERE user_id = $1
        ORDER BY created_at DESC
    `
    
    rows, err := s.db.QueryContext(ctx, query, userID)
    if err != nil {
        return nil, fmt.Errorf("failed to list api keys: %w", err)
    }
    defer rows.Close()
    
    var keys []*APIKey
    for rows.Next() {
        var key APIKey
        var lastUsedAt, expiresAt sql.NullTime
        var createdBy sql.NullString
        
        err := rows.Scan(
            &key.ID, &key.UserID, &key.Name, &key.KeyHash, &key.Prefix,
            &lastUsedAt, &expiresAt, &key.CreatedAt, &key.Revoked,
            &createdBy,
        )
        if err != nil {
            return nil, err
        }
        
        if lastUsedAt.Valid {
            key.LastUsedAt = &lastUsedAt.Time
        }
        if expiresAt.Valid {
            key.ExpiresAt = &expiresAt.Time
        }
        if createdBy.Valid {
            key.CreatedBy = &createdBy.String
        }
        
        keys = append(keys, &key)
    }
    
    return keys, nil
}

func (s *storage) RevokeAPIKey(ctx context.Context, id string) error {
    query := `UPDATE api_keys SET revoked = true WHERE id = $1`
    result, err := s.db.ExecContext(ctx, query, id)
    if err != nil {
        return fmt.Errorf("failed to revoke api key: %w", err)
    }
    
    rows, _ := result.RowsAffected()
    if rows == 0 {
        return fmt.Errorf("api key not found")
    }
    
    return nil
}

func (s *storage) UpdateAPIKeyLastUsed(ctx context.Context, id string) error {
    query := `UPDATE api_keys SET last_used_at = NOW() WHERE id = $1`
    _, err := s.db.ExecContext(ctx, query, id)
    return err
}

func (s *storage) DeleteAPIKey(ctx context.Context, id string) error {
    query := `DELETE FROM api_keys WHERE id = $1`
    result, err := s.db.ExecContext(ctx, query, id)
    if err != nil {
        return fmt.Errorf("failed to delete api key: %w", err)
    }
    
    rows, _ := result.RowsAffected()
    if rows == 0 {
        return fmt.Errorf("api key not found")
    }
    
    return nil
}