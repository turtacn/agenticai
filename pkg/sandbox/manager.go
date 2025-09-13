// pkg/sandbox/manager.go
package sandbox

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/trace"
)

type Type string

const (
	TypeGvisor      = "gvisor"
	TypeKata        = "kata"
	TypeFirecracker = "firecracker"
)

type SandboxSpec struct {
	Type     Type
	ImageRef string
	Cmd      []string
	Env      map[string]string
	Resource ResourceLimit
	Network  bool
	Volume   string
}

type ResourceLimit struct {
	CPU string
	Mem string
}

type Sandbox interface {
	Start(ctx context.Context) error
	Kill(ctx context.Context) error
	Wait(ctx context.Context) error
	Info(ctx context.Context) (*Info, error)
}

type Info struct {
	ID        string
	Pid       int
	StartTime time.Time
}

type Manager interface {
	Start(ctx context.Context, spec *SandboxSpec) (Sandbox, error)
	Stop(ctx context.Context, id string) error
	List(ctx context.Context) ([]*Info, error)
	Close() error
}

type manager struct {
	factory map[Type]Factory
}

type Factory func(spec *SandboxSpec) Sandbox

func NewManager(ctx context.Context, image string, t Type) (Manager, error) {
	f := map[Type]Factory{
		TypeGvisor:    newGvisorRunner,
		TypeKata:      newKataRunner,
		TypeFirecracker: newFirecrackerRunner,
	}
	if _, ok := f[t]; !ok {
		return nil, fmt.Errorf("unsupported sandbox type %q", t)
	}
	return &manager{factory: f}, nil
}

func (m *manager) Start(ctx context.Context, spec *SandboxSpec) (Sandbox, error) {
	_, span := trace.SpanFromContext(ctx).TracerProvider().Tracer("sandbox").Start(ctx, "Start")
	defer span.End()

	if spec.ImageRef == "" {
		return nil, errors.New("imageRef required")
	}
	// 随机 id
	sb := m.factory[spec.Type](spec)
	if err := sb.Start(ctx); err != nil {
		return nil, err
	}
	return sb, nil
}

func (m *manager) Stop(ctx context.Context, id string) error    { return nil }
func (m *manager) List(ctx context.Context) ([]*Info, error)    { return nil, nil }
func (m *manager) Close() error                                 { return nil }
//Personal.AI order the ending
