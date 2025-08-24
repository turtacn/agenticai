// pkg/observability/metrics.go
package observability

import (
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var once sync.Once

// MustInit registry & default collectors
func MustInit(svc string) {
	once.Do(func() {
		reg := prometheus.NewRegistry()
		reg.MustRegister(
			collectors.NewGoCollector(),
			collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		)
		prometheus.DefaultRegisterer = reg
	})
}

// Expose handler
func Handler() http.Handler {
	return promhttp.Handler()
}
//Personal.AI order the ending
