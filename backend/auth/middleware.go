package auth

import (
    "context"
	"time"
    "crypto/sha256"
    "encoding/hex"
    "net/http"
    "strings"
    "log"
    "humanguard/middleware"
    "humanguard/storage"
)

type contextKey string
const APIKeyUserIDKey contextKey = "api_key_user_id"
const APIKeyIDKey contextKey = "api_key_id"

const (
    KeyUserID    contextKey = "userID"
    KeyRole      contextKey = "role"
    KeySessionID contextKey = "sessionID"
    KeyAPIKeyID  contextKey = "api_key_id"
)

type AuthMiddleware struct {
    jwtService     *JWTService
    sessionManager *UserSessionManager
    storage        storage.Storage
}

func NewAuthMiddleware(jwt *JWTService, sm *UserSessionManager, store storage.Storage) *AuthMiddleware {
    return &AuthMiddleware{
        jwtService:     jwt,
        sessionManager: sm,
        storage:        store,
    }
}

func (am *AuthMiddleware) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        requestID := middleware.GetRequestID(r.Context())
        
        apiKey := r.Header.Get("X-API-Key")
        if apiKey != "" {
            userID, role, apiKeyID, err := am.validateAPIKey(r.Context(), apiKey)
            if err == nil {
                log.Printf("[%s] API key authenticated: key_id=%s, user=%s, role=%s", 
                    requestID, apiKeyID[:8], userID, role)
                
                ctx := context.WithValue(r.Context(), KeyUserID, userID)
                ctx = context.WithValue(ctx, KeyRole, role)
                ctx = context.WithValue(ctx, KeyAPIKeyID, apiKeyID)
                next.ServeHTTP(w, r.WithContext(ctx))
                return
            }
            log.Printf("[%s] API key validation failed: %v", requestID, err)
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusUnauthorized)
            w.Write([]byte(`{"error":"invalid api key"}`))
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
        userID, role, sessionID, err := am.jwtService.ValidateToken(token)
        if err != nil {
            log.Printf("[%s] Auth failed: invalid token - %v", requestID, err)
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusUnauthorized)
            w.Write([]byte(`{"error":"invalid token"}`))
            return
        }
        
        if am.sessionManager != nil {
            if _, ok := am.sessionManager.Get(sessionID); ok {
                am.sessionManager.UpdateLastSeen(sessionID)
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
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func (am *AuthMiddleware) validateAPIKey(ctx context.Context, apiKey string) (userID, role, keyID string, err error) {
    if !strings.HasPrefix(apiKey, "hg_v1_") {
        return "", "", "", logError("invalid api key format")
    }
    
    hash := sha256.Sum256([]byte(apiKey))
    keyHash := hex.EncodeToString(hash[:])
    
    key, err := am.storage.GetAPIKeyByHash(ctx, keyHash)
    if err != nil || key == nil {
        return "", "", "", logError("api key not found")
    }
    
    if key.Revoked {
        return "", "", "", logError("api key revoked")
    }
    
    if key.ExpiresAt != nil && time.Now().After(*key.ExpiresAt) {
        return "", "", "", logError("api key expired")
    }
    
    go func() {
        am.storage.UpdateAPIKeyLastUsed(context.Background(), key.ID)
    }()
    
    user, err := am.storage.GetUserByID(ctx, key.UserID)
    if err != nil {
        return "", "", "", logError("user not found")
    }
    
    return user.ID, user.Role, key.ID, nil
}

func logError(msg string) error {
    log.Println("[API Key]", msg)
    return &APIKeyError{msg}
}

type APIKeyError struct {
    msg string
}

func (e *APIKeyError) Error() string {
    return e.msg
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

func GetAPIKeyUserID(ctx context.Context) string {
    if id, ok := ctx.Value(APIKeyUserIDKey).(string); ok {
        return id
    }
    return ""
}