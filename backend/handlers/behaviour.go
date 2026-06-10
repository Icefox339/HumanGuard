package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"humanguard/auth"
	"humanguard/detector"
	"humanguard/storage"
)

type BehaviorHandler struct {
	store storage.Storage
}

func NewBehaviorHandler(store storage.Storage) *BehaviorHandler {
	return &BehaviorHandler{store: store}
}

type BehaviorBatch struct {
	SessionID string                 `json:"session_id"`
	Metrics   map[string]interface{} `json:"metrics"`
}

// POST /api/sessions/{id}/behavior - ПУБЛИЧНЫЙ эндпоинт
func (h *BehaviorHandler) SubmitBehavior(w http.ResponseWriter, r *http.Request) {
	// Обработка OPTIONS preflight запросов
	if r.Method == "OPTIONS" {
		siteID := r.Header.Get("X-Site-ID")
		log.Printf("[Behavior] OPTIONS request: siteID=%s, origin=%s", siteID, r.Header.Get("Origin"))

		if siteID != "" {
			site, err := h.store.GetSite(r.Context(), siteID)
			if err != nil {
				log.Printf("[Behavior] OPTIONS: GetSite error: %v", err)
			}
			if site != nil {
				log.Printf("[Behavior] OPTIONS: site found: domain=%s, status=%s", site.Domain, site.Status)
			}
			if err == nil && site != nil && site.Status == "active" {
				origin := r.Header.Get("Origin")
				if origin == "http://"+site.Domain || origin == "https://"+site.Domain {
					log.Printf("[Behavior] OPTIONS: CORS allowed for origin=%s", origin)
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
					w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Site-ID")
					w.Header().Set("Access-Control-Max-Age", "86400")
					w.WriteHeader(http.StatusNoContent)
					return
				} else {
					log.Printf("[Behavior] OPTIONS: origin mismatch: got=%s, expected=http://%s or https://%s", origin, site.Domain, site.Domain)
				}
			} else {
				log.Printf("[Behavior] OPTIONS: site not active or not found")
			}
		} else {
			log.Printf("[Behavior] OPTIONS: siteID is empty")
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	sessionID := r.PathValue("id")
	if sessionID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "session_id required"})
		return
	}

	session, err := h.store.GetSession(r.Context(), sessionID)
	if err != nil || session == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "session not found"})
		return
	}

	siteID := r.Header.Get("X-Site-ID")

	if siteID != "" && session.SiteID != siteID {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "site_id mismatch"})
		return
	}

	var batch BehaviorBatch
	if err := json.NewDecoder(r.Body).Decode(&batch); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	if batch.SessionID != sessionID {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "session_id mismatch"})
		return
	}

	if len(batch.Metrics) > 0 {
		existingMetrics, _ := h.store.GetSessionMetrics(r.Context(), sessionID)
		if existingMetrics == nil {
			existingMetrics = make(map[string]interface{})
		}

		for k, v := range batch.Metrics {
			existingMetrics[k] = v
		}

		if err := h.store.UpdateSessionMetrics(r.Context(), sessionID, existingMetrics); err != nil {
			log.Printf("Failed to update metrics for session %s: %v", sessionID, err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update metrics"})
			return
		}
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		d := detector.New(h.store)
		if err := d.AnalyzeAndUpdate(ctx, sessionID, session.IP, session.UserAgent); err != nil {
			log.Printf("Deep analysis failed for session %s: %v", sessionID, err)
		}
	}()

	// Устанавливаем CORS заголовки ПЕРЕД отправкой ответа
	origin := r.Header.Get("Origin")
	if origin != "" {
		site, err := h.store.GetSite(r.Context(), siteID)
		if err == nil && site != nil && site.Status == "active" {
			if origin == "http://"+site.Domain || origin == "https://"+site.Domain {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
		}
	}

	writeJSON(w, http.StatusAccepted, map[string]string{"status": "accepted"})
}

// POST /api/sessions/{id}/analyze
func (h *BehaviorHandler) TriggerAnalysis(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")
	if sessionID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "session_id required"})
		return
	}

	userID := auth.GetUserID(r.Context())
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	d := detector.New(h.store)
	if err := d.AnalyzeAndUpdate(r.Context(), sessionID, "", ""); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	session, err := h.store.GetSession(r.Context(), sessionID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get session"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"risk_score": session.RiskScore,
		"is_blocked": session.IsBlocked,
	})
}
