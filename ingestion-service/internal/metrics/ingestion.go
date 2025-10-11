package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type IngestionMetrics struct {
	JobsEnqueued       prometheus.Counter
	JobsProcessed      *prometheus.CounterVec
	JobDuration        *prometheus.HistogramVec
	JobQueueDepth      prometheus.Gauge
	TransformerErrors  *prometheus.CounterVec
	APICallsTotal      *prometheus.CounterVec
	APICallDuration    *prometheus.HistogramVec
	CircuitBreakerOpen *prometheus.GaugeVec
}

func NewIngestionMetrics(reg *prometheus.Registry, namespace string) *IngestionMetrics {
	m := &IngestionMetrics{
		JobsEnqueued: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "ingestion_jobs_enqueued_total",
			Help:      "Total number of ingestion jobs enqueued",
		}),
		JobsProcessed: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "ingestion_jobs_processed_total",
			Help:      "Total number of ingestion jobs processed",
		}, []string{"chain", "status"}),
		JobDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "ingestion_job_duration_seconds",
			Help:      "Ingestion job processing duration in seconds",
			Buckets:   prometheus.ExponentialBuckets(0.1, 2, 10), // 100ms to ~100s
		}, []string{"chain"}),
		JobQueueDepth: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "ingestion_job_queue_depth",
			Help:      "Current number of jobs in the queue",
		}),
		TransformerErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "ingestion_transformer_errors_total",
			Help:      "Total number of transformer errors",
		}, []string{"transformer", "error_type"}),
		APICallsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "ingestion_api_calls_total",
			Help:      "Total number of external API calls",
		}, []string{"provider", "endpoint", "status"}),
		APICallDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "ingestion_api_call_duration_seconds",
			Help:      "External API call duration in seconds",
			Buckets:   prometheus.ExponentialBuckets(0.01, 2, 10), // 10ms to ~10s
		}, []string{"provider", "endpoint"}),
		CircuitBreakerOpen: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "ingestion_circuit_breaker_open",
			Help:      "1 if circuit breaker is open, 0 otherwise",
		}, []string{"provider"}),
	}

	reg.MustRegister(
		m.JobsEnqueued,
		m.JobsProcessed,
		m.JobDuration,
		m.JobQueueDepth,
		m.TransformerErrors,
		m.APICallsTotal,
		m.APICallDuration,
		m.CircuitBreakerOpen,
	)

	return m
}

func (m *IngestionMetrics) ObserveJob(chain, status string, start time.Time) {
	if m == nil {
		return
	}

	duration := time.Since(start).Seconds()
	m.JobsProcessed.WithLabelValues(chain, status).Inc()
	m.JobDuration.WithLabelValues(chain).Observe(duration)
}

func (m *IngestionMetrics) ObserveAPICall(provider, endpoint, status string, start time.Time) {
	if m == nil {
		return
	}

	duration := time.Since(start).Seconds()
	m.APICallsTotal.WithLabelValues(provider, endpoint, status).Inc()
	m.APICallDuration.WithLabelValues(provider, endpoint).Observe(duration)
}

func (m *IngestionMetrics) SetCircuitBreakerState(provider string, open bool) {
	if m == nil {
		return
	}

	value := 0.0
	if open {
		value = 1.0
	}
	m.CircuitBreakerOpen.WithLabelValues(provider).Set(value)
}
