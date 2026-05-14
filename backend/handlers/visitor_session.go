package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"net"
	"humanguard/storage"
	"strings"
	"github.com/google/uuid"
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
	json.NewEncoder(w).Encode(sessions)
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
	json.NewEncoder(w).Encode(sessions)
}

func (h *VisitorSessionHandler) GetSessionStats(w http.ResponseWriter, r *http.Request) {
	siteID := r.PathValue("id")
	stats, err := h.storage.GetSessionStats(r.Context(), siteID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (h *VisitorSessionHandler) CheckRequest(w http.ResponseWriter, r *http.Request) {
    siteID := r.Header.Get("X-Site-ID")
    if siteID == "" {
        writeJSON(w, http.StatusBadRequest, map[string]string{
            "error": "X-Site-ID header or site_id query param required",
        })
        return
    }
    
    site, err := h.storage.GetSite(r.Context(), siteID)
    if err != nil || site == nil {
        writeJSON(w, http.StatusNotFound, map[string]string{
            "error": "site not found",
        })
        return
    }
    
    if site.Status != "active" {
        writeJSON(w, http.StatusForbidden, map[string]string{
            "error": "site is not active",
        })
        return
    }
    
    sessionID := r.Header.Get("X-Session-ID")
    if sessionID == "" {
        cookie, err := r.Cookie("hg_session")
        if err == nil {
            sessionID = cookie.Value
        }
    }
    
    var session *storage.ActiveSession
    
    if sessionID == "" {
        session = &storage.ActiveSession{
            ID:        uuid.New().String(),
            SiteID:    siteID,  
            IP:        getRealIP(r),
            UserAgent: r.UserAgent(),
            IsActive:  true,
            Metrics:   make(map[string]interface{}),
        }
        
        if err := h.storage.CreateSession(r.Context(), session); err != nil {
            writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create session"})
            return
        }
        sessionID = session.ID
        
        http.SetCookie(w, &http.Cookie{
            Name:     "hg_session",
            Value:    sessionID,
            Path:     "/",
            HttpOnly: true,
            SameSite: http.SameSiteLaxMode,
        })
    } else {
        session, err = h.storage.GetSession(r.Context(), sessionID)
        if err != nil || session == nil {
            writeJSON(w, http.StatusNotFound, map[string]string{"error": "session not found or expired"})
            return
        }
        
        if session.SiteID != siteID {
            writeJSON(w, http.StatusForbidden, map[string]string{
                "error": "session does not belong to this site",
            })
            return
        }
    }
    
    h.storage.UpdateSessionActivity(r.Context(), sessionID)
    
    action := "allow"
    if session.RiskScore >= 80 {
        action = "block"
    } else if session.RiskScore >= 50 {
        action = "captcha"
    }
    
    writeJSON(w, http.StatusOK, map[string]interface{}{
        "action":      action,
        "session_id":  sessionID,
        "risk_score":  session.RiskScore,
    })
}
