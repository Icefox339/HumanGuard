package handlers

import (
	"encoding/json"
	"errors"
	"humanguard/auth"
	"humanguard/storage"
	"net/http"
)

type SiteHandler struct {
	storage storage.Storage
}

func NewSiteHandler(store storage.Storage) *SiteHandler {
	return &SiteHandler{storage: store}
}

// POST /api/sites
func (h *SiteHandler) CreateSite(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID       string `json:"user_id"`
		Name         string `json:"name"`
		Domain       string `json:"domain"`
		OriginServer string `json:"origin_server"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.UserID == "" || req.Name == "" || req.Domain == "" || req.OriginServer == "" {
		http.Error(w, "user_id, name, domain, origin_server are required", http.StatusBadRequest)
		return
	}

	site := &storage.Site{
		UserID:       req.UserID,
		Name:         req.Name,
		Domain:       req.Domain,
		OriginServer: req.OriginServer,
		Status:       "verifying",
	}

	if err := h.storage.CreateSite(r.Context(), site); err != nil {
		if errors.Is(err, storage.ErrSiteAlreadyExists) {
			http.Error(w, "Site already exists", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(site)
}

// GET /api/sites
func (h *SiteHandler) ListSites(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	sites, err := h.storage.GetSitesByUserID(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if sites == nil {
		sites = []*storage.Site{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sites)
}

// GET /api/sites/{id}
func (h *SiteHandler) GetSite(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	site, err := h.storage.GetSite(r.Context(), id)
	if err != nil {
		if errors.Is(err, storage.ErrSiteNotFound) {
			http.Error(w, "Site not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(site)
}

// PUT /api/sites/{id}
func (h *SiteHandler) UpdateSite(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req struct {
		Name         string `json:"name"`
		OriginServer string `json:"origin_server"`
		Status       string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	site, err := h.storage.GetSite(r.Context(), id)
	if err != nil {
		http.Error(w, "Site not found", http.StatusNotFound)
		return
	}

	if req.Name != "" {
		site.Name = req.Name
	}
	if req.OriginServer != "" {
		site.OriginServer = req.OriginServer
	}
	if req.Status != "" {
		site.Status = req.Status
	}

	if err := h.storage.UpdateSite(r.Context(), site); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(site)
}

// DELETE /api/sites/{id}
func (h *SiteHandler) DeleteSite(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := h.storage.DeleteSite(r.Context(), id); err != nil {
		if errors.Is(err, storage.ErrSiteNotFound) {
			http.Error(w, "Site not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// POST /api/sites/{id}/activate
func (h *SiteHandler) ActivateSite(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := h.storage.ActivateSite(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// POST /api/sites/{id}/suspend
func (h *SiteHandler) SuspendSite(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := h.storage.SuspendSite(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GET /api/sites/{id}/settings
func (h *SiteHandler) GetSiteSettings(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	settings, err := h.storage.GetSiteSettings(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}

// PUT /api/sites/{id}/settings
func (h *SiteHandler) UpdateSiteSettings(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var settings storage.ModuleSettings
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.storage.UpdateSiteSettings(r.Context(), id, &settings); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}
