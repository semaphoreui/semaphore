package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics owns Semaphore's Prometheus registry. It is constructed once and
// injected wherever metrics need to be exposed or recorded, rather than
// relying on the client library's global DefaultRegisterer.
type Metrics struct {
	registry *prometheus.Registry
	handler  http.Handler
}

func NewMetrics() *Metrics {
	registry := prometheus.NewRegistry()

	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	return &Metrics{
		registry: registry,
		handler:  promhttp.HandlerFor(registry, promhttp.HandlerOpts{}),
	}
}

// ServeHTTP makes Metrics itself an http.Handler, serving the current
// metrics in the Prometheus text exposition format.
func (m *Metrics) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.handler.ServeHTTP(w, r)
}
