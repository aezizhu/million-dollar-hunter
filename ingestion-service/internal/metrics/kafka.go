package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type KafkaMetrics struct {
	PublishTotal      *prometheus.CounterVec
	PublishDuration   *prometheus.HistogramVec
	PublishErrors     *prometheus.CounterVec
	MessageSize       *prometheus.HistogramVec
	ConnectionErrors  prometheus.Counter
	Connected         prometheus.Gauge
}

func NewKafkaMetrics(reg *prometheus.Registry, namespace string) *KafkaMetrics {
	m := &KafkaMetrics{
		PublishTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "kafka_publish_total",
			Help:      "Total number of Kafka publish attempts",
		}, []string{"topic", "status"}),
		PublishDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "kafka_publish_duration_seconds",
			Help:      "Kafka publish latency in seconds",
			Buckets:   prometheus.ExponentialBuckets(0.001, 2, 10), // 1ms to ~1s
		}, []string{"topic"}),
		PublishErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "kafka_publish_errors_total",
			Help:      "Total number of Kafka publish errors",
		}, []string{"topic", "error_type"}),
		MessageSize: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "kafka_message_size_bytes",
			Help:      "Kafka message size in bytes",
			Buckets:   prometheus.ExponentialBuckets(100, 2, 10), // 100B to ~100KB
		}, []string{"topic"}),
		ConnectionErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "kafka_connection_errors_total",
			Help:      "Total number of Kafka connection errors",
		}),
		Connected: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "kafka_connected",
			Help:      "1 if connected to Kafka, 0 otherwise",
		}),
	}

	reg.MustRegister(
		m.PublishTotal,
		m.PublishDuration,
		m.PublishErrors,
		m.MessageSize,
		m.ConnectionErrors,
		m.Connected,
	)

	return m
}

func (m *KafkaMetrics) ObservePublish(topic string, err error, messageSize int, start time.Time) {
	if m == nil {
		return
	}

	duration := time.Since(start).Seconds()
	m.PublishDuration.WithLabelValues(topic).Observe(duration)
	m.MessageSize.WithLabelValues(topic).Observe(float64(messageSize))

	if err != nil {
		m.PublishTotal.WithLabelValues(topic, "error").Inc()
		errorType := "unknown"
		if err.Error() != "" {
			errorType = err.Error()[:min(50, len(err.Error()))] // Limit error type length
		}
		m.PublishErrors.WithLabelValues(topic, errorType).Inc()
	} else {
		m.PublishTotal.WithLabelValues(topic, "success").Inc()
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
