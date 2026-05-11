package detector

import (
    "context"
    "database/sql"
    "testing"
    "time"

    "humanguard/storage"
)

type mockStorage struct {
    storage.Storage
    sessions       map[string]*storage.Session
    blacklist      map[string]bool
    behaviorEvents map[string][]storage.BehaviorEvent
    riskScores     map[string]int
    db             *sql.DB
}

func (m *mockStorage) GetSession(ctx context.Context, id string) (*storage.Session, error) {
    if sess, ok := m.sessions[id]; ok {
        return sess, nil
    }
    return nil, storage.ErrSessionNotFound
}

func (m *mockStorage) IsBlacklisted(ctx context.Context, siteID, ip string) (bool, error) {
    return m.blacklist[ip], nil
}

func (m *mockStorage) GetSiteSettings(ctx context.Context, siteID string) (*storage.ModuleSettings, error) {
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
    }, nil
}

func (m *mockStorage) UpdateRiskScore(ctx context.Context, id string, score int) error {
    if m.riskScores == nil {
        m.riskScores = make(map[string]int)
    }
    m.riskScores[id] = score
    return nil
}

func (m *mockStorage) GetFingerprint(ctx context.Context, id string) (string, error) {
    if sess, ok := m.sessions[id]; ok {
        return sess.UserAgent, nil
    }
    return "", nil
}

func (m *mockStorage) GetDB() *sql.DB {
    return m.db
}

func TestHeadlessScore(t *testing.T) {
    d := &Detector{}

    tests := []struct {
        name      string
        userAgent string
        minScore  int
        maxScore  int
    }{
        {"normal chrome", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0.0.0", 0, 15},
        {"headless chrome", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 HeadlessChrome/120.0.0.0", 25, 100},
        {"phantomjs", "Mozilla/5.0 (Windows NT 6.1) AppleWebKit/534.34 PhantomJS/1.9.8", 25, 100},
        {"selenium webdriver", "Mozilla/5.0 (Windows NT 10.0) webdriver", 25, 100},
        {"multiple indicators", "HeadlessChrome/120.0.0.0 webdriver puppeteer", 75, 100},
        {"very short ua", "short", 35, 100},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            score := d.headlessScore(tt.userAgent)
            if score < tt.minScore {
                t.Errorf("score %d < min %d", score, tt.minScore)
            }
            if score > tt.maxScore {
                t.Errorf("score %d > max %d", score, tt.maxScore)
            }
        })
    }
}

func TestIsIPBlacklisted(t *testing.T) {
    siteID := "test-site"
    ctx := context.Background()

    tests := []struct {
        name      string
        ip        string
        blacklist map[string]bool
        expected  bool
    }{
        {"blacklisted ip", "1.2.3.4", map[string]bool{"1.2.3.4": true}, true},
        {"not blacklisted", "5.6.7.8", map[string]bool{}, false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mock := &mockStorage{
                sessions: map[string]*storage.Session{
                    "sess1": {ID: "sess1", SiteID: &siteID, IP: tt.ip},
                },
                blacklist: tt.blacklist,
            }
            d := New(mock)
            sess, _ := mock.GetSession(ctx, "sess1")
            result := d.isIPBlacklisted(ctx, sess)
            if result != tt.expected {
                t.Errorf("expected %v, got %v", tt.expected, result)
            }
        })
    }
}

func TestRateLimitScore(t *testing.T) {
    mock := &mockStorage{db: nil}
    d := &Detector{store: mock}
    ctx := context.Background()

    score := d.rateLimitScore(ctx, "192.168.1.1")
    if score != 0 {
        t.Errorf("expected 0 without DB, got %d", score)
    }
}

func TestBehaviorScore(t *testing.T) {
    mock := &mockStorage{db: nil}
    d := &Detector{store: mock}
    ctx := context.Background()

    score := d.behaviorScore(ctx, "nonexistent-session")
    if score != 0 {
        t.Errorf("expected 0 without DB, got %d", score)
    }
}

func TestFingerprintAnomaly(t *testing.T) {
    siteID := "test-site"
    ctx := context.Background()

    tests := []struct {
        name         string
        sessionFP    string
        currentFP    string
        expectedRisk int
    }{
        {"no fingerprint", "", "", 0},
        {"same fingerprint", "fp123", "fp123", 0},
        {"different fingerprint", "fp123", "fp456", 20},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mock := &mockStorage{
                sessions: map[string]*storage.Session{
                    "sess1": {ID: "sess1", SiteID: &siteID, UserAgent: tt.sessionFP},
                },
            }
            d := New(mock)
            risk := d.fingerprintAnomaly(ctx, "sess1", tt.currentFP)
            if risk != tt.expectedRisk {
                t.Errorf("expected %d, got %d", tt.expectedRisk, risk)
            }
        })
    }
}

