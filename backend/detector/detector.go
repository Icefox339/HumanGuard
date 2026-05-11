package detector

import (
    "context"
    "database/sql"
    "fmt"
    "strings"
    "sync"
    "time"

    "humanguard/storage"
)

type Detector struct {
    store storage.Storage
    cache sync.Map
    ttl   time.Duration
}

type cachedRisk struct {
    score int
    ts    time.Time
}

type PrefilterResult struct {
    Risk        int
    ShouldBlock bool
    NeedDeep    bool
    Reason      string
}

func New(store storage.Storage) *Detector {
    return &Detector{
        store: store,
        ttl:   10 * time.Second,
    }
}

func (d *Detector) Prefilter(ctx context.Context, sessionID string) (*PrefilterResult, error) {
    if cached, ok := d.cache.Load(sessionID); ok {
        if cr, ok := cached.(cachedRisk); ok && time.Since(cr.ts) < d.ttl {
            return &PrefilterResult{
                Risk:        cr.score,
                ShouldBlock: cr.score >= 80,
                NeedDeep:    false,
                Reason:      "cached",
            }, nil
        }
    }

    session, err := d.store.GetSession(ctx, sessionID)
    if err != nil {
        return nil, fmt.Errorf("get session failed: %w", err)
    }
    if session == nil || session.SiteID == nil {
        return &PrefilterResult{Risk: 0, ShouldBlock: false, NeedDeep: false, Reason: "no_session"}, nil
    }

    risk := 0

    if d.isIPBlacklisted(ctx, session) {
        return &PrefilterResult{
            Risk:        100,
            ShouldBlock: true,
            NeedDeep:    false,
            Reason:      "ip_blacklisted",
        }, nil
    }

    settings, err := d.store.GetSiteSettings(ctx, *session.SiteID)
    if err != nil {
        settings = d.defaultSettings()
    }

    if settings.Analyzer.HeadlessDetection {
        risk += d.headlessScore(session.UserAgent)
    }

    if settings.Analyzer.RateLimiting {
        risk += d.rateLimitScore(ctx, session.IP)
    }

    if risk >= 70 {
        return &PrefilterResult{
            Risk:        risk,
            ShouldBlock: risk >= 80,
            NeedDeep:    false,
            Reason:      "high_risk_from_prefilter",
        }, nil
    }

    return &PrefilterResult{
        Risk:        risk,
        ShouldBlock: false,
        NeedDeep:    true,
        Reason:      "need_deep_analysis",
    }, nil
}

func (d *Detector) DeepFilter(ctx context.Context, sessionID string, currentRisk int) (int, error) {
    risk := currentRisk

    behaviorRisk := d.behaviorScore(ctx, sessionID)
    risk += behaviorRisk

    fpRisk := d.fingerprintAnomaly(ctx, sessionID, "")
    risk += fpRisk

    if risk > 100 {
        risk = 100
    }
    if risk < 0 {
        risk = 0
    }

    return risk, nil
}

func (d *Detector) AnalyzeAndUpdate(ctx context.Context, sessionID string) error {
    prefilter, err := d.Prefilter(ctx, sessionID)
    if err != nil {
        return err
    }

    if prefilter.Reason == "cached" {
        return nil
    }

    finalRisk := prefilter.Risk

    if prefilter.NeedDeep {
        deepRisk, err := d.DeepFilter(ctx, sessionID, finalRisk)
        if err != nil {
            return err
        }
        finalRisk = deepRisk
    }

    if err := d.store.UpdateRiskScore(ctx, sessionID, finalRisk); err != nil {
        return fmt.Errorf("update risk score failed: %w", err)
    }

    d.cache.Store(sessionID, cachedRisk{score: finalRisk, ts: time.Now()})
    return nil
}

func (d *Detector) isIPBlacklisted(ctx context.Context, session *storage.Session) bool {
    if session.SiteID == nil {
        return false
    }
    blacklisted, err := d.store.IsBlacklisted(ctx, *session.SiteID, session.IP)
    return err == nil && blacklisted
}

