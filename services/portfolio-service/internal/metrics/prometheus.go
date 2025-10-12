package metrics

import (
	"net/http"
	"sort"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type PrometheusRecorder struct {
	namespace string
	mu        sync.RWMutex
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

func (p *PrometheusRecorder) IncCounter(name string, labels map[string]string) {
	lbls := sortedKeys(labels)
	key := name + "|" + strings.Join(lbls, ",")
	p.mu.RLock()
	cv, ok := p.counters[key]
	p.mu.RUnlock()
	if !ok {
		p.mu.Lock()
		if cv, ok = p.counters[key]; !ok {
			cv = prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: p.namespace,
				Name:      sanitize(name),
				Help:      name,
			}, lbls)
			prometheus.MustRegister(cv)
			p.counters[key] = cv
		}
		p.mu.Unlock()
	}
	cv.With(labels).Inc()
}

func (p *PrometheusRecorder) ObserveDuration(name string, seconds float64, labels map[string]string) {
	lbls := sortedKeys(labels)
	key := name + "|" + strings.Join(lbls, ",")
	p.mu.RLock()
	hv, ok := p.hists[key]
	p.mu.RUnlock()
	if !ok {
		p.mu.Lock()
		if hv, ok = p.hists[key]; !ok {
			hv = prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: p.namespace,
				Name:      sanitize(name),
				Help:      name,
				Buckets:   prometheus.DefBuckets,
			}, lbls)
			prometheus.MustRegister(hv)
			p.hists[key] = hv
		}
		p.mu.Unlock()
	}
	hv.With(labels).Observe(seconds)
}

func Handler() http.Handler {
	return promhttp.Handler()
}

func sortedKeys(m map[string]string) []string {
	if len(m) == 0 {
		return nil
	}
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
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
