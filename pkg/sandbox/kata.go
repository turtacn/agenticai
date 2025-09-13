// pkg/sandbox/kata.go
package sandbox

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"github.com/turtacn/agenticai/internal/logger"
)

type kata struct {
	ID   string
	spec *SandboxSpec
	cmd  *exec.Cmd
}

func newKataRunner(spec *SandboxSpec) Sandbox {
	return &kata{ID: spec.ImageRef + "-kata", spec: spec}
}

func (k *kata) Start(ctx context.Context) error {
	ctx, span := trace.SpanFromContext(ctx).TracerProvider().Tracer("sandbox").Start(ctx, "kata.start")
	defer span.End()

	dir := "/var/run/ag/kata/" + k.ID
	_ = os.MkdirAll(dir, 0755)
	config := k.generateSpec()
	cfgFile := dir + "/config.json"
	if err := os.WriteFile(cfgFile, []byte(config), 0644); err != nil {
		return err
	}
	k.cmd = exec.CommandContext(ctx, "kata-runtime", "--config", cfgFile, "create", "--bundle", dir, k.ID)
	if err := k.cmd.Start(); err != nil {
		return err
	}
	logger.Info(ctx, "kata started", zap.String("ID", k.ID))
	return nil
}

func (k *kata) Kill(_ context.Context) error {
	return exec.Command("kata-runtime", "kill", k.ID).Run()
}
func (k *kata) Wait(_ context.Context) error {
	return exec.Command("kata-runtime", "wait", k.ID).Run()
}
func (k *kata) Info(_ context.Context) (*Info, error) {
	return &Info{ID: k.ID, StartTime: time.Now()}, nil
}

func (k *kata) generateSpec() string {
	return fmt.Sprintf(`{
		"ociVersion":"1.0.0",
		"process":{"cwd":"/","args":["/bin/sh"]},
		"linux":{"resources":{"memory":{"limit":1000000000},"cpu":{"quota":100000}},
		         "annotations":{"com.io.kata.vcpus":"1","com.io.kata.memory":"1G"}}
	}`)
}
//Personal.AI order the ending
