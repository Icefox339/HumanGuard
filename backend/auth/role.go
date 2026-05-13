package auth

import (
    "net/http"
)

func RequireRole(allowedRoles ...string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            userRole := GetRole(r.Context())
            
            for _, role := range allowedRoles {
                if userRole == role {
                    next.ServeHTTP(w, r)
                    return
                }
            }
            
            http.Error(w, "Forbidden: insufficient permissions", http.StatusForbidden)
        })
    }
}

func RequireAdmin() func(http.Handler) http.Handler {
    return RequireRole("admin")
}
