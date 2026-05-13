package auth

import (
    "context"
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "log"
    "net/http"
    "strings"
    "time"
    "humanguard/storage"
    "humanguard/middleware"
)

const APIKeyUserIDKey contextKey = "api_key_user_id"
const APIKeyIDKey contextKey = "api_key_id"

type APIKeyAuthenticator struct {
    storage storage.Storage
}

func NewAPIKeyAuthenticator(store storage.Storage) *APIKeyAuthenticator {
    return &APIKeyAuthenticator{storage: store}
}

func (a *APIKeyAuthenticator) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if shouldSkipAuth(r) {
            next.ServeHTTP(w, r)
            return
        }
        
        apiKey := r.Header.Get("X-API-Key")
        if apiKey == "" {
            next.ServeHTTP(w, r)
            return
        }
        
        requestID := middleware.GetRequestID(r.Context())
        
        prefix := extractPrefix(apiKey)
        if prefix == "" {
            log.Printf("[%s] Invalid API key format: no prefix", requestID)
            writeAPIKeyError(w, "Invalid API key format")
            return
        }
        
        hash := sha256.Sum256([]byte(apiKey))
        keyHash := hex.EncodeToString(hash[:])
        
        apiKeyRecord, err := a.storage.GetAPIKeyByHash(r.Context(), keyHash)
        if err != nil {
            log.Printf("[%s] Database error while validating API key: %v", requestID, err)
            writeAPIKeyError(w, "Authentication error")
            return
        }
        
        if apiKeyRecord == nil {
            log.Printf("[%s] API key not found: prefix=%s", requestID, prefix)
            writeAPIKeyError(w, "Invalid API key")
            return
        }
        
        if apiKeyRecord.Revoked {
            log.Printf("[%s] API key revoked: id=%s, user=%s", requestID, apiKeyRecord.ID, apiKeyRecord.UserID)
            writeAPIKeyError(w, "API key has been revoked")
            return
        }
        
        if apiKeyRecord.ExpiresAt != nil && time.Now().After(*apiKeyRecord.ExpiresAt) {
            log.Printf("[%s] API key expired: id=%s, expires_at=%s", requestID, apiKeyRecord.ID, apiKeyRecord.ExpiresAt)
            writeAPIKeyError(w, "API key has expired")
            return
        }
        
        go func() {
            if err := a.storage.UpdateAPIKeyLastUsed(context.Background(), apiKeyRecord.ID); err != nil {
                log.Printf("Failed to update API key last used: %v", err)
            }
        }()
        
        user, err := a.storage.GetUserByID(r.Context(), apiKeyRecord.UserID)
        if err != nil {
            log.Printf("[%s] User not found for API key: %v", requestID, err)
            writeAPIKeyError(w, "Authentication error")
            return
        }
        
        ctx := context.WithValue(r.Context(), APIKeyUserIDKey, apiKeyRecord.UserID)
        ctx = context.WithValue(ctx, APIKeyIDKey, apiKeyRecord.ID)
        ctx = context.WithValue(ctx, KeyRole, user.Role)
        ctx = context.WithValue(ctx, KeyUserID, apiKeyRecord.UserID)
        
        log.Printf("[%s] API key authenticated: key_id=%s, user=%s, name=%s, role=%s", 
            requestID, apiKeyRecord.ID[:8], apiKeyRecord.UserID, apiKeyRecord.Name, user.Role)
        
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func shouldSkipAuth(r *http.Request) bool {
    skipPaths := []string{
        "/health",
        "/api/login",
        "/api/users",
        "/api/auth/keycloak/login",
        "/api/auth/keycloak/callback",
    }
    
    for _, path := range skipPaths {
        if r.URL.Path == path {
            return true
        }
    }
    return false
}

func extractPrefix(apiKey string) string {
    parts := strings.SplitN(apiKey, "_", 3)
    if len(parts) >= 2 {
        return parts[0] + "_" + parts[1]
    }
    return ""
}

func writeAPIKeyError(w http.ResponseWriter, message string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusUnauthorized)
    json.NewEncoder(w).Encode(map[string]string{
        "error": message,
        "code":  "INVALID_API_KEY",
    })
}

func GetAPIKeyUserID(ctx context.Context) string {
    if id, ok := ctx.Value(APIKeyUserIDKey).(string); ok {
        return id
    }
    return ""
}

func GetAPIKeyID(ctx context.Context) string {
    if id, ok := ctx.Value(APIKeyIDKey).(string); ok {
        return id
    }
    return ""
}