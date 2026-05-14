package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    HTTPRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )

    HTTPRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "endpoint"},
    )

    ActiveSessions = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "active_sessions_total",
            Help: "Total number of active visitor sessions",
        },
    )

    BlockedRequests = promauto.NewCounter(
        prometheus.CounterOpts{
            Name: "blocked_requests_total",
            Help: "Total number of blocked requests",
        },
    )

    CaptchaRequests = promauto.NewCounter(
        prometheus.CounterOpts{
            Name: "captcha_requests_total",
            Help: "Total number of requests that required captcha",
        },
    )

    RateLimitExceeded = promauto.NewCounter(
        prometheus.CounterOpts{
            Name: "rate_limit_exceeded_total",
            Help: "Total number of rate limit exceeded (HTTP 429)",
        },
    )

    AverageRiskScore = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "risk_score_average",
            Help: "Average risk score across all sessions",
        },
    )

    HighRiskSessions = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "high_risk_sessions_total",
            Help: "Number of sessions with risk score >= 80",
        },
    )
)