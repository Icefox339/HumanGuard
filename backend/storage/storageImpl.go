package storage

import (
    "context"
    "database/sql"
    "fmt"

    _ "github.com/lib/pq"
)

type storage struct {
    db           *sql.DB
    sessionStore *MemorySessionStore
}

func NewStorage(cfg *Config) (Storage, error) {
    db, err := sql.Open("postgres", cfg.DBURL)
    if err != nil {
        return nil, fmt.Errorf("failed to open database: %w", err)
    }
    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }
    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(5)

    s := &storage{db: db}
    s.sessionStore = NewMemorySessionStore()
    return s, nil
}

func (s *storage) Close() error {
    return s.db.Close()
}

func (s *storage) Ping() error {
    return s.db.Ping()
}

func (s *storage) CreateSession(ctx context.Context, session *ActiveSession) error {
    return s.sessionStore.CreateSession(ctx, session)
}

func (s *storage) GetSession(ctx context.Context, id string) (*ActiveSession, error) {
    return s.sessionStore.GetSession(ctx, id)
}

func (s *storage) UpdateSessionActivity(ctx context.Context, id string) error {
    return s.sessionStore.UpdateSessionActivity(ctx, id)
}

func (s *storage) DeactivateSession(ctx context.Context, id string) error {
    return s.sessionStore.DeactivateSession(ctx, id)
}

func (s *storage) BlockSession(ctx context.Context, id string) error {
    return s.sessionStore.BlockSession(ctx, id)
}

func (s *storage) UnblockSession(ctx context.Context, id string) error {
    return s.sessionStore.UnblockSession(ctx, id)
}

func (s *storage) UpdateRiskScore(ctx context.Context, id string, score int) error {
    return s.sessionStore.UpdateRiskScore(ctx, id, score)
}

func (s *storage) MarkCaptchaShown(ctx context.Context, id string) error {
    return s.sessionStore.MarkCaptchaShown(ctx, id)
}

func (s *storage) UpdateSessionMetrics(ctx context.Context, id string, metrics map[string]interface{}) error {
    return s.sessionStore.UpdateSessionMetrics(ctx, id, metrics)
}

func (s *storage) GetSessionMetrics(ctx context.Context, id string) (map[string]interface{}, error) {
    return s.sessionStore.GetSessionMetrics(ctx, id)
}

func (s *storage) GetActiveSessionsBySite(ctx context.Context, siteID string, limit int) ([]*ActiveSession, error) {
    return s.sessionStore.GetActiveSessionsBySite(ctx, siteID, limit)
}

func (s *storage) GetSuspiciousSessions(ctx context.Context, siteID string, minRisk int) ([]*ActiveSession, error) {
    return s.sessionStore.GetSuspiciousSessions(ctx, siteID, minRisk)
}

func (s *storage) GetSessionStats(ctx context.Context, siteID string) (*SessionStats, error) {
    return s.sessionStore.GetSessionStats(ctx, siteID)
}

func (s *storage) CleanupExpiredSessions(ctx context.Context) (int64, error) {
    return s.sessionStore.CleanupExpiredSessions(ctx)
}