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
			log.Printf("[Detector] PREFILTER: session=%s, using CACHED risk=%d", sessionID[:8], cr.score)
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
		log.Printf("[Detector] PREFILTER: session=%s, no session found", sessionID[:8])
		return &PrefilterResult{Risk: 0, ShouldBlock: false, NeedDeep: false, Reason: "no_session"}, nil
	}

	settings, err := d.store.GetSiteSettings(ctx, session.SiteID)
	if err != nil {
		settings = getDefaultSettings()
	}

	risk := 0
	log.Printf("[Detector] PREFILTER: session=%s, starting risk=0", sessionID[:8])

	// Проверка блэклиста
	blacklisted, err := d.store.IsBlacklisted(ctx, session.SiteID, ip)
	if err == nil && blacklisted {
		log.Printf("[Detector] PREFILTER: session=%s, IP BLACKLISTED +100", sessionID[:8])
		return &PrefilterResult{
			Risk:        100,
			ShouldBlock: true,
			NeedDeep:    false,
			Reason:      "ip_blacklisted",
		}, nil
	}

	if settings.Analyzer.HeadlessDetection {
		hs := d.headlessScore(userAgent)
		if hs > 0 {
			log.Printf("[Detector] PREFILTER: session=%s, headlessScore +%d (UA=%s)", sessionID[:8], hs, userAgent)
		}
		risk += hs
	}

	if settings.Analyzer.RateLimiting {
		rls := d.rateLimitScore(ip)
		if rls > 0 {
			log.Printf("[Detector] PREFILTER: session=%s, rateLimitScore +%d (IP=%s)", sessionID[:8], rls, ip)
		}
		risk += rls
	}

	log.Printf("[Detector] PREFILTER: session=%s, prefilter risk=%d", sessionID[:8], risk)

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
	log.Printf("[Detector] DEEP: session=%s, starting risk=%d", sessionID[:8], risk)

	metrics, err := d.store.GetSessionMetrics(ctx, sessionID)
	if err != nil {
		log.Printf("[Detector] DEEP: session=%s, failed to get metrics: %v", sessionID[:8], err)
		return risk, nil
	}

	if len(metrics) == 0 {
		log.Printf("[Detector] DEEP: session=%s, no metrics found", sessionID[:8])
		return risk, nil
	}

	if counters, ok := metrics["counters"].(map[string]interface{}); ok {
		mouseMoves, _ := counters["mouse_move"].(float64)
		clicks, _ := counters["click"].(float64)
		scrolls, _ := counters["scroll"].(float64)
		keystrokes, _ := counters["keydown"].(float64)
		duration, _ := counters["duration_sec"].(float64)

		log.Printf("[Detector] DEEP: session=%s, counters: mouse=%.0f, clicks=%.0f, scroll=%.0f, keys=%.0f, duration=%.0f",
			sessionID[:8], mouseMoves, clicks, scrolls, keystrokes, duration)

		if mouseMoves > 1000 && duration < 30 {
			log.Printf("[Detector] DEEP: session=%s, too many moves in short time +15", sessionID[:8])
			risk += 15
		}

		if clicks == 0 && scrolls > 20 {
			log.Printf("[Detector] DEEP: session=%s, no clicks but many scrolls +25", sessionID[:8])
			risk += 25
		}

		if keystrokes == 0 && mouseMoves > 100 {
			log.Printf("[Detector] DEEP: session=%s, no keystrokes but many moves +20", sessionID[:8])
			risk += 20
		}

		if duration > 0 {
			eventsPerMinute := (mouseMoves + clicks + scrolls + keystrokes) / (duration / 60)
			log.Printf("[Detector] DEEP: session=%s, events per minute: %.2f", sessionID[:8], eventsPerMinute)

			if eventsPerMinute > 300 {
				log.Printf("[Detector] DEEP: session=%s, too high event rate +15", sessionID[:8])
				risk += 15
			}
			if eventsPerMinute < 5 && duration > 60 {
				log.Printf("[Detector] DEEP: session=%s, too low event rate +30", sessionID[:8])
				risk += 30
			}
		}
	} else {
		log.Printf("[Detector] DEEP: session=%s, no counters in metrics", sessionID[:8])
	}

	if fingerprint, ok := metrics["fingerprint"].(map[string]interface{}); ok {
		webgl, _ := fingerprint["webgl_renderer"].(string)

		log.Printf("[Detector] DEEP: session=%s, fingerprint: webgl=%s", sessionID[:8], webgl)

		// ПРОВЕРКА JS_HASH И CANVAS_HASH ОТКЛЮЧЕНА ДЛЯ ТЕСТОВОГО СТЕНДА
		// if jsHash == "" {
		// 	log.Printf("[Detector] DEEP: session=%s, empty js_hash +30", sessionID[:8])
		// 	risk += 30
		// }

		if strings.Contains(strings.ToLower(webgl), "swiftshader") {
			log.Printf("[Detector] DEEP: session=%s, SwiftShader detected +25", sessionID[:8])
			risk += 25
		}

		// if canvasHash == "" {
		// 	log.Printf("[Detector] DEEP: session=%s, empty canvas_hash +20", sessionID[:8])
		// 	risk += 20
		// }
	} else {
		log.Printf("[Detector] DEEP: session=%s, no fingerprint in metrics", sessionID[:8])
	}

	if timing, ok := metrics["timing"].(map[string]interface{}); ok {
		if loadTime, _ := timing["load_time_ms"].(float64); loadTime > 0 && loadTime < 100 {
			log.Printf("[Detector] DEEP: session=%s, too fast load (%.0fms) +20", sessionID[:8], loadTime)
			risk += 20
		}
	} else {
		log.Printf("[Detector] DEEP: session=%s, no timing in metrics", sessionID[:8])
	}

	if risk < 0 {
		risk = 0
	}
	if risk > 100 {
		risk = 100
	}

	log.Printf("[Detector] DEEP: session=%s, FINAL risk=%d", sessionID[:8], risk)
	return risk, nil
}

