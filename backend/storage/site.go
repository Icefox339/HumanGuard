// internal/storage/site.go
package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

func (s *storage) CreateSite(ctx context.Context, site *Site) error {
	if site.ID == "" {
		site.ID = generateID()
	}

	now := time.Now()
	site.CreatedAt = now
	site.UpdatedAt = now

	if site.Status == "" {
		site.Status = "verifying"
	}

	var settingsJSON []byte
	if site.Settings != nil {
		settingsJSON, _ = json.Marshal(site.Settings)
	} else {
		settingsJSON = []byte("{}")
	}

	query := `
		INSERT INTO sites (
			id, user_id, name, domain, origin_server,
			status, settings, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9
		)
	`

	_, err := s.db.ExecContext(ctx, query,
		site.ID,
		site.UserID,
		site.Name,
		site.Domain,
		site.OriginServer,
		site.Status,
		settingsJSON,
		site.CreatedAt,
		site.UpdatedAt,
	)

	if err != nil {
		if isUniqueViolation(err) {
			return ErrSiteAlreadyExists
		}
		return fmt.Errorf("failed to create site: %w", err)
	}

	return nil
}

func (s *storage) GetSite(ctx context.Context, id string) (*Site, error) {
	query := `
		SELECT
			id, user_id, name, domain, origin_server,
			status, settings, created_at, updated_at
		FROM sites
		WHERE id = $1
	`

	var site Site
	var settingsJSON []byte

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&site.ID,
		&site.UserID,
		&site.Name,
		&site.Domain,
		&site.OriginServer,
		&site.Status,
		&settingsJSON,
		&site.CreatedAt,
		&site.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSiteNotFound
		}
		return nil, fmt.Errorf("failed to get site: %w", err)
	}

	if len(settingsJSON) > 0 {
		var settings ModuleSettings
		if err := json.Unmarshal(settingsJSON, &settings); err == nil {
			site.Settings = &settings
		}
	}

	return &site, nil
}

func (s *storage) GetSiteByDomain(ctx context.Context, domain string) (*Site, error) {
	query := `
		SELECT
			id, user_id, name, domain, origin_server,
			status, settings, created_at, updated_at
		FROM sites
		WHERE domain = $1
	`

	var site Site
	var settingsJSON []byte

	err := s.db.QueryRowContext(ctx, query, domain).Scan(
		&site.ID,
		&site.UserID,
		&site.Name,
		&site.Domain,
		&site.OriginServer,
		&site.Status,
		&settingsJSON,
		&site.CreatedAt,
		&site.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSiteNotFound
		}
		return nil, fmt.Errorf("failed to get site by domain: %w", err)
	}

	if len(settingsJSON) > 0 {
		var settings ModuleSettings
		if err := json.Unmarshal(settingsJSON, &settings); err == nil {
			site.Settings = &settings
		}
	}

	return &site, nil
}

func (s *storage) UpdateSite(ctx context.Context, site *Site) error {
	site.UpdatedAt = time.Now()

	var settingsJSON []byte
	if site.Settings != nil {
		settingsJSON, _ = json.Marshal(site.Settings)
	}

	query := `
		UPDATE sites
		SET
			name = $1,
			origin_server = $2,
			status = $3,
			settings = COALESCE($4, settings),
			updated_at = $5
		WHERE id = $6
	`

	result, err := s.db.ExecContext(ctx, query,
		site.Name,
		site.OriginServer,
		site.Status,
		settingsJSON,
		site.UpdatedAt,
		site.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update site: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrSiteNotFound
	}

	return nil
}

func (s *storage) DeleteSite(ctx context.Context, id string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `DELETE FROM sites WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete site: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrSiteNotFound
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *storage) UpdateSiteStatus(ctx context.Context, siteID, status string) error {
	query := `
		UPDATE sites
		SET
			status = $1,
			updated_at = $2
		WHERE id = $3
	`

	result, err := s.db.ExecContext(ctx, query, status, time.Now(), siteID)
	if err != nil {
		return fmt.Errorf("failed to update site status: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrSiteNotFound
	}

	return nil
}

func (s *storage) ActivateSite(ctx context.Context, siteID string) error {
	return s.UpdateSiteStatus(ctx, siteID, "active")
}

func (s *storage) SuspendSite(ctx context.Context, siteID string) error {
	return s.UpdateSiteStatus(ctx, siteID, "suspended")
}

func (s *storage) GetSitesByUserID(ctx context.Context, userID string) ([]*Site, error) {
	query := `
		SELECT id, user_id, name, domain, origin_server, status, settings, created_at, updated_at
		FROM sites
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sites by user: %w", err)
	}
	defer rows.Close()

	var sites []*Site
	for rows.Next() {
		var site Site
		var settingsJSON []byte

		err := rows.Scan(
			&site.ID, &site.UserID, &site.Name, &site.Domain, &site.OriginServer,
			&site.Status, &settingsJSON, &site.CreatedAt, &site.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan site: %w", err)
		}

		if len(settingsJSON) > 0 {
			var settings ModuleSettings
			if err := json.Unmarshal(settingsJSON, &settings); err == nil {
				site.Settings = &settings
			}
		}

		sites = append(sites, &site)
	}

	return sites, nil
}
