package auth

import (
    "context"
    "net/http"
    "strings"
    "log"
    "humanguard/middleware"
)

type contextKey string

const (
    KeyUserID    contextKey = "userID"
    KeyRole      contextKey = "role"
    KeySessionID contextKey = "sessionID"
)

func AuthMiddleware(jwt *JWTService, sessionManager *UserSessionManager) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            requestID := middleware.GetRequestID(r.Context())
            
            apiKeyUserID := GetAPIKeyUserID(r.Context())
            isAPIKeyAuth := apiKeyUserID != ""
            
            if isAPIKeyAuth {
                ctx := context.WithValue(r.Context(), KeyUserID, apiKeyUserID)
                next.ServeHTTP(w, r.WithContext(ctx))
                return
            }
            
            authHeader := r.Header.Get("Authorization")
            if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
                log.Printf("[%s] Auth failed: no bearer token", requestID)
                w.Header().Set("Content-Type", "application/json")
                w.WriteHeader(http.StatusUnauthorized)
                w.Write([]byte(`{"error":"unauthorized"}`))
                return
            }
            
            token := strings.TrimPrefix(authHeader, "Bearer ")
            userID, role, sessionID, err := jwt.ValidateToken(token)
			log.Printf("[DEBUG] JWT extracted - userID: %s, role: %s, sessionID: %s", userID, role, sessionID)
            if err != nil {
                log.Printf("[%s] Auth failed: invalid token - %v", requestID, err)
                w.Header().Set("Content-Type", "application/json")
                w.WriteHeader(http.StatusUnauthorized)
                w.Write([]byte(`{"error":"invalid token"}`))
                return
            }
            
            if sessionManager != nil {
                if _, ok := sessionManager.Get(sessionID); ok {
                    sessionManager.UpdateLastSeen(sessionID)
                    log.Printf("[%s] Session validated: user=%s, session=%s", requestID, userID, sessionID)
                } else {
                    log.Printf("[%s] Session not found or expired: session=%s", requestID, sessionID)
                    w.Header().Set("Content-Type", "application/json")
                    w.WriteHeader(http.StatusUnauthorized)
                    w.Write([]byte(`{"error":"session expired or invalid"}`))
                    return
                }
            }
            
            ctx := context.WithValue(r.Context(), KeyUserID, userID)
            ctx = context.WithValue(ctx, KeyRole, role)
            ctx = context.WithValue(ctx, KeySessionID, sessionID)
			log.Printf("[DEBUG] Context set with sessionID: %s", sessionID)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

func GetUserID(ctx context.Context) string {
    id, _ := ctx.Value(KeyUserID).(string)
    return id
}

func GetRole(ctx context.Context) string {
    role, _ := ctx.Value(KeyRole).(string)
    return role
}

func GetSessionID(ctx context.Context) string {
    sid, _ := ctx.Value(KeySessionID).(string)
    return sid
}