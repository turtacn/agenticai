// pkg/sandbox/firecracker.go
package sandbox

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/firecracker-microvm/firecracker-go-sdk"
	fcclient "github.com/firecracker-microvm/firecracker-go-sdk"
	"go.opentelemetry.io/otel/trace"
	"github.com/turtacn/agenticai/internal/logger"
)

type firecracker struct {
	ID   string
	mcfg *firecracker.Config
	proc *fcclient.Machine
}

func newFirecrackerRunner(spec *SandboxSpec) Sandbox {
	return &firecracker{ID: spec.ImageRef + "-fc"}
}

func (fc *firecracker) Start(ctx context.Context) error {
	ctx, span := trace.SpanFromContext(ctx).TracerProvider().Tracer("sandbox").Start(ctx, "firecracker.start")
	defer span.End()

	tmp := "/var/run/ag/fc/" + fc.ID
	_ = os.MkdirAll(tmp, 0755)

	kern := "vmlinux"
	img := "rootfs.ext4"
	fc.mcfg = &firecracker.Config{
		SocketPath:        filepath.Join(tmp, "firecracker.sock"),
		KernelImagePath:   kern,
		RootDrive:         firecracker.NewDrive(img, "rw"),
		MachineCfgs:       firecracker.VmmConfig{MemSizeMib: 512, VcpuCount: 1},
		LogLevel:          firecracker.String("ERROR"),
		LogFifo:           filepath.Join(tmp, "log.fifo"),
		ForwardSignals:    []string{"SIGINT", "SIGTERM"},
	}
	m, err := fcclient.NewMachine(ctx, *fc.mcfg, firecracker.WithLoggerFunc(logger.Infof))
	if err != nil {
		return err
	}
	fc.proc = m
	logger.Info(ctx, "firecracker microvm started", "ID", fc.ID)
	return m.Start(ctx)
}

func (fc *firecracker) Kill(ctx context.Context) error {
	return fc.proc.Shutdown(ctx)
}
func (fc *firecracker) Wait(ctx context.Context) error {
	return fc.proc.Wait(ctx)
}
func (fc *firecracker) Info(ctx context.Context) (*Info, error) {
	return &Info{ID: fc.ID, StartTime: time.Now()}, nil
}
//Personal.AI order the ending
