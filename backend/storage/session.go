package storage

import (
    "context"
    "sync"
    "time"
)

type MemorySessionStore struct {
    sessions sync.Map
}

type ActiveSession struct {
    ID           string                 `json:"id"`
    SiteID       string                 `json:"site_id"`
    IP           string                 `json:"ip"`
    UserAgent    string                 `json:"user_agent"`
    Device       string                 `json:"device"`
    Location     string                 `json:"location"`
    IsActive     bool                   `json:"is_active"`
    RiskScore    int                    `json:"risk_score"`
    IsBlocked    bool                   `json:"is_blocked"`
    CaptchaShown bool                   `json:"captcha_shown"`
    Fingerprint  string                 `json:"fingerprint"`
    Metrics      map[string]interface{} `json:"metrics"`
    CreatedAt    time.Time              `json:"created_at"`
    LastActivity time.Time              `json:"last_activity"`
    ExpiresAt    time.Time              `json:"expires_at"`
}

func NewMemorySessionStore() *MemorySessionStore {
    store := &MemorySessionStore{}
    go store.cleanupLoop()
    return store
}

func (m *MemorySessionStore) cleanupLoop() {
    ticker := time.NewTicker(5 * time.Minute)
    for range ticker.C {
        now := time.Now()
        m.sessions.Range(func(key, value interface{}) bool {
            session := value.(*ActiveSession)
            if now.After(session.ExpiresAt) || !session.IsActive {
                m.sessions.Delete(key)
            }
            return true
        })
    }
}

func (m *MemorySessionStore) CreateSession(ctx context.Context, session *ActiveSession) error {
    now := time.Now()
    session.CreatedAt = now
    session.LastActivity = now
    if session.ExpiresAt.IsZero() {
        session.ExpiresAt = now.Add(30 * time.Minute)
    }
    if session.Metrics == nil {
        session.Metrics = make(map[string]interface{})
    }
    m.sessions.Store(session.ID, session)
    return nil
}

func (m *MemorySessionStore) GetSession(ctx context.Context, id string) (*ActiveSession, error) {
    if val, ok := m.sessions.Load(id); ok {
        return val.(*ActiveSession), nil
    }
    return nil, ErrSessionNotFound
}

func (m *MemorySessionStore) UpdateSessionActivity(ctx context.Context, id string) error {
    if val, ok := m.sessions.Load(id); ok {
        session := val.(*ActiveSession)
        session.LastActivity = time.Now()
        session.ExpiresAt = time.Now().Add(30 * time.Minute)
        return nil
    }
    return ErrSessionNotFound
}

func (m *MemorySessionStore) DeactivateSession(ctx context.Context, id string) error {
    if _, ok := m.sessions.Load(id); ok {
        m.sessions.Delete(id)
        return nil
    }
    return ErrSessionNotFound
}
func (m *MemorySessionStore) BlockSession(ctx context.Context, id string) error {
    if val, ok := m.sessions.Load(id); ok {
        session := val.(*ActiveSession)
        session.IsBlocked = true
        session.IsActive = false
        return nil
    }
    return ErrSessionNotFound
}

func (m *MemorySessionStore) UnblockSession(ctx context.Context, id string) error {
    if val, ok := m.sessions.Load(id); ok {
        session := val.(*ActiveSession)
        session.IsBlocked = false
        session.IsActive = true
        return nil
    }
    return ErrSessionNotFound
}

func (m *MemorySessionStore) UpdateRiskScore(ctx context.Context, id string, score int) error {
    if val, ok := m.sessions.Load(id); ok {
        session := val.(*ActiveSession)
        session.RiskScore = score
        return nil
    }
    return ErrSessionNotFound
}

func (m *MemorySessionStore) MarkCaptchaShown(ctx context.Context, id string) error {
    if val, ok := m.sessions.Load(id); ok {
        session := val.(*ActiveSession)
        session.CaptchaShown = true
        return nil
    }
    return ErrSessionNotFound
}

func (m *MemorySessionStore) UpdateSessionMetrics(ctx context.Context, id string, metrics map[string]interface{}) error {
    if val, ok := m.sessions.Load(id); ok {
        session := val.(*ActiveSession)
        for k, v := range metrics {
            session.Metrics[k] = v
        }
        return nil
    }
    return ErrSessionNotFound
}

func (m *MemorySessionStore) GetSessionMetrics(ctx context.Context, id string) (map[string]interface{}, error) {
    if val, ok := m.sessions.Load(id); ok {
        session := val.(*ActiveSession)
        return session.Metrics, nil
    }
    return nil, ErrSessionNotFound
}

func (m *MemorySessionStore) GetActiveSessionsBySite(ctx context.Context, siteID string, limit int) ([]*ActiveSession, error) {
    var sessions []*ActiveSession
    m.sessions.Range(func(key, value interface{}) bool {
        session := value.(*ActiveSession)
        if session.SiteID == siteID && session.IsActive && time.Now().Before(session.ExpiresAt) {
            sessions = append(sessions, session)
        }
        return len(sessions) < limit
    })
    return sessions, nil
}

func (m *MemorySessionStore) GetSuspiciousSessions(ctx context.Context, siteID string, minRisk int) ([]*ActiveSession, error) {
    var sessions []*ActiveSession
    m.sessions.Range(func(key, value interface{}) bool {
        session := value.(*ActiveSession)
        if session.SiteID == siteID && session.RiskScore >= minRisk && session.IsActive {
            sessions = append(sessions, session)
        }
        return true
    })
    return sessions, nil
}

func (m *MemorySessionStore) GetSessionStats(ctx context.Context, siteID string) (*SessionStats, error) {
    var stats SessionStats
    ips := make(map[string]bool)
    m.sessions.Range(func(key, value interface{}) bool {
        session := value.(*ActiveSession)
        if session.SiteID == siteID {
            stats.Total++
            if session.IsActive && time.Now().Before(session.ExpiresAt) {
                stats.Active++
            }
            if session.IsBlocked {
                stats.Blocked++
            }
            stats.AvgRisk += float64(session.RiskScore)
            ips[session.IP] = true
        }
        return true
    })
    if stats.Total > 0 {
        stats.AvgRisk /= float64(stats.Total)
    }
    stats.UniqueIPs = int64(len(ips))
    return &stats, nil
}

func (m *MemorySessionStore) CleanupExpiredSessions(ctx context.Context) (int64, error) {
    var deleted int64
    now := time.Now()
    m.sessions.Range(func(key, value interface{}) bool {
        session := value.(*ActiveSession)
        if now.After(session.ExpiresAt) || !session.IsActive {
            m.sessions.Delete(key)
            deleted++
        }
        return true
    })
    return deleted, nil
}