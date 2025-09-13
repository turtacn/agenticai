// pkg/sandbox/firecracker.go
package sandbox

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/firecracker-microvm/firecracker-go-sdk"
	models "github.com/firecracker-microvm/firecracker-go-sdk/client/models"
	"github.com/sirupsen/logrus"
	"github.com/turtacn/agenticai/internal/logger"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type firecrackerRunner struct {
	ID   string
	mcfg *firecracker.Config
	proc *firecracker.Machine
}

func newFirecrackerRunner(spec *SandboxSpec) Sandbox {
	return &firecrackerRunner{ID: spec.ImageRef + "-fc"}
}

func (fc *firecrackerRunner) Start(ctx context.Context) error {
	_, span := trace.SpanFromContext(ctx).TracerProvider().Tracer("sandbox").Start(ctx, "firecracker.start")
	defer span.End()

	tmp := "/var/run/ag/fc/" + fc.ID
	_ = os.MkdirAll(tmp, 0755)

	// NOTE: These paths are placeholders and will need to be configured correctly.
	kernelImagePath := "vmlinux.bin"
	rootfsPath := "rootfs.ext4"

	fc.mcfg = &firecracker.Config{
		SocketPath:      filepath.Join(tmp, "firecracker.sock"),
		KernelImagePath: kernelImagePath,
		Drives:          firecracker.NewDrivesBuilder(rootfsPath).Build(),
		MachineCfg: models.MachineConfiguration{
			VcpuCount:  firecracker.Int64(1),
			MemSizeMib: firecracker.Int64(512),
		},
		LogLevel: "Error",
	}

	log := logger.WithCtx(ctx)
	fcLogger := logrus.New()
	fcLogger.SetOutput(os.Stdout)
	entry := logrus.NewEntry(fcLogger)

	m, err := firecracker.NewMachine(ctx, *fc.mcfg, firecracker.WithLogger(entry))
	if err != nil {
		log.Error("Failed to create firecracker machine", zap.Error(err))
		return err
	}
	fc.proc = m
	log.Info("firecracker microvm starting", zap.String("ID", fc.ID))
	return m.Start(ctx)
}

func (fc *firecrackerRunner) Kill(ctx context.Context) error {
	return fc.proc.StopVMM()
}
func (fc *firecrackerRunner) Wait(ctx context.Context) error {
	return fc.proc.Wait(ctx)
}
func (fc *firecrackerRunner) Info(ctx context.Context) (*Info, error) {
	// TODO: Populate with actual info from the machine
	return &Info{ID: fc.ID, StartTime: time.Now()}, nil
}
//Personal.AI order the ending
