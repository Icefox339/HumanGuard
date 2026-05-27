package middleware

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type Rule struct {
	Pattern *regexp.Regexp
	Limit   rate.Limit
	Burst   int
}

type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	rules    []Rule
}

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
	ruleKey  string
}

func NewRateLimiter(rules []Rule) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rules:    rules,
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) getLimiter(ip, path string) *rate.Limiter {
	key := ip + ":" + path

	rl.mu.RLock()
	v, exists := rl.visitors[key]
	rl.mu.RUnlock()

	if exists {
		v.lastSeen = time.Now()
		return v.limiter
	}

	var limit rate.Limit = 100.0 / 60.0
	burst := 20

	for _, rule := range rl.rules {
		if rule.Pattern.MatchString(path) {
			limit = rule.Limit
			burst = rule.Burst
			break
		}
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter := rate.NewLimiter(limit, burst)
	rl.visitors[key] = &visitor{
		limiter:  limiter,
		lastSeen: time.Now(),
		ruleKey:  path,
	}
	return limiter
}

func (rl *RateLimiter) cleanup() {
	for {
		time.Sleep(5 * time.Minute)
		rl.mu.Lock()
		for key, v := range rl.visitors {
			if time.Since(v.lastSeen) > 10*time.Minute {
				delete(rl.visitors, key)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" || strings.HasPrefix(r.URL.Path, "/metrics") {
			next.ServeHTTP(w, r)
			return
		}

		ip := getRealIP(r)
		path := r.URL.Path
		limiter := rl.getLimiter(ip, path)

		if !limiter.Allow() {
			requestID := GetRequestID(r.Context())
			if requestID == "" {
				requestID = "unknown"
			}
			log.Printf("[%s] Rate limit exceeded for IP: %s, path: %s", requestID, ip, path)

			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limiter.Burst()))
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("Retry-After", "60")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			if err := json.NewEncoder(w).Encode(map[string]string{
				"error": "Too many requests. Please try again later.",
			}); err != nil {
				log.Printf("[%s] Failed to encode rate limit response: %v", requestID, err)
			}
			return
		}

		next.ServeHTTP(w, r)
	})
}

func getRealIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}
	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return xrip
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
