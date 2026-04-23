package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)


func (s *storage) AddToBlacklist(ctx context.Context, entry *BlacklistEntry) error {
	if entry.ID == "" {
		entry.ID = generateID()
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}

	query := `
		INSERT INTO blacklist (id, site_id, ip, reason, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (site_id, ip) 
		DO UPDATE SET reason = EXCLUDED.reason, expires_at = EXCLUDED.expires_at
	`

	_, err := s.db.ExecContext(ctx, query,
		entry.ID, entry.SiteID, entry.IP, entry.Reason,
		entry.CreatedAt, entry.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("failed to add to blacklist: %w", err)
	}

	return nil
}

func (s *storage) RemoveFromBlacklist(ctx context.Context, siteID, ip string) error {
	query := `DELETE FROM blacklist WHERE site_id = $1 AND ip = $2`

	result, err := s.db.ExecContext(ctx, query, siteID, ip)
	if err != nil {
		return fmt.Errorf("failed to remove from blacklist: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrBlacklistEntryNotFound
	}

	return nil
}

func (s *storage) IsBlacklisted(ctx context.Context, siteID, ip string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM blacklist 
			WHERE site_id = $1 AND ip = $2 
			AND (expires_at IS NULL OR expires_at > NOW())
		)
	`

	var exists bool
	err := s.db.QueryRowContext(ctx, query, siteID, ip).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check blacklist: %w", err)
	}

	return exists, nil
}

func (s *storage) ListBlacklist(ctx context.Context, siteID string) ([]*BlacklistEntry, error) {
	query := `
		SELECT id, site_id, ip, reason, created_at, expires_at
		FROM blacklist 
		WHERE site_id = $1 
		AND (expires_at IS NULL OR expires_at > NOW())
		ORDER BY created_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query, siteID)
	if err != nil {
		return nil, fmt.Errorf("failed to list blacklist: %w", err)
	}
	defer rows.Close()

	var entries []*BlacklistEntry
	for rows.Next() {
		var entry BlacklistEntry
		var expiresAt sql.NullTime

		err := rows.Scan(&entry.ID, &entry.SiteID, &entry.IP, &entry.Reason, &entry.CreatedAt, &expiresAt)
		if err != nil {
			return nil, err
		}
		if expiresAt.Valid {
			entry.ExpiresAt = &expiresAt.Time
		}
		entries = append(entries, &entry)
	}

	return entries, nil
}

