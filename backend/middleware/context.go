// backend/middleware/context.go
package middleware

type contextKey string

const (
    RequestIDKey contextKey = "request_id"
    
    RouteKey contextKey = "route"
)