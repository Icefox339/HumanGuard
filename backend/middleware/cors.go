package middleware

import (
	"net/http"
	"os"
	"strings"
)

// APICORSMiddleware для API эндпоинтов (админка, фронтенд)
func APICORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/check") ||
			strings.HasPrefix(r.URL.Path, "/api/behavior") {
			next.ServeHTTP(w, r)
			return
		}

		allowedOriginsStr := os.Getenv("API_CORS_ORIGINS")
		allowedOrigins := strings.Split(allowedOriginsStr, ",")

		if len(allowedOrigins) == 0 || (len(allowedOrigins) == 1 && allowedOrigins[0] == "") {
			allowedOrigins = []string{"http://localhost:5173", "http://localhost:3000", "http://localhost:80"}
		}

		origin := r.Header.Get("Origin")
		for _, allowed := range allowedOrigins {
			if strings.TrimSpace(allowed) == origin {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key, X-CSRF-Token")
				w.Header().Set("Access-Control-Expose-Headers", "X-Request-ID, X-RateLimit-Limit, X-RateLimit-Remaining")
				break
			}
		}

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
