package handlers

import (
    "encoding/json"
    "log"
    "net/http"

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

// POST /api/sessions/{id}/behavior
func (h *BehaviorHandler) SubmitBehavior(w http.ResponseWriter, r *http.Request) {
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
        if err := h.store.UpdateSessionMetrics(r.Context(), sessionID, batch.Metrics); err != nil {
            log.Printf("Failed to update metrics for session %s: %v", sessionID, err)
            writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update metrics"})
            return
        }
    }

    go func() {
        d := detector.New(h.store)
        if err := d.AnalyzeAndUpdate(r.Context(), sessionID, "", ""); err != nil {
            log.Printf("Deep analysis failed for session %s: %v", sessionID, err)
        }
    }()

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