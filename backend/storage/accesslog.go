package storage

import (
	"context"
	"fmt"
	"time"
)

func (s *storage) LogAccess(ctx context.Context, log *AccessLog) error {
	if log.ID == "" {
		log.ID = generateID()
	}
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}
	if log.Action == "" {
		log.Action = "allowed"
	}

	query := `
		INSERT INTO access_logs (id, session_id, site_id, ip, path, method, user_agent, referer, status_code, risk_score, action, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := s.db.ExecContext(ctx, query,
		log.ID, log.SessionID, log.SiteID, log.IP,
		log.Path, log.Method, log.UserAgent, log.Referer,
		log.StatusCode, log.RiskScore, log.Action, log.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to log access: %w", err)
	}

	return nil
}

func (s *storage) GetAccessLogs(ctx context.Context, siteID string) ([]*AccessLog, error) {
	query := `
		SELECT id, session_id, site_id, ip, path, method, user_agent, referer, status_code, risk_score, action, created_at
		FROM access_logs 
		WHERE site_id = $1 
		ORDER BY created_at DESC
		LIMIT 100
	`

	rows, err := s.db.QueryContext(ctx, query, siteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get access logs: %w", err)
	}
	defer rows.Close()

	var logs []*AccessLog
	for rows.Next() {
		var l AccessLog
		err := rows.Scan(&l.ID, &l.SessionID, &l.SiteID, &l.IP, &l.Path, &l.Method,
			&l.UserAgent, &l.Referer, &l.StatusCode, &l.RiskScore, &l.Action, &l.CreatedAt)
		if err != nil {
			return nil, err
		}
		logs = append(logs, &l)
	}

	return logs, nil
}

func (s *storage) GetLogStats(ctx context.Context, siteID string, from, to time.Time) (*LogStats, error) {
	query := `
		SELECT 
			COUNT(*) as total_requests,
			COUNT(CASE WHEN action = 'blocked' THEN 1 END) as blocked_requests,
			COUNT(CASE WHEN action = 'captcha' THEN 1 END) as captcha_shown,
			COUNT(CASE WHEN action = 'allowed' THEN 1 END) as allowed_requests,
			COUNT(DISTINCT ip) as unique_ips,
			COALESCE(AVG(risk_score), 0) as avg_risk_score
		FROM access_logs 
		WHERE site_id = $1 AND created_at >= $2 AND created_at <= $3
	`

	var stats LogStats
	err := s.db.QueryRowContext(ctx, query, siteID, from, to).Scan(
		&stats.TotalRequests, &stats.BlockedRequests, &stats.CaptchaShown,
		&stats.AllowedRequests, &stats.UniqueIPs, &stats.AvgRiskScore,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get log stats: %w", err)
	}

	return &stats, nil
}

func (s *storage) CleanupOldLogs(ctx context.Context, siteID string, before time.Time) (int64, error) {
	query := `DELETE FROM access_logs WHERE site_id = $1 AND created_at < $2`
	result, err := s.db.ExecContext(ctx, query, siteID, before)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old logs: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	return rowsAffected, nil
}