package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

func (s *storage) CreateSession(ctx context.Context, session *Session) error {
	if session.ID == "" {
		session.ID = generateID()
	}

	now := time.Now()
	session.CreatedAt = now
	session.LastActivity = now
	
	if session.ExpiresAt.IsZero() {
		session.ExpiresAt = now.Add(30 * time.Minute)
	}

	query := `
		INSERT INTO sessions (
			id, site_id, ip, user_agent, device, location,
			is_active, risk_score, is_blocked, captcha_shown,
			created_at, last_activity, expires_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10,
			$11, $12, $13
		)
	`

	_, err := s.db.ExecContext(ctx, query,
		session.ID,
		session.SiteID,
		session.IP,
		session.UserAgent,
		session.Device,
		session.Location,
		session.IsActive,
		session.RiskScore,
		session.IsBlocked,
		session.CaptchaShown,
		session.CreatedAt,
		session.LastActivity,
		session.ExpiresAt,
	)

	if err != nil {
		if isUniqueViolation(err) {
			return ErrSessionAlreadyExists
		}
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

func (s *storage) GetSession(ctx context.Context, id string) (*Session, error) {
	query := `
		SELECT 
			id, site_id, ip, user_agent, device, location,
			is_active, risk_score, is_blocked, captcha_shown,
			created_at, last_activity, expires_at
		FROM sessions 
		WHERE id = $1
	`

	var session Session

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&session.ID,
		&session.SiteID,
		&session.IP,
		&session.UserAgent,
		&session.Device,
		&session.Location,
		&session.IsActive,
		&session.RiskScore,
		&session.IsBlocked,
		&session.CaptchaShown,
		&session.CreatedAt,
		&session.LastActivity,
		&session.ExpiresAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return &session, nil
}

func (s *storage) GetSessionByCookie(ctx context.Context, cookie string) (*Session, error) {
	return s.GetSession(ctx, cookie)
}

func (s *storage) UpdateSession(ctx context.Context, session *Session) error {
	session.LastActivity = time.Now()

	query := `
		UPDATE sessions 
		SET 
			ip = $1,
			user_agent = $2,
			device = $3,
			location = $4,
			is_active = $5,
			risk_score = $6,
			is_blocked = $7,
			captcha_shown = $8,
			last_activity = $9,
			expires_at = $10
		WHERE id = $11
	`

	result, err := s.db.ExecContext(ctx, query,
		session.IP,
		session.UserAgent,
		session.Device,
		session.Location,
		session.IsActive,
		session.RiskScore,
		session.IsBlocked,
		session.CaptchaShown,
		session.LastActivity,
		session.ExpiresAt,
		session.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrSessionNotFound
	}

	return nil
}

func (s *storage) UpdateSessionActivity(ctx context.Context, id string) error {
	query := `
		UPDATE sessions 
		SET 
			last_activity = $1,
			expires_at = $2
		WHERE id = $3 AND is_active = true
	`

	now := time.Now()
	expiresAt := now.Add(30 * time.Minute)

	result, err := s.db.ExecContext(ctx, query, now, expiresAt, id)
	if err != nil {
		return fmt.Errorf("failed to update session activity: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrSessionNotFound
	}

	return nil
}

func (s *storage) DeactivateSession(ctx context.Context, id string) error {
	query := `
		UPDATE sessions 
		SET 
			is_active = false,
			expires_at = $1
		WHERE id = $2
	`

	result, err := s.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to deactivate session: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrSessionNotFound
	}

	return nil
}

func (s *storage) GetActiveSessionsBySite(ctx context.Context, siteID string, limit int) ([]*Session, error) {
	query := `
		SELECT 
			id, site_id, ip, user_agent, device, location,
			is_active, risk_score, is_blocked, captcha_shown,
			created_at, last_activity, expires_at
		FROM sessions 
		WHERE site_id = $1 AND is_active = true AND expires_at > NOW()
		ORDER BY last_activity DESC
		LIMIT $2
	`

	rows, err := s.db.QueryContext(ctx, query, siteID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get active sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*Session
	for rows.Next() {
		var session Session
		err := rows.Scan(
			&session.ID,
			&session.SiteID,
			&session.IP,
			&session.UserAgent,
			&session.Device,
			&session.Location,
			&session.IsActive,
			&session.RiskScore,
			&session.IsBlocked,
			&session.CaptchaShown,
			&session.CreatedAt,
			&session.LastActivity,
			&session.ExpiresAt,
		)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, &session)
	}

	return sessions, nil
}

func (s *storage) GetSuspiciousSessions(ctx context.Context, siteID string, minRisk int) ([]*Session, error) {
	query := `
		SELECT 
			id, site_id, ip, user_agent, device, location,
			is_active, risk_score, is_blocked, captcha_shown,
			created_at, last_activity, expires_at
		FROM sessions 
		WHERE site_id = $1 AND risk_score >= $2 AND is_active = true
		ORDER BY risk_score DESC, last_activity DESC
		LIMIT 100
	`

	rows, err := s.db.QueryContext(ctx, query, siteID, minRisk)
	if err != nil {
		return nil, fmt.Errorf("failed to get suspicious sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*Session
	for rows.Next() {
		var session Session
		err := rows.Scan(
			&session.ID,
			&session.SiteID,
			&session.IP,
			&session.UserAgent,
			&session.Device,
			&session.Location,
			&session.IsActive,
			&session.RiskScore,
			&session.IsBlocked,
			&session.CaptchaShown,
			&session.CreatedAt,
			&session.LastActivity,
			&session.ExpiresAt,
		)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, &session)
	}

	return sessions, nil
}

func (s *storage) BlockSession(ctx context.Context, id string) error {
	query := `
		UPDATE sessions 
		SET 
			is_blocked = true,
			is_active = false,
			expires_at = $1
		WHERE id = $2
	`

	result, err := s.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to block session: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrSessionNotFound
	}

	return nil
}

func (s *storage) UnblockSession(ctx context.Context, id string) error {
	query := `
		UPDATE sessions 
		SET 
			is_blocked = false,
			is_active = true,
			expires_at = $1
		WHERE id = $2
	`

	expiresAt := time.Now().Add(30 * time.Minute)
	result, err := s.db.ExecContext(ctx, query, expiresAt, id)
	if err != nil {
		return fmt.Errorf("failed to unblock session: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrSessionNotFound
	}

	return nil
}

func (s *storage) UpdateRiskScore(ctx context.Context, id string, score int) error {
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	query := `
		UPDATE sessions 
		SET risk_score = $1 
		WHERE id = $2
	`

	_, err := s.db.ExecContext(ctx, query, score, id)
	return err
}

func (s *storage) MarkCaptchaShown(ctx context.Context, id string) error {
	query := `
		UPDATE sessions 
		SET captcha_shown = true 
		WHERE id = $1
	`

	_, err := s.db.ExecContext(ctx, query, id)
	return err
}

func (s *storage) CleanupExpiredSessions(ctx context.Context) (int64, error) {
	query := `
		DELETE FROM sessions 
		WHERE expires_at < NOW() OR (is_active = false AND last_activity < NOW() - INTERVAL '1 day')
	`

	result, err := s.db.ExecContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	return rowsAffected, nil
}

func (s *storage) GetSessionStats(ctx context.Context, siteID string) (*SessionStats, error) {
	query := `
		SELECT 
			COUNT(*) as total,
			COUNT(CASE WHEN is_active = true AND expires_at > NOW() THEN 1 END) as active,
			COUNT(CASE WHEN is_blocked = true THEN 1 END) as blocked,
			AVG(risk_score) as avg_risk,
			COUNT(DISTINCT ip) as unique_ips
		FROM sessions 
		WHERE site_id = $1 AND created_at > NOW() - INTERVAL '24 hours'
	`

	var stats SessionStats
	err := s.db.QueryRowContext(ctx, query, siteID).Scan(
		&stats.Total,
		&stats.Active,
		&stats.Blocked,
		&stats.AvgRisk,
		&stats.UniqueIPs,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get session stats: %w", err)
	}

	return &stats, nil
}

type SessionStats struct {
	Total     int64   `json:"total"`
	Active    int64   `json:"active"`
	Blocked   int64   `json:"blocked"`
	AvgRisk   float64 `json:"avg_risk"`
	UniqueIPs int64   `json:"unique_ips"`
}

