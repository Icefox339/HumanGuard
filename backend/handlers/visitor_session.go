package handlers

import (
	"bytes"
	"encoding/json"
	"github.com/google/uuid"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"

	"humanguard/storage"
)

type VisitorSessionHandler struct {
	storage storage.Storage
}

func NewVisitorSessionHandler(store storage.Storage) *VisitorSessionHandler {
	return &VisitorSessionHandler{storage: store}
}

func getRealIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}
	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return xrip
	}
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	if host == "" {
		return r.RemoteAddr
	}
	return host
}

func (h *VisitorSessionHandler) GetSessionsBySite(w http.ResponseWriter, r *http.Request) {
	siteID := r.PathValue("id")
	limit := 100
	if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && l > 0 {
		limit = l
	}
	sessions, err := h.storage.GetActiveSessionsBySite(r.Context(), siteID, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if sessions == nil {
		sessions = []*storage.ActiveSession{}
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(sessions); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *VisitorSessionHandler) GetSuspiciousSessions(w http.ResponseWriter, r *http.Request) {
	siteID := r.PathValue("id")
	minRisk := 60
	if mr, err := strconv.Atoi(r.URL.Query().Get("min_risk")); err == nil && mr >= 0 && mr <= 100 {
		minRisk = mr
	}
	sessions, err := h.storage.GetSuspiciousSessions(r.Context(), siteID, minRisk)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if sessions == nil {
		sessions = []*storage.ActiveSession{}
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(sessions); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *VisitorSessionHandler) GetSessionStats(w http.ResponseWriter, r *http.Request) {
	siteID := r.PathValue("id")
	stats, err := h.storage.GetSessionStats(r.Context(), siteID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *VisitorSessionHandler) CheckRequest(w http.ResponseWriter, r *http.Request) {
	siteID := r.Header.Get("X-Site-ID")
	log.Printf("[CheckRequest] SiteID: %s", siteID)

	// Обработка OPTIONS preflight запросов
	if r.Method == "OPTIONS" {
		siteID := r.Header.Get("X-Site-ID")
		if siteID != "" {
			site, err := h.storage.GetSite(r.Context(), siteID)
			if err == nil && site != nil && site.Status == "active" {
				origin := r.Header.Get("Origin")
				if origin == "http://"+site.Domain || origin == "https://"+site.Domain {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
					w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Site-ID, X-Session-ID")
					w.Header().Set("Access-Control-Max-Age", "86400")
					w.WriteHeader(http.StatusNoContent)
					return
				}
			}
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if siteID == "" {
		log.Printf("[CheckRequest] ERROR: no X-Site-ID")
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "X-Site-ID header required"})
		return
	}

	site, err := h.storage.GetSite(r.Context(), siteID)
	if err != nil || site == nil {
		log.Printf("[CheckRequest] ERROR: site not found: %v", err)
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "site not found"})
		return
	}

	if site.Status != "active" {
		log.Printf("[CheckRequest] ERROR: site not active: %s", site.Status)
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "site is not active"})
		return
	}

	ip := getRealIP(r)
	log.Printf("[CheckRequest] IP: %s", ip)

	blacklisted, err := h.storage.IsBlacklisted(r.Context(), siteID, ip)
	if err == nil && blacklisted {
		log.Printf("[CheckRequest] ERROR: IP blacklisted")
		writeJSON(w, http.StatusForbidden, map[string]interface{}{
			"error":  "your IP is blocked",
			"action": "block",
		})
		return
	}

	var reqBody struct {
		SiteID        string `json:"site_id"`
		CaptchaPassed bool   `json:"captcha_passed"`
	}

	if r.Body != nil {
		bodyBytes, _ := io.ReadAll(r.Body)
		log.Printf("[CheckRequest] Raw body: %s", string(bodyBytes))
		if len(bodyBytes) > 0 {
			json.Unmarshal(bodyBytes, &reqBody)
			log.Printf("[CheckRequest] Parsed: CaptchaPassed=%v", reqBody.CaptchaPassed)
		}
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	sessionID := r.Header.Get("X-Session-ID")
	if sessionID == "" {
		cookie, err := r.Cookie("hg_session")
		if err == nil && cookie.Value != "" {
			sessionID = cookie.Value
			log.Printf("[CheckRequest] Got session from cookie: %s", sessionID)
		}
	}

	var session *storage.ActiveSession

	if sessionID != "" {
		session, err = h.storage.GetSession(r.Context(), sessionID)
		if err != nil || session == nil {
			log.Printf("[CheckRequest] Session not found or expired: %s", sessionID)
			session = nil
			sessionID = ""
			http.SetCookie(w, &http.Cookie{
				Name:     "hg_session",
				Value:    "",
				Path:     "/",
				HttpOnly: true,
				Secure:   true,
				MaxAge:   -1,
			})
		} else if session.SiteID != siteID {
			log.Printf("[CheckRequest] Session belongs to different site: %s vs %s", session.SiteID, siteID)
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "session does not belong to this site"})
			return
		} else {
			log.Printf("[CheckRequest] Existing session found: %s, risk=%d", sessionID, session.RiskScore)
		}
	}

	if session == nil {
		log.Printf("[CheckRequest] Creating new session")
		session = &storage.ActiveSession{
			ID:        uuid.New().String(),
			SiteID:    siteID,
			IP:        ip,
			UserAgent: r.UserAgent(),
			IsActive:  true,
			RiskScore: 0,
			Metrics:   make(map[string]interface{}),
		}

		if err := h.storage.CreateSession(r.Context(), session); err != nil {
			log.Printf("[CheckRequest] Failed to create session: %v", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create session"})
			return
		}
		sessionID = session.ID
		log.Printf("[CheckRequest] New session created: %s", sessionID)

		http.SetCookie(w, &http.Cookie{
			Name:     "hg_session",
			Value:    sessionID,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			Secure:   true,
			MaxAge:   1800,
		})
	}

	if err := h.storage.UpdateSessionActivity(r.Context(), sessionID); err != nil {
		log.Printf("[CheckRequest] Failed to update activity: %v", err)
	}

	action := "allow"
	currentRisk := session.RiskScore

	log.Printf("[CheckRequest] Current risk: %d, CaptchaPassed: %v", currentRisk, reqBody.CaptchaPassed)

	if reqBody.CaptchaPassed {
		log.Printf("[CheckRequest] Captcha passed! Setting risk to 0")
		h.storage.UpdateRiskScore(r.Context(), sessionID, 0)
		currentRisk = 0
		action = "allow"
	} else if currentRisk >= 80 {
		log.Printf("[CheckRequest] High risk, action: block")
		action = "block"
	} else if currentRisk >= 50 {
		log.Printf("[CheckRequest] Medium risk, action: captcha")
		action = "captcha"
	} else {
		log.Printf("[CheckRequest] Low risk, action: allow")
	}

	// Устанавливаем CORS заголовки ПЕРЕД отправкой ответа
	origin := r.Header.Get("Origin")
	if origin != "" && site.Status == "active" {
		if origin == "http://"+site.Domain || origin == "https://"+site.Domain {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Site-ID, X-Session-ID")
			w.Header().Set("Access-Control-Expose-Headers", "X-Session-ID, X-Request-ID")
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"action":     action,
		"session_id": sessionID,
		"risk_score": currentRisk,
	})
}