func (d *Detector) AnalyzeAndUpdate(ctx context.Context, sessionID, ip, userAgent string) error {
	log.Printf("[Detector] ANALYZE: session=%s, starting analysis (ip=%s, ua=%s)", sessionID[:8], ip, userAgent)

	prefilter, err := d.Prefilter(ctx, sessionID, ip, userAgent)
	if err != nil {
		log.Printf("[Detector] ANALYZE: session=%s, prefilter error: %v", sessionID[:8], err)
		return err
	}

	if prefilter.Reason == "cached" {
		log.Printf("[Detector] ANALYZE: session=%s, using cached result, risk=%d", sessionID[:8], prefilter.Risk)
		return nil
	}

	finalRisk := prefilter.Risk
	log.Printf("[Detector] ANALYZE: session=%s, after prefilter risk=%d, needDeep=%v", sessionID[:8], finalRisk, prefilter.NeedDeep)

	if prefilter.NeedDeep {
		deepRisk, err := d.DeepFilter(ctx, sessionID, finalRisk)
		if err != nil {
			log.Printf("[Detector] ANALYZE: session=%s, deep filter error: %v", sessionID[:8], err)
			return err
		}
		finalRisk = deepRisk
		log.Printf("[Detector] ANALYZE: session=%s, after deep filter risk=%d", sessionID[:8], finalRisk)
	}

	if err := d.store.UpdateRiskScore(ctx, sessionID, finalRisk); err != nil {
		log.Printf("[Detector] ANALYZE: session=%s, failed to update risk: %v", sessionID[:8], err)
		return err
	}

	d.cache.Store(sessionID, cachedRisk{score: finalRisk, ts: time.Now()})
	log.Printf("[Detector] ANALYZE: session=%s, FINAL risk=%d saved to cache", sessionID[:8], finalRisk)

	if finalRisk >= 80 {
		session, err := d.store.GetSession(ctx, sessionID)
		if err == nil && session != nil {
			log.Printf("[Detector] ANALYZE: session=%s, HIGH RISK! Blocking session and adding IP to blacklist", sessionID[:8])
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
