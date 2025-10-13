package observability

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type HTTPMetrics struct {
	RequestsTotal           *prometheus.CounterVec
	RequestsDuration        *prometheus.HistogramVec
	InFlight                prometheus.Gauge
	RateLimitAllowed        *prometheus.CounterVec
	RateLimitBlocked        *prometheus.CounterVec
	ViolationsByIP          prometheus.Counter
	HierarchicalDenials     *prometheus.CounterVec
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
		}, []string{"route", "dimension"}),
		RateLimitBlocked: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "rate_limit_blocked_total",
			Help:      "Total number of requests blocked by rate limiter",
		}, []string{"route", "dimension"}),
		ViolationsByIP: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "rate_limit_violations_by_ip",
			Help:      "Count of IP-based rate limit violations",
		}),
		HierarchicalDenials: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "rate_limit_hierarchical_denials_total",
			Help:      "Total number of hierarchical denials by dimension",
		}, []string{"route", "dimension"}),
	}
	reg.MustRegister(m.RequestsTotal, m.RequestsDuration, m.InFlight, m.RateLimitAllowed, m.RateLimitBlocked, m.ViolationsByIP, m.HierarchicalDenials)
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

type AuthGRPCMetrics struct {
	total   *prometheus.CounterVec
	latency prometheus.Observer
}

func NewAuthGRPCMetrics(reg *prometheus.Registry, namespace string) *AuthGRPCMetrics {
	total := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "auth_grpc_validation_total",
		Help:      "Total number of gRPC auth validations by outcome",
	}, []string{"outcome"})
	latency := prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "auth_grpc_validation_duration_seconds",
		Help:      "Duration of gRPC auth validation calls",
		Buckets:   prometheus.DefBuckets,
	})
	reg.MustRegister(total, latency)
	return &AuthGRPCMetrics{total: total, latency: latency}
}

func (m *AuthGRPCMetrics) Inc(outcome string) {
	m.total.WithLabelValues(outcome).Inc()
}

func (m *AuthGRPCMetrics) Time(start time.Time) {
	m.latency.Observe(time.Since(start).Seconds())
}
