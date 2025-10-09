package observability

import (
	"github.com/prometheus/client_golang/prometheus"
)

func InitMetricsRegistry(cfg interface{}) *prometheus.Registry {
	reg := prometheus.NewRegistry()
	return reg
}
