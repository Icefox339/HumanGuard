// backend/detector/detector_test.go
package detector

import (
    "context"
    "sync"
    "testing"
    "time"

    "humanguard/storage"
)

type mockStorage struct {
    storage.Storage
    blacklisted bool
}

func (m *mockStorage) IsBlacklisted(ctx context.Context, siteID, ip string) (bool, error) {
    return m.blacklisted, nil
}

func (m *mockStorage) GetSessionMetrics(ctx context.Context, id string) (map[string]interface{}, error) {
    return make(map[string]interface{}), nil
}

func (m *mockStorage) UpdateRiskScore(ctx context.Context, id string, score int) error {
    return nil
}

func (m *mockStorage) GetSession(ctx context.Context, id string) (*storage.Session, error) {
    siteID := "test-site"
    return &storage.Session{
        ID:        id,
        SiteID:    &siteID,
        IP:        "192.168.1.1",
        RiskScore: 0,
    }, nil
}

func (m *mockStorage) BlockSession(ctx context.Context, id string) error {
    return nil
}

func (m *mockStorage) AddToBlacklist(ctx context.Context, entry *storage.BlacklistEntry) error {
    return nil
}

func (m *mockStorage) GetSiteSettings(ctx context.Context, siteID string) (*storage.ModuleSettings, error) {
    return &storage.ModuleSettings{
        Analyzer: storage.AnalyzerSettings{
            Enabled:           true,
            HeadlessDetection: true,
            RateLimiting:      true,
            Thresholds: storage.AnalyzerThreshold{
                Low:    30,
                Medium: 60,
                High:   80,
            },
        },
    }, nil
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

func TestRateLimitScore(t *testing.T) {
    d := &Detector{}
    d.requestCounts = sync.Map{}

    for i := 0; i < 30; i++ {
        score := d.rateLimitScore("192.168.1.1")
        if i < 30 && score != 0 {
            t.Errorf("expected 0 for request %d, got %d", i, score)
        }
    }

    for i := 0; i < 40; i++ {
        d.rateLimitScore("192.168.1.1")
    }
    score := d.rateLimitScore("192.168.1.1")
    if score < 20 {
        t.Errorf("expected score >=20 after many requests, got %d", score)
    }
}

func TestPrefilter(t *testing.T) {
    store := &mockStorage{blacklisted: false}
    d := New(store)
    ctx := context.Background()

    result, err := d.Prefilter(ctx, "test-session", "192.168.1.1", "Mozilla/5.0")
    if err != nil {
        t.Fatalf("Prefilter failed: %v", err)
    }

    if result.Risk < 0 || result.Risk > 100 {
        t.Errorf("invalid risk score: %d", result.Risk)
    }
}

func TestPrefilterWithBlacklist(t *testing.T) {
    store := &mockStorage{blacklisted: true}
    d := New(store)
    ctx := context.Background()

    result, err := d.Prefilter(ctx, "test-session", "192.168.1.100", "Mozilla/5.0")
    if err != nil {
        t.Fatalf("Prefilter failed: %v", err)
    }

    if result.Risk != 100 {
        t.Errorf("expected risk 100 for blacklisted IP, got %d", result.Risk)
    }
    if !result.ShouldBlock {
        t.Error("expected ShouldBlock=true for blacklisted IP")
    }
    if result.Reason != "ip_blacklisted" {
        t.Errorf("expected reason 'ip_blacklisted', got '%s'", result.Reason)
    }
}

func TestCache(t *testing.T) {
    store := &mockStorage{blacklisted: false}
    d := New(store)
    ctx := context.Background()

    if _, ok := d.GetCachedRisk("test-session"); ok {
        t.Error("session should not be cached initially")
    }

    err := d.AnalyzeAndUpdate(ctx, "test-session", "1.1.1.1", "Mozilla/5.0")
    if err != nil {
        t.Fatalf("AnalyzeAndUpdate failed: %v", err)
    }

    risk, ok := d.GetCachedRisk("test-session")
    if !ok {
        t.Error("session should be cached after AnalyzeAndUpdate")
    }
    if risk < 0 || risk > 100 {
        t.Errorf("invalid risk score: %d", risk)
    }

    result, err := d.Prefilter(ctx, "test-session", "1.1.1.1", "Mozilla/5.0")
    if err != nil {
        t.Fatalf("Prefilter failed: %v", err)
    }

    if result.Reason != "cached" {
        t.Errorf("expected cached result, got reason: %s", result.Reason)
    }

    if result.Risk != risk {
        t.Errorf("expected risk %d from cache, got %d", risk, result.Risk)
    }
}

func TestCleanCache(t *testing.T) {
    store := &mockStorage{blacklisted: false}
    d := New(store)
    d.ttl = 1 * time.Second

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

func TestDeepFilterWithMetrics(t *testing.T) {
    store := &mockStorageWithMetrics{
        metrics: map[string]interface{}{
            "counters": map[string]interface{}{
                "mouse_move":   1500,
                "click":        0,
                "scroll":       100,
                "keydown":      0,
                "duration_sec": 45.0,
            },
            "fingerprint": map[string]interface{}{
                "js_hash":        "",
                "webgl_renderer": "SwiftShader",
                "canvas_hash":    "",
            },
            "timing": map[string]interface{}{
                "load_time_ms": 45,
            },
        },
    }
    d := New(store)

    newRisk, err := d.DeepFilter(context.Background(), "test-session", 30)
    if err != nil {
        t.Fatalf("DeepFilter failed: %v", err)
    }

    if newRisk != 100 {
        t.Errorf("Expected risk 100, got %d", newRisk)
    }
}

func TestAnalyzeAndUpdate(t *testing.T) {
    store := &mockStorageWithMetrics{
        metrics: map[string]interface{}{
            "counters": map[string]interface{}{
                "mouse_move":   50,
                "click":        10,
                "scroll":       5,
                "keydown":      20,
                "duration_sec": 60.0,
            },
            "fingerprint": map[string]interface{}{
                "js_hash":        "abc123def456",
                "webgl_renderer": "ANGLE (NVIDIA Corporation, NVIDIA GeForce RTX 3080, Direct3D 11 vs_5_0 ps_5_0)",
                "canvas_hash":    "canvas_hash_123",
            },
            "timing": map[string]interface{}{
                "load_time_ms": 800,
            },
        },
    }
    d := New(store)
    ctx := context.Background()

    err := d.AnalyzeAndUpdate(ctx, "test-session", "192.168.1.1", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
    if err != nil {
        t.Fatalf("AnalyzeAndUpdate failed: %v", err)
    }

    risk, ok := d.GetCachedRisk("test-session")
    if !ok {
        t.Error("session should be cached after analysis")
    }

    if risk > 20 {
        t.Errorf("expected risk <= 20 for ideal behavior, got %d", risk)
    }
}

type mockStorageWithMetrics struct {
    storage.Storage
    metrics     map[string]interface{}
    riskScore   int
    blockCalled bool
}

func (m *mockStorageWithMetrics) IsBlacklisted(ctx context.Context, siteID, ip string) (bool, error) {
    return false, nil
}

func (m *mockStorageWithMetrics) GetSessionMetrics(ctx context.Context, id string) (map[string]interface{}, error) {
    return m.metrics, nil
}

func (m *mockStorageWithMetrics) UpdateRiskScore(ctx context.Context, id string, score int) error {
    m.riskScore = score
    return nil
}

func (m *mockStorageWithMetrics) GetSession(ctx context.Context, id string) (*storage.Session, error) {
    siteID := "test-site"
    return &storage.Session{
        ID:        id,
        SiteID:    &siteID,
        IP:        "192.168.1.1",
        RiskScore: m.riskScore,
    }, nil
}

func (m *mockStorageWithMetrics) BlockSession(ctx context.Context, id string) error {
    m.blockCalled = true
    return nil
}

func (m *mockStorageWithMetrics) AddToBlacklist(ctx context.Context, entry *storage.BlacklistEntry) error {
    return nil
}

func (m *mockStorageWithMetrics) GetSiteSettings(ctx context.Context, siteID string) (*storage.ModuleSettings, error) {
    return &storage.ModuleSettings{
        Analyzer: storage.AnalyzerSettings{
            Enabled:           true,
            HeadlessDetection: true,
            RateLimiting:      true,
            Thresholds: storage.AnalyzerThreshold{
                Low:    30,
                Medium: 60,
                High:   80,
            },
        },
    }, nil
}

func BenchmarkHeadlessScore(b *testing.B) {
    d := &Detector{}
    ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0.0.0"

    for i := 0; i < b.N; i++ {
        d.headlessScore(ua)
    }
}

func BenchmarkRateLimitScore(b *testing.B) {
    d := &Detector{}
    d.requestCounts = sync.Map{}

    for i := 0; i < b.N; i++ {
        d.rateLimitScore("192.168.1.1")
    }
}

func BenchmarkDeepFilter(b *testing.B) {
    store := &mockStorageWithMetrics{
        metrics: map[string]interface{}{
            "counters": map[string]interface{}{
                "mouse_move":   1500,
                "click":        0,
                "scroll":       100,
                "keydown":      0,
                "duration_sec": 45.0,
            },
            "fingerprint": map[string]interface{}{
                "js_hash":        "",
                "webgl_renderer": "SwiftShader",
                "canvas_hash":    "",
            },
        },
    }
    d := New(store)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        d.DeepFilter(context.Background(), "test-session", 30)
    }
}