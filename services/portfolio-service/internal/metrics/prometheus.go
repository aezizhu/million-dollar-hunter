package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type PrometheusRecorder struct {
	namespace string
	counters  map[string]*prometheus.CounterVec
	hists     map[string]*prometheus.HistogramVec
}

func NewPrometheusRecorder(namespace string) *PrometheusRecorder {
	return &PrometheusRecorder{
		namespace: namespace,
		counters:  make(map[string]*prometheus.CounterVec),
		hists:     make(map[string]*prometheus.HistogramVec),
	}
}

func (p *PrometheusRecorder) ensureCounter(name string, labels map[string]string) *prometheus.CounterVec {
	lbls := keys(labels)
	key := name + "|" + joinKeys(lbls)
	if c, ok := p.counters[key]; ok {
		return c
	}
	cv := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: p.namespace,
		Name:      sanitize(name),
		Help:      name,
	}, lbls)
	prometheus.MustRegister(cv)
	p.counters[key] = cv
	return cv
}

func (p *PrometheusRecorder) ensureHist(name string, labels map[string]string) *prometheus.HistogramVec {
	lbls := keys(labels)
	key := name + "|" + joinKeys(lbls)
	if h, ok := p.hists[key]; ok {
		return h
	}
	hv := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: p.namespace,
		Name:      sanitize(name),
		Help:      name,
		Buckets: prometheus.DefBuckets,
	}, lbls)
	prometheus.MustRegister(hv)
	p.hists[key] = hv
	return hv
}

func (p *PrometheusRecorder) IncCounter(name string, labels map[string]string) {
	cv := p.ensureCounter(name, labels)
	cv.With(labels).Inc()
}

func (p *PrometheusRecorder) ObserveDuration(name string, seconds float64, labels map[string]string) {
	hv := p.ensureHist(name, labels)
	hv.With(labels).Observe(seconds)
}

func Handler() http.Handler {
	return promhttp.Handler()
}

func keys(m map[string]string) []string {
	if len(m) == 0 {
		return nil
	}
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func joinKeys(arr []string) string {
	if len(arr) == 0 {
		return ""
	}
	sep := ","
	out := arr[0]
	for i := 1; i < len(arr); i++ {
		out += sep + arr[i]
	}
	return out
}

func sanitize(s string) string {
	b := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if (ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '_' || ch == ':' {
			b = append(b, ch)
		} else if ch == '.' || ch == '-' {
			b = append(b, '_')
		}
	}
	if len(b) == 0 {
		return "metric"
	}
	first := b[0]
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z') || first == '_' || first == ':') {
		return "_" + string(b)
	}
	return string(b)
}
