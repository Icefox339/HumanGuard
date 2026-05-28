package detector

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"humanguard/storage"
)

type Detector struct {
	store         storage.Storage
	cache         sync.Map
	ttl           time.Duration
	requestCounts sync.Map
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

func (d *Detector) Prefilter(ctx context.Context, sessionID string, ip, userAgent string) (*PrefilterResult, error) {
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
		return nil, fmt.Errorf("get visitor session failed: %w", err)
	}
	if session == nil {
		return &PrefilterResult{Risk: 0, ShouldBlock: false, NeedDeep: false, Reason: "no_session"}, nil
	}

	settings, err := d.store.GetSiteSettings(ctx, session.SiteID)
	if err != nil {
		settings = getDefaultSettings()
	}

	risk := 0

	// Проверка блэклиста (уже есть, оставляем)
	blacklisted, err := d.store.IsBlacklisted(ctx, session.SiteID, ip)
	if err == nil && blacklisted {
		return &PrefilterResult{
			Risk:        100,
			ShouldBlock: true,
			NeedDeep:    false,
			Reason:      "ip_blacklisted",
		}, nil
	}

	if settings.Analyzer.HeadlessDetection {
		risk += d.headlessScore(userAgent)
	}

	if settings.Analyzer.RateLimiting {
		risk += d.rateLimitScore(ip)
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

	metrics, err := d.store.GetSessionMetrics(ctx, sessionID)
	if err != nil {
		return risk, nil
	}

	if len(metrics) == 0 {
		return risk, nil
	}

	if counters, ok := metrics["counters"].(map[string]interface{}); ok {
		mouseMoves, _ := counters["mouse_move"].(float64)
		clicks, _ := counters["click"].(float64)
		scrolls, _ := counters["scroll"].(float64)
		keystrokes, _ := counters["keydown"].(float64)
		duration, _ := counters["duration_sec"].(float64)

		if mouseMoves > 1000 && duration < 30 {
			risk += 15
		}

		if clicks == 0 && scrolls > 20 {
			risk += 25
		}

		if keystrokes == 0 && mouseMoves > 100 {
			risk += 20
		}

		if duration > 0 {
			eventsPerMinute := (mouseMoves + clicks + scrolls + keystrokes) / (duration / 60)
			if eventsPerMinute > 300 {
				risk += 15
			}
			if eventsPerMinute < 5 && duration > 60 {
				risk += 30
			}
		}
	}

	if fingerprint, ok := metrics["fingerprint"].(map[string]interface{}); ok {
		if jsHash, _ := fingerprint["js_hash"].(string); jsHash == "" {
			risk += 30
		}

		if webgl, _ := fingerprint["webgl_renderer"].(string); strings.Contains(strings.ToLower(webgl), "swiftshader") {
			risk += 25
		}

		if canvasHash, _ := fingerprint["canvas_hash"].(string); canvasHash == "" {
			risk += 20
		}
	}

	if timing, ok := metrics["timing"].(map[string]interface{}); ok {
		if loadTime, _ := timing["load_time_ms"].(float64); loadTime > 0 && loadTime < 100 {
			risk += 20
		}
	}

	if risk < 0 {
		risk = 0
	}
	if risk > 100 {
		risk = 100
	}

	return risk, nil
}

func (d *Detector) AnalyzeAndUpdate(ctx context.Context, sessionID, ip, userAgent string) error {
	prefilter, err := d.Prefilter(ctx, sessionID, ip, userAgent)
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
		return err
	}

	d.cache.Store(sessionID, cachedRisk{score: finalRisk, ts: time.Now()})

	if finalRisk >= 80 {
		session, err := d.store.GetSession(ctx, sessionID)
		if err == nil && session != nil {
			if err := d.store.BlockSession(ctx, sessionID); err != nil {
				log.Printf("Failed to block session %s: %v", sessionID, err)
			}

			if err := d.store.AddToBlacklist(ctx, &storage.BlacklistEntry{
				SiteID:    session.SiteID,
				IP:        session.IP,
				Reason:    fmt.Sprintf("Auto-blocked by detector with risk score: %d", finalRisk),
				ExpiresAt: nil,
			}); err != nil {
				log.Printf("Failed to add session %s (IP: %s) to blacklist: %v", sessionID, session.IP, err)
			}
		}
	}

	return nil
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

func (d *Detector) rateLimitScore(ip string) int {
	now := time.Now()

	val, _ := d.requestCounts.LoadOrStore(ip, &[]time.Time{})
	timestamps := val.(*[]time.Time)

	clean := make([]time.Time, 0)
	for _, ts := range *timestamps {
		if now.Sub(ts) < time.Minute {
			clean = append(clean, ts)
		}
	}

	clean = append(clean, now)
	*timestamps = clean

	count := len(clean)

	if count > 100 {
		return 100
	}
	if count > 60 {
		return 50
	}
	if count > 30 {
		return 20
	}
	return 0
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

func getDefaultSettings() *storage.ModuleSettings {
	return &storage.ModuleSettings{
		Analyzer: storage.AnalyzerSettings{
			Enabled:           true,
			HeadlessDetection: true,
			RateLimiting:      true,
			PatternAnalysis:   true,
			Thresholds: storage.AnalyzerThreshold{
				Low:    30,
				Medium: 60,
				High:   80,
			},
			Weights: storage.DefaultWeights(),
		},
	}
}