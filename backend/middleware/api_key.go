// backend/middleware/api_key.go
package middleware

import (
    "context"
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "log"
    "net/http"
    "strings"
    
    "humanguard/storage"
)

type contextKey string

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
        // Skip auth for public endpoints
        if shouldSkipAuth(r) {
            next.ServeHTTP(w, r)
            return
        }
        
        // Check for API key in header
        apiKey := r.Header.Get("X-API-Key")
        if apiKey == "" {
            // No API key, continue to JWT auth
            next.ServeHTTP(w, r)
            return
        }
        
        requestID := GetRequestID(r.Context())
        
        // Extract prefix and hash the key
        prefix := extractPrefix(apiKey)
        if prefix == "" {
            log.Printf("[%s] Invalid API key format: no prefix", requestID)
            writeAPIKeyError(w, "Invalid API key format")
            return
        }
        
        // Hash the key for lookup
        hash := sha256.Sum256([]byte(apiKey))
        keyHash := hex.EncodeToString(hash[:])
        
        // Validate API key
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
        
        // Check if revoked
        if apiKeyRecord.Revoked {
            log.Printf("[%s] API key revoked: id=%s, user=%s", requestID, apiKeyRecord.ID, apiKeyRecord.UserID)
            writeAPIKeyError(w, "API key has been revoked")
            return
        }
        
        // Check expiration
        if apiKeyRecord.ExpiresAt != nil && time.Now().After(*apiKeyRecord.ExpiresAt) {
            log.Printf("[%s] API key expired: id=%s, expires_at=%s", 
                requestID, apiKeyRecord.ID, apiKeyRecord.ExpiresAt)
            writeAPIKeyError(w, "API key has expired")
            return
        }
        
        // Update last used timestamp (async, don't block)
        go func() {
            if err := a.storage.UpdateAPIKeyLastUsed(context.Background(), apiKeyRecord.ID); err != nil {
                log.Printf("Failed to update API key last used: %v", err)
            }
        }()
        
        // Add user info to context
        ctx := context.WithValue(r.Context(), APIKeyUserIDKey, apiKeyRecord.UserID)
        ctx = context.WithValue(ctx, APIKeyIDKey, apiKeyRecord.ID)
        
        log.Printf("[%s] API key authenticated: key_id=%s, user=%s, name=%s", 
            requestID, apiKeyRecord.ID[:8], apiKeyRecord.UserID, apiKeyRecord.Name)
        
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
    // Format: hg_v1_xxxxxxxxxxxxxxxxxxxx
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