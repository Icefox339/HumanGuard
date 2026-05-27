// backend/middleware/metrics.go
package middleware

import (
    "net/http"
    "strconv"
    "time"

    "humanguard/metrics"
)

func MetricsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()

        rw := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}

        next.ServeHTTP(rw, r)

        duration := time.Since(start).Seconds()
        status := strconv.Itoa(rw.statusCode)
        method := r.Method
        
        endpoint := getEndpointFromContext(r)
        if endpoint == "" {
            endpoint = method + ":" + r.URL.Path
        }

        metrics.HTTPRequestsTotal.WithLabelValues(method, endpoint, status).Inc()
        metrics.HTTPRequestDuration.WithLabelValues(method, endpoint).Observe(duration)
    })
}

type statusRecorder struct {
    http.ResponseWriter
    statusCode int
}

func (r *statusRecorder) WriteHeader(code int) {
    r.statusCode = code
    r.ResponseWriter.WriteHeader(code)
}

func getEndpointFromContext(r *http.Request) string {
    if route := r.Context().Value(RouteKey); route != nil {
        if s, ok := route.(string); ok {
            return s
        }
    }
    return ""
}