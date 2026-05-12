// backend/auth/middleware.go
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
	KeyUserID contextKey = "userID"
	KeyRole   contextKey = "role"
)

func AuthMiddleware(jwt *JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := middleware.GetRequestID(r.Context())

			apiKeyUserID := middleware.GetAPIKeyUserID(r.Context())
			isAPIKeyAuth := apiKeyUserID != ""

			adminEndpoints := []string{
				"/api/keys",
			}

			for _, endpoint := range adminEndpoints {
				if strings.HasPrefix(r.URL.Path, endpoint) {
					if isAPIKeyAuth {
						log.Printf("[%s] Admin endpoint %s rejected: API key not allowed", requestID, r.URL.Path)
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusForbidden)
						w.Write([]byte(`{"error":"API keys cannot access admin endpoints. Use JWT authentication."}`))
						return
					}
					break
				}
			}

			userID := GetUserID(r.Context())
			if userID == "" && apiKeyUserID != "" {
				userID = apiKeyUserID
			}

			if userID != "" {
				ctx := context.WithValue(r.Context(), KeyUserID, userID)
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
			userID, role, err := jwt.ValidateToken(token)
			if err != nil {
				log.Printf("[%s] Auth failed: invalid token - %v", requestID, err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":"invalid token"}`))
				return
			}

			ctx := context.WithValue(r.Context(), KeyUserID, userID)
			ctx = context.WithValue(ctx, KeyRole, role)
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