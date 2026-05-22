// backend/middleware/csp.go
package middleware

import (
	"net/http"
	"os"
	"strings"
)

func CSPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		env := os.Getenv("ENV")
		
		var csp string
		
		if env == "development" {
			csp = strings.Join([]string{
				"default-src 'self'",
				"script-src 'self' 'unsafe-inline' 'unsafe-eval' http://localhost:*",
				"style-src 'self' 'unsafe-inline'",
				"img-src 'self' data: https:",
				"font-src 'self'",
				"connect-src 'self' http://localhost:* ws://localhost:*",
				"frame-ancestors 'none'",
				"base-uri 'self'",
				"form-action 'self'",
			}, "; ")
		} else {
			csp = strings.Join([]string{
				"default-src 'self'",
				"script-src 'self'",
				"style-src 'self'",
				"img-src 'self' data:",
				"font-src 'self'",
				"connect-src 'self'",
				"frame-ancestors 'none'",
				"base-uri 'self'",
				"form-action 'self'",
				"upgrade-insecure-requests",
			}, "; ")
		}
		
		w.Header().Set("Content-Security-Policy", csp)
		next.ServeHTTP(w, r)
	})
}