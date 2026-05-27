// backend/metrics/metrics.go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // Основные метрики для HTTP запросов
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
            Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
        },
        []string{"method", "endpoint"},
    )

    // Метрики сессий и рисков
    ActiveSessions = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "active_sessions_total",
            Help: "Total number of active visitor sessions",
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
    
    // Действия блокировки с детализацией
    BlockedActions = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "blocked_actions_total",
            Help: "Number of actions taken (block, captcha) based on risk score",
        },
        []string{"action"},
    )
)