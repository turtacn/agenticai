// pkg/agent/runtime.go
package agent

import (
	"context"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/turtacn/agenticai/internal/logger"
	"github.com/turtacn/agenticai/pkg/tools"
	"github.com/turtacn/agenticai/pkg/security"
	"github.com/turtacn/agenticai/pkg/sandbox"
)

// Runtime 智能体运行时实例
type Runtime struct {
	ID           string
	Spec         *AgentSpec            // 来自 ConfigMap/Env
	Listener     net.Listener          // 监听 unix:// 或 tcp://50052
	GRPCSrv      *grpc.Server
	ToolRegistry tools.Registry
	SandboxMgr   sandbox.Manager
	// StorageIface storage.Storage
	// MetricCollector *observability.MetricsCollector
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
}

type AgentSpec struct {
	Image         string
	Envs          map[string]string
	GPU           bool
	WorkloadID    string
	ResourceQuota *ResourceQuota
}

type ResourceQuota struct {
	CPU, Mem string
}

func New(spec *AgentSpec) (*Runtime, error) {
	ctx, cancel := context.WithCancel(context.Background())
	l, err := net.Listen("unix", "/var/run/agenticai/agent.sock")
	if err != nil {
		return nil, err
	}
	srv := grpc.NewServer(
		grpc.UnaryInterceptor(security.SPIFFEInterceptor()),
		grpc.StreamInterceptor(security.SPIFFEStreamInterceptor()),
	)
	rt := &Runtime{
		ID:       os.Getenv("RUNTIME_ID"),
		Spec:     spec,
		Listener: l,
		GRPCSrv:  srv,
		ctx:      ctx,
		cancel:   cancel,
	}
	// 初始化组件
	rt.ToolRegistry = tools.NewInMemRegistry()
	// rt.StorageIface = storage.NewLocalStorage("/tmp/agents", ctx)
	rt.SandboxMgr, _ = sandbox.NewManager(ctx, spec.Image, sandbox.TypeGvisor)
	// rt.MetricCollector = observability.NewLocalMetricsCollector()
	// 注册自服务
	// pb.RegisterAgentServer(srv, rt)
	reflection.Register(srv)
	return rt, nil
}

func (r *Runtime) Start() error {
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		r.GRPCSrv.Serve(r.Listener)
	}()
	r.wg.Add(1)
	go r.keepAliveLoop()
	return nil
}

func (r *Runtime) keepAliveLoop() {
	defer r.wg.Done()
	t := time.NewTicker(15 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-r.ctx.Done():
			return
		case <-t.C:
			if err := r.report(); err != nil {
				logger.Error(r.ctx, "heartbeat: report", zap.Error(err))
			}
		}
	}
}

func (r *Runtime) report() error {
	return nil // TODO: gRPC to controller
}

func (r *Runtime) Stop() {
	r.cancel()
	r.GRPCSrv.GracefulStop()
	r.SandboxMgr.Close()
	r.wg.Wait()
}

func (r *Runtime) awaitShutdown() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c
	r.Stop()
}

/*
type pbAgentServer struct{ *Runtime }
func (r *pbAgentServer) ExecuteTask(ctx context.Context, req *pb.ExecuteTaskRequest) (*pb.ExecuteTaskResponse, error) {
	_, span := otel.Tracer("runtime").Start(ctx, "ExecuteTask")
	defer span.End()
	tr, err := NewExecutor(r.Runtime).ExecuteTask(ctx, req.Task)
	if err != nil {
		return nil, re.ToGRPC(err)
	}
	return &pb.ExecuteTaskResponse{Result: tr}, nil
}
*/
//Personal.AI order the ending
