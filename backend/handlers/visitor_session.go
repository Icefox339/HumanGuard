package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"humanguard/storage"

	"github.com/google/uuid"
)

type VisitorSessionHandler struct {
	storage storage.Storage
}

func NewVisitorSessionHandler(store storage.Storage) *VisitorSessionHandler {
	return &VisitorSessionHandler{storage: store}
}

func (h *VisitorSessionHandler) CreateSession(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SiteID    string `json:"site_id"`
		IP        string `json:"ip"`
		UserAgent string `json:"user_agent"`
		Device    string `json:"device"`
		Location  string `json:"location"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.SiteID == "" || req.IP == "" {
		http.Error(w, "site_id and ip are required", http.StatusBadRequest)
		return
	}
	session := &storage.ActiveSession{
		ID:        uuid.New().String(),
		SiteID:    req.SiteID,
		IP:        req.IP,
		UserAgent: req.UserAgent,
		Device:    req.Device,
		Location:  req.Location,
		IsActive:  true,
		RiskScore: 0,
		Metrics:   make(map[string]interface{}),
	}
	if err := h.storage.CreateSession(r.Context(), session); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(session)
}

func (h *VisitorSessionHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	session, err := h.storage.GetSession(r.Context(), id)
	if err != nil {
		if errors.Is(err, storage.ErrSessionNotFound) {
			http.Error(w, "Session not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}

func (h *VisitorSessionHandler) DeactivateSession(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.storage.DeactivateSession(r.Context(), id); err != nil {
		if errors.Is(err, storage.ErrSessionNotFound) {
			http.Error(w, "Session not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *VisitorSessionHandler) BlockSession(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.storage.BlockSession(r.Context(), id); err != nil {
		if errors.Is(err, storage.ErrSessionNotFound) {
			http.Error(w, "Session not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *VisitorSessionHandler) UnblockSession(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.storage.UnblockSession(r.Context(), id); err != nil {
		if errors.Is(err, storage.ErrSessionNotFound) {
			http.Error(w, "Session not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *VisitorSessionHandler) UpdateRiskScore(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct{ RiskScore int `json:"risk_score"` }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.RiskScore < 0 || req.RiskScore > 100 {
		http.Error(w, "risk_score must be between 0 and 100", http.StatusBadRequest)
		return
	}
	if err := h.storage.UpdateRiskScore(r.Context(), id, req.RiskScore); err != nil {
		if errors.Is(err, storage.ErrSessionNotFound) {
			http.Error(w, "Session not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *VisitorSessionHandler) UpdateSessionActivity(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.storage.UpdateSessionActivity(r.Context(), id); err != nil {
		if errors.Is(err, storage.ErrSessionNotFound) {
			http.Error(w, "Session not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *VisitorSessionHandler) MarkCaptchaShown(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.storage.MarkCaptchaShown(r.Context(), id); err != nil {
		if errors.Is(err, storage.ErrSessionNotFound) {
			http.Error(w, "Session not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *VisitorSessionHandler) CleanupExpiredSessions(w http.ResponseWriter, r *http.Request) {
	count, err := h.storage.CleanupExpiredSessions(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int64{"deleted": count})
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