func (d *Detector) headlessScore(userAgent string) int {
    score := 0
    ua := strings.ToLower(userAgent)

    headlessIndicators := []string{
        "headless", "headlesschrome", "phantomjs",
        "selenium", "webdriver", "puppeteer",
        "playwright", "cypress",
    }

    for _, ind := range headlessIndicators {
        if strings.Contains(ua, ind) {
            score += 25
        }
    }

    if len(ua) < 20 {
        score += 20
    }

    if !strings.Contains(ua, "windows") && !strings.Contains(ua, "mac") &&
        !strings.Contains(ua, "linux") && !strings.Contains(ua, "android") &&
        !strings.Contains(ua, "iphone") && !strings.Contains(ua, "ipad") {
        score += 15
    }

    if score > 100 {
        score = 100
    }
    return score
}

func (d *Detector) rateLimitScore(ctx context.Context, ip string) int {
    dbProvider, ok := d.store.(interface{ GetDB() *sql.DB })
    if !ok {
        return 0
    }
    db := dbProvider.GetDB()
    if db == nil {
        return 0
    }

    var count int
    query := `SELECT COUNT(*) FROM access_logs WHERE ip = $1 AND created_at > NOW() - interval '60 seconds'`
    err := db.QueryRowContext(ctx, query, ip).Scan(&count)
    if err != nil {
        return 0
    }

    const limit = 60
    if count > limit*5 {
        return 100
    }
    if count > limit {
        excess := float64(count-limit) / float64(limit)
        add := int(excess * 25)
        if add > 50 {
            add = 50
        }
        return add
    }
    return 0
}

func (d *Detector) behaviorScore(ctx context.Context, sessionID string) int {
    dbProvider, ok := d.store.(interface{ GetDB() *sql.DB })
    if !ok {
        return 0
    }
    db := dbProvider.GetDB()
    if db == nil {
        return 0
    }

    var eventCount int
    var avgGap float64

    query := `
        SELECT 
            COUNT(*) as event_count,
            EXTRACT(EPOCH FROM (MAX(recorded_at) - MIN(recorded_at))) / NULLIF(COUNT(*) - 1, 0) as avg_gap
        FROM behavior_events 
        WHERE session_id = $1 AND recorded_at > NOW() - interval '5 minutes'
    `
    err := db.QueryRowContext(ctx, query, sessionID).Scan(&eventCount, &avgGap)
    if err != nil {
        return 0
    }

    if eventCount == 0 {
        return 50
    }

    risk := 0
    if avgGap > 0 && avgGap < 0.5 && eventCount > 10 {
        risk += 40
    }
    if eventCount > 500 {
        risk += 30
    }

    var hasMouse, hasClick bool
    mouseQuery := `SELECT EXISTS(SELECT 1 FROM behavior_events WHERE session_id = $1 AND event_type IN ('mousemove', 'mousedown', 'mouseup'))`
    clickQuery := `SELECT EXISTS(SELECT 1 FROM behavior_events WHERE session_id = $1 AND event_type = 'click')`
    db.QueryRowContext(ctx, mouseQuery, sessionID).Scan(&hasMouse)
    db.QueryRowContext(ctx, clickQuery, sessionID).Scan(&hasClick)

    if hasClick && !hasMouse {
        risk += 30
    }

    if risk > 100 {
        risk = 100
    }
    return risk
}

func (d *Detector) fingerprintAnomaly(ctx context.Context, sessionID, currentFP string) int {
    lastFP, err := d.store.GetFingerprint(ctx, sessionID)
    if err != nil || lastFP == "" {
        return 0
    }
    if currentFP != "" && lastFP != currentFP {
        return 20
    }
    return 0
}

func (d *Detector) defaultSettings() *storage.ModuleSettings {
    return &storage.ModuleSettings{
        Analyzer: storage.AnalyzerSettings{
            Enabled:           true,
            RateLimiting:      true,
            PatternAnalysis:   true,
            HeadlessDetection: true,
            Thresholds: storage.AnalyzerThreshold{
                Low:    30,
                Medium: 60,
                High:   80,
            },
            Weights: storage.DefaultWeights(),
        },
    }
}

func (d *Detector) CleanCache() {
    d.cache.Range(func(key, value interface{}) bool {
        if cr, ok := value.(cachedRisk); ok {
            if time.Since(cr.ts) > d.ttl*2 {
                d.cache.Delete(key)
            }
        }
        return true
    })
}

func (d *Detector) GetCachedRisk(sessionID string) (int, bool) {
    if cached, ok := d.cache.Load(sessionID); ok {
        if cr, ok := cached.(cachedRisk); ok {
            return cr.score, true
        }
    }
    return 0, false
}