func TestPrefilter(t *testing.T) {
    siteID := "test-site"
    ctx := context.Background()

    tests := []struct {
        name          string
        session       *storage.Session
        blacklist     map[string]bool
        expectedRisk  int
        expectedBlock bool
        expectedDeep  bool
    }{
        {
            name: "blacklisted ip",
            session: &storage.Session{ID: "sess1", SiteID: &siteID, IP: "1.2.3.4", UserAgent: "normal"},
            blacklist:     map[string]bool{"1.2.3.4": true},
            expectedRisk:  100,
            expectedBlock: true,
            expectedDeep:  false,
        },
        {
            name: "headless browser",
            session: &storage.Session{ID: "sess2", SiteID: &siteID, IP: "5.6.7.8", UserAgent: "HeadlessChrome/120.0.0.0"},
            blacklist:     map[string]bool{},
            expectedRisk:  25,
            expectedBlock: false,
            expectedDeep:  true,
        },
        {
            name: "normal user",
            session: &storage.Session{ID: "sess3", SiteID: &siteID, IP: "9.10.11.12", UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/120.0.0.0"},
            blacklist:     map[string]bool{},
            expectedRisk:  0,
            expectedBlock: false,
            expectedDeep:  true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mock := &mockStorage{
                sessions:  map[string]*storage.Session{tt.session.ID: tt.session},
                blacklist: tt.blacklist,
                db:        nil,
            }
            d := New(mock)

            result, err := d.Prefilter(ctx, tt.session.ID)
            if err != nil {
                t.Fatalf("Prefilter failed: %v", err)
            }

            if result.Risk < tt.expectedRisk {
                t.Errorf("risk %d < expected %d", result.Risk, tt.expectedRisk)
            }
            if result.ShouldBlock != tt.expectedBlock {
                t.Errorf("shouldBlock %v != expected %v", result.ShouldBlock, tt.expectedBlock)
            }
            if result.NeedDeep != tt.expectedDeep {
                t.Errorf("needDeep %v != expected %v", result.NeedDeep, tt.expectedDeep)
            }
        })
    }
}

func TestAnalyzeAndUpdate(t *testing.T) {
    siteID := "test-site"
    ctx := context.Background()

    mock := &mockStorage{
        sessions: map[string]*storage.Session{
            "test-session": {
                ID:        "test-session",
                SiteID:    &siteID,
                IP:        "192.168.1.100",
                UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/120.0.0.0",
                RiskScore: 0,
            },
        },
        blacklist:  map[string]bool{},
        riskScores: make(map[string]int),
        db:         nil,
    }

    d := New(mock)

    err := d.AnalyzeAndUpdate(ctx, "test-session")
    if err != nil {
        t.Errorf("AnalyzeAndUpdate failed: %v", err)
    }

    if score, ok := mock.riskScores["test-session"]; !ok {
        t.Error("risk score not updated")
    } else if score < 0 || score > 100 {
        t.Errorf("invalid risk score: %d", score)
    }
}

func TestCache(t *testing.T) {
    siteID := "test-site"
    ctx := context.Background()

    mock := &mockStorage{
        sessions: map[string]*storage.Session{
            "cached-session": {
                ID:        "cached-session",
                SiteID:    &siteID,
                IP:        "192.168.1.100",
                UserAgent: "normal",
                RiskScore: 0,
            },
        },
        blacklist:  map[string]bool{},
        riskScores: make(map[string]int),
        db:         nil,
    }

    d := New(mock)
    d.ttl = 2 * time.Second

    err := d.AnalyzeAndUpdate(ctx, "cached-session")
    if err != nil {
        t.Fatalf("first analyze failed: %v", err)
    }

    if len(mock.riskScores) == 0 {
        t.Fatal("risk score not saved after first analyze")
    }

    mock.riskScores = make(map[string]int)

    err = d.AnalyzeAndUpdate(ctx, "cached-session")
    if err != nil {
        t.Fatalf("second analyze failed: %v", err)
    }

    if len(mock.riskScores) > 0 {
        t.Error("cache not used - risk score was updated again")
    }

    time.Sleep(3 * time.Second)

    err = d.AnalyzeAndUpdate(ctx, "cached-session")
    if err != nil {
        t.Fatalf("analyze after ttl failed: %v", err)
    }

    if len(mock.riskScores) == 0 {
        t.Error("cache should have expired, but no update occurred")
    }
}

func TestCleanCache(t *testing.T) {
    d := &Detector{ttl: 1 * time.Second}

    d.cache.Store("old1", cachedRisk{score: 50, ts: time.Now().Add(-3 * time.Second)})
    d.cache.Store("old2", cachedRisk{score: 60, ts: time.Now().Add(-5 * time.Second)})
    d.cache.Store("fresh", cachedRisk{score: 70, ts: time.Now()})

    d.CleanCache()

    if _, ok := d.cache.Load("old1"); ok {
        t.Error("old entry should be removed")
    }
    if _, ok := d.cache.Load("old2"); ok {
        t.Error("old entry should be removed")
    }
    if _, ok := d.cache.Load("fresh"); !ok {
        t.Error("fresh entry should remain")
    }
}

func TestGetCachedRisk(t *testing.T) {
    d := &Detector{}

    d.cache.Store("test-session", cachedRisk{score: 75, ts: time.Now()})

    score, ok := d.GetCachedRisk("test-session")
    if !ok {
        t.Error("cached risk not found")
    }
    if score != 75 {
        t.Errorf("expected 75, got %d", score)
    }

    _, ok = d.GetCachedRisk("nonexistent")
    if ok {
        t.Error("found risk for nonexistent session")
    }
}

func TestDefaultSettings(t *testing.T) {
    d := &Detector{}
    settings := d.defaultSettings()

    if !settings.Analyzer.Enabled {
        t.Error("analyzer should be enabled by default")
    }
    if settings.Analyzer.Thresholds.Low != 30 {
        t.Errorf("expected low threshold 30, got %d", settings.Analyzer.Thresholds.Low)
    }
}

func BenchmarkHeadlessScore(b *testing.B) {
    d := &Detector{}
    ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0.0.0"

    for i := 0; i < b.N; i++ {
        d.headlessScore(ua)
    }
}