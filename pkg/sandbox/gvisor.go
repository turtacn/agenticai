// pkg/sandbox/gvisor.go
package sandbox

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"github.com/turtacn/agenticai/internal/logger"
)

type gvisor struct {
	ID   string
	dir  string
	bin  string
	cmd  *exec.Cmd
}

const runsc = "/usr/local/bin/runsc"

func newGvisorRunner(spec *SandboxSpec) Sandbox {
	return &gvisor{ID: spec.ImageRef + "-gvisor", bin: runsc}
}

func (g *gvisor) Start(ctx context.Context) error {
	ctx, span := trace.SpanFromContext(ctx).TracerProvider().Tracer("sandbox").Start(ctx, "gvisor.start")
	defer span.End()
	span.SetAttributes(attribute.String("sandbox.id", g.ID))

	g.dir = "/var/run/ag/sb/" + g.ID
	_ = os.MkdirAll(g.dir, 0755)
	if err := g.createBundle(); err != nil {
		return err
	}
	g.cmd = exec.CommandContext(ctx, g.bin, "create",
		"--bundle", g.dir, "--pid-file", "pid", g.ID)
	g.cmd.Dir = g.dir
	if err := g.cmd.Run(); err != nil {
		return err
	}
	logger.Info(ctx, "gvisor started", zap.String("ID", g.ID))
	return nil
}

func (g *gvisor) Kill(_ context.Context) error {
	return exec.Command(g.bin, "kill", g.ID).Run()
}

func (g *gvisor) Wait(_ context.Context) error {
	return exec.Command(g.bin, "wait", g.ID).Run()
}

func (g *gvisor) Info(_ context.Context) (*Info, error) {
	pidBytes, _ := os.ReadFile(filepath.Join(g.dir, "pid"))
	pid := 0
	fmt.Sscan(string(pidBytes), &pid)
	return &Info{ID: g.ID, Pid: pid, StartTime: time.Now()}, nil
}

func (g *gvisor) createBundle() error {
	config := fmt.Sprintf(`{"ociVersion":"1.0.0","process":{"cwd":"/","env":[],"args":["/bin/echo","hello"]},"linux,omitempty":{}}`)
	return os.WriteFile(filepath.Join(g.dir, "config.json"), []byte(config), 0644)
}
//Personal.AI order the ending
