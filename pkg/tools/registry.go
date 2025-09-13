// pkg/tools/registry.go
package tools

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/turtacn/agenticai/internal/errors"
	"github.com/turtacn/agenticai/internal/logger"
	"github.com/turtacn/agenticai/pkg/apis"
)

type Registry interface {
	Register(ctx context.Context, spec *apis.ToolSpec) error
	Deregister(ctx context.Context, toolID string) error
	List(ctx context.Context, filter *apis.ToolFilter) ([]*apis.Metadata, error)
	Get(ctx context.Context, toolID string) (*apis.Metadata, error)
}

// ------------------ 内存实现 ------------------

const defaultTTL = 5 * time.Minute

type item struct {
	meta      apis.Metadata
	spec      apis.ToolSpec
	expiresAt time.Time
}

type inMemRegistry struct {
	mu   sync.RWMutex
	data map[string]*item
	ttl  time.Duration

	registeredTotal prometheus.Counter
	deregistered    prometheus.Counter
	activeGauge     prometheus.Gauge
}

func NewInMemRegistry() Registry {
	return &inMemRegistry{
		data:     make(map[string]*item),
		ttl:      defaultTTL,
		registeredTotal: promauto.NewCounter(prometheus.CounterOpts{Name: "agenticai_tools_registered_total", Help: "total tools ever registered"}),
		deregistered:   promauto.NewCounter(prometheus.CounterOpts{Name: "agenticai_tools_deregistered_total", Help: "total tools deregistered"}),
		activeGauge:    promauto.NewGauge(prometheus.GaugeOpts{Name: "agenticai_tools_active", Help: "currently active tools"}),
	}
}

func (r *inMemRegistry) Register(ctx context.Context, spec *apis.ToolSpec) error {
	ctx, span := trace.SpanFromContext(ctx).TracerProvider().Tracer("tools").Start(ctx, "Registry.Register")
	defer span.End()

	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	id := spec.ID
	r.data[id] = &item{
		meta:      apis.Metadata{ID: id, Name: spec.Name, Version: spec.Version, Digest: spec.Digest},
		spec:      *spec,
		expiresAt: now.Add(r.ttl),
	}
	r.registeredTotal.Inc()
	r.activeGauge.Set(float64(len(r.data)))
	logger.Info(ctx, "tool registered", zap.String("id", id), zap.String("name", spec.Name))
	return nil
}

func (r *inMemRegistry) Deregister(ctx context.Context, toolID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.data[toolID]; !ok {
		return errors.E(errors.KindNotFound, fmt.Sprintf("tool %s not found", toolID))
	}
	delete(r.data, toolID)
	r.deregistered.Inc()
	r.activeGauge.Set(float64(len(r.data)))
	logger.Info(ctx, "tool deregistered", zap.String("toolID", toolID))
	return nil
}

func (r *inMemRegistry) List(ctx context.Context, filter *apis.ToolFilter) ([]*apis.Metadata, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]*apis.Metadata, 0, len(r.data))
	for _, v := range r.data {
		if filter == nil {
			out = append(out, &v.meta)
			continue
		}
		if filter.Name != "" && v.meta.Name != filter.Name {
			continue
		}
		out = append(out, &v.meta)
	}
	return out, nil
}

func (r *inMemRegistry) Get(ctx context.Context, toolID string) (*apis.Metadata, error) {
	r.mu.RLock()
	item, ok := r.data[toolID]
	r.mu.RUnlock()
	if !ok {
		return nil, errors.E(errors.KindNotFound, fmt.Sprintf("tool %s not found", toolID))
	}
	return &item.meta, nil
}
//Personal.AI order the ending
