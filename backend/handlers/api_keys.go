package handlers

import (
    "crypto/rand"
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "net/http"
    "time"

    "humanguard/auth"
    "humanguard/storage"
)

type APIKeyHandler struct {
    storage storage.Storage
}

func NewAPIKeyHandler(store storage.Storage) *APIKeyHandler {
    return &APIKeyHandler{storage: store}
}

type CreateAPIKeyRequest struct {
    Name      string `json:"name"`
    ExpiresIn *int   `json:"expires_in_days,omitempty"`
}

type APIKeyResponse struct {
    ID         string     `json:"id"`
    Name       string     `json:"name"`
    Key        string     `json:"key"`
    Prefix     string     `json:"prefix"`
    CreatedAt  time.Time  `json:"created_at"`
    ExpiresAt  *time.Time `json:"expires_at,omitempty"`
    LastUsedAt *time.Time `json:"last_used_at,omitempty"`
    Revoked    bool       `json:"revoked"`
}

type APIKeyListResponse struct {
    ID         string     `json:"id"`
    Name       string     `json:"name"`
    Prefix     string     `json:"prefix"`
    CreatedAt  time.Time  `json:"created_at"`
    ExpiresAt  *time.Time `json:"expires_at,omitempty"`
    LastUsedAt *time.Time `json:"last_used_at,omitempty"`
    Revoked    bool       `json:"revoked"`
}

func (h *APIKeyHandler) CreateAPIKey(w http.ResponseWriter, r *http.Request) {
    userID := auth.GetUserID(r.Context())
    if userID == "" {
        apiKeyUserID := auth.GetAPIKeyUserID(r.Context())
        if apiKeyUserID != "" {
            writeJSON(w, http.StatusForbidden, map[string]string{
                "error": "API keys cannot create new API keys. Use JWT authentication.",
            })
            return
        }
        writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
        return
    }

    var req CreateAPIKeyRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
        return
    }

    if req.Name == "" {
        writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name is required"})
        return
    }

    prefix := "hg_v1_"
    bytes := make([]byte, 32)
    if _, err := rand.Read(bytes); err != nil {
        writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to generate key"})
        return
    }
    keyRaw := prefix + hex.EncodeToString(bytes)

    hash := sha256.Sum256([]byte(keyRaw))
    keyHash := hex.EncodeToString(hash[:])

    var expiresAt *time.Time
    if req.ExpiresIn != nil && *req.ExpiresIn > 0 {
        t := time.Now().Add(time.Duration(*req.ExpiresIn) * 24 * time.Hour)
        expiresAt = &t
    }

    apiKey := &storage.APIKey{
        UserID:    userID,
        Name:      req.Name,
        KeyHash:   keyHash,
        Prefix:    prefix,
        ExpiresAt: expiresAt,
        CreatedBy: &userID,
    }

    if err := h.storage.CreateAPIKey(r.Context(), apiKey); err != nil {
        writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create API key"})
        return
    }

    writeJSON(w, http.StatusCreated, APIKeyResponse{
        ID:        apiKey.ID,
        Name:      apiKey.Name,
        Key:       keyRaw,
        Prefix:    apiKey.Prefix,
        CreatedAt: apiKey.CreatedAt,
        ExpiresAt: apiKey.ExpiresAt,
        Revoked:   false,
    })
}

func (h *APIKeyHandler) ListAPIKeys(w http.ResponseWriter, r *http.Request) {
    userID := auth.GetUserID(r.Context())
    if userID == "" {
        userID = auth.GetAPIKeyUserID(r.Context())
        if userID == "" {
            writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
            return
        }
    }

    keys, err := h.storage.ListAPIKeys(r.Context(), userID)
    if err != nil {
        writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list API keys"})
        return
    }

    response := make([]APIKeyListResponse, len(keys))
    for i, key := range keys {
        response[i] = APIKeyListResponse{
            ID:         key.ID,
            Name:       key.Name,
            Prefix:     key.Prefix,
            CreatedAt:  key.CreatedAt,
            ExpiresAt:  key.ExpiresAt,
            LastUsedAt: key.LastUsedAt,
            Revoked:    key.Revoked,
        }
    }

    writeJSON(w, http.StatusOK, response)
}

func (h *APIKeyHandler) RevokeAPIKey(w http.ResponseWriter, r *http.Request) {
    keyID := r.PathValue("id")
    if keyID == "" {
        writeJSON(w, http.StatusBadRequest, map[string]string{"error": "key id required"})
        return
    }

    userID := auth.GetUserID(r.Context())
    if userID == "" {
        userID = auth.GetAPIKeyUserID(r.Context())
        if userID == "" {
            writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
            return
        }
    }

    key, err := h.storage.GetAPIKeyByID(r.Context(), keyID)
    if err != nil || key == nil {
        writeJSON(w, http.StatusNotFound, map[string]string{"error": "API key not found"})
        return
    }

    if key.UserID != userID {
        writeJSON(w, http.StatusForbidden, map[string]string{"error": "not your API key"})
        return
    }

    if err := h.storage.RevokeAPIKey(r.Context(), keyID); err != nil {
        writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to revoke key"})
        return
    }

    writeJSON(w, http.StatusNoContent, nil)
}

func (h *APIKeyHandler) DeleteAPIKey(w http.ResponseWriter, r *http.Request) {
    keyID := r.PathValue("id")
    if keyID == "" {
        writeJSON(w, http.StatusBadRequest, map[string]string{"error": "key id required"})
        return
    }

    userID := auth.GetUserID(r.Context())
    if userID == "" {
        userID = auth.GetAPIKeyUserID(r.Context())
        if userID == "" {
            writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
            return
        }
    }

    key, err := h.storage.GetAPIKeyByID(r.Context(), keyID)
    if err != nil || key == nil {
        writeJSON(w, http.StatusNotFound, map[string]string{"error": "API key not found"})
        return
    }

    if key.UserID != userID {
        writeJSON(w, http.StatusForbidden, map[string]string{"error": "not your API key"})
        return
    }

    if err := h.storage.DeleteAPIKey(r.Context(), keyID); err != nil {
        writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete key"})
        return
    }

    writeJSON(w, http.StatusNoContent, nil)
}