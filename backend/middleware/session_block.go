package middleware

import (
	"context"
	"humanguard/storage"
	"net/http"
	"strings"
)

type sessionContextKey string

const KeySessionID sessionContextKey = "session_id"

func SessionBlockMiddleware(store storage.Storage) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isPublicEndpoint(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			sessionID := r.Header.Get("X-Session-ID")
			if sessionID == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"error":"X-Session-ID header is required for this endpoint"}`))
				return
			}

			session, err := store.GetSession(r.Context(), sessionID)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":"invalid or expired session"}`))
				return
			}

			if session.IsBlocked {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(`{"error":"session is blocked","reason":"suspicious activity detected"}`))
				return
			}

			ctx := context.WithValue(r.Context(), KeySessionID, sessionID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func isPublicEndpoint(path string) bool {
	publicPaths := []string{
		"/health",
		"/api/login",
		"/api/users",
		"/api/auth/keycloak/login",
		"/api/auth/keycloak/callback",
		"/api/files/share/",
		"/api/sessions",
	}

	for _, p := range publicPaths {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}
