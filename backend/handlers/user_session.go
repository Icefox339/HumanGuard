package handlers

import (
    "encoding/json"
    "net/http"
    "humanguard/auth"
)

type UserSessionHandler struct {
    sessionManager *auth.UserSessionManager
}

func NewUserSessionHandler(sm *auth.UserSessionManager) *UserSessionHandler {
    return &UserSessionHandler{sessionManager: sm}
}

// GET /api/admin/users/sessions - список сессий ВСЕХ пользователей (только для пользователей с role=admin)
func (h *UserSessionHandler) ListAllUserSessions(w http.ResponseWriter, r *http.Request) {
    currentUserRole := auth.GetRole(r.Context())
    if currentUserRole != "admin" {
        http.Error(w, "Forbidden: admin role required", http.StatusForbidden)
        return
    }
    
    sessions := h.sessionManager.ListAll()
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "total":    len(sessions),
        "sessions": sessions,
    })
}

// DELETE /api/admin/users/sessions/{session_id} - принудительное удаление любой сессии (только для админов)
func (h *UserSessionHandler) ForceRevokeSession(w http.ResponseWriter, r *http.Request) {
    currentUserRole := auth.GetRole(r.Context())
    if currentUserRole != "admin" {
        http.Error(w, "Forbidden: admin role required", http.StatusForbidden)
        return
    }
    
    targetSessionID := r.PathValue("session_id")
    if targetSessionID == "" {
        http.Error(w, "session_id required", http.StatusBadRequest)
        return
    }
    
    sess, ok := h.sessionManager.Get(targetSessionID)
    if !ok {
        http.Error(w, "Session not found", http.StatusNotFound)
        return
    }
    
    h.sessionManager.Delete(targetSessionID)
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "message":      "Session revoked",
        "session_id":   sess.ID,
        "user_id":      sess.UserID,
        "user_email":   sess.Email,
    })
}

// GET /api/admin/users/sessions/stats - статистика по сессиям (только для админов)
func (h *UserSessionHandler) GetSessionsStats(w http.ResponseWriter, r *http.Request) {
    currentUserRole := auth.GetRole(r.Context())
    if currentUserRole != "admin" {
        http.Error(w, "Forbidden: admin role required", http.StatusForbidden)
        return
    }
    
    sessions := h.sessionManager.ListAll()
    userSessionsMap := make(map[string]int)
    userInfoMap := make(map[string]map[string]string)
    
    for _, sess := range sessions {
        userSessionsMap[sess.UserID]++
        if _, exists := userInfoMap[sess.UserID]; !exists {
            userInfoMap[sess.UserID] = map[string]string{
                "email": sess.Email,
                "role":  sess.Role,
            }
        }
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "total_sessions":      len(sessions),
        "total_users":         len(userSessionsMap),
        "sessions_per_user":   userSessionsMap,
        "users_info":          userInfoMap,
    })
}