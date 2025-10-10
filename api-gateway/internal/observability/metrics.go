package observability

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type HTTPMetrics struct {
	RequestsTotal       *prometheus.CounterVec
	RequestsDuration    *prometheus.HistogramVec
	InFlight            prometheus.Gauge
	RateLimitAllowed    *prometheus.CounterVec
	RateLimitBlocked    *prometheus.CounterVec
}

func InitMetricsRegistry(cfg interface{}) *prometheus.Registry {
	reg := prometheus.NewRegistry()
	return reg
}

func NewHTTPMetrics(reg *prometheus.Registry, namespace string) *HTTPMetrics {
	m := &HTTPMetrics{
		RequestsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests",
		}, []string{"method", "route", "status"}),
		RequestsDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request duration in seconds",
			Buckets:   prometheus.ExponentialBuckets(0.005, 2, 10),
		}, []string{"method", "route", "status"}),
		InFlight: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "http_requests_in_flight",
			Help:      "Current number of in-flight requests",
		}),
		RateLimitAllowed: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "rate_limit_allowed_total",
			Help:      "Total number of requests allowed by rate limiter",
		}, []string{"route"}),
		RateLimitBlocked: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "rate_limit_blocked_total",
			Help:      "Total number of requests blocked by rate limiter",
		}, []string{"route"}),
	}
	reg.MustRegister(m.RequestsTotal, m.RequestsDuration, m.InFlight, m.RateLimitAllowed, m.RateLimitBlocked)
	return m
}

func Observe(m *HTTPMetrics, method, route, status string, start time.Time) {
	if m == nil {
		return
	}
	m.InFlight.Dec()
	m.RequestsTotal.WithLabelValues(method, route, status).Inc()
	m.RequestsDuration.WithLabelValues(method, route, status).Observe(time.Since(start).Seconds())
}
