// cmd/agent-runtime/main.go
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/turtacn/agenticai/internal/config"
	"github.com/turtacn/agenticai/pkg/agent"
	"github.com/turtacn/agenticai/pkg/observability"
	"github.com/turtacn/agenticai/pkg/sandbox"
	"github.com/turtacn/agenticai/pkg/tools"
)

const ServiceName = "agent-runtime"

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg, err := loadRuntimeConfig()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	// 可观测性
	shut, err := observability.Init(ctx, cfg.Observability, ServiceName)
	if err != nil {
		log.Fatalf("observability: %v", err)
	}
	defer shut()

	// 沙箱
	sbMgr, err := sandbox.NewManager(cfg.Sandbox)
	if err != nil {
		log.Fatalf("sandbox: %v", err)
	}
	defer sbMgr.Shutdown(context.Background())

	// 工具网关
	tgw, err := tools.NewGateway(cfg.Tools)
	if err != nil {
		log.Fatalf("tool gateway: %v", err)
	}
	defer tgw.Close()

	// 运行时实例
	r := agent.NewRuntime(cfg, sbMgr, tgw)
	go func() {
		if err := r.Run(ctx); err != nil {
			log.Fatalf("runtime: %v", err)
		}
	}()

	// 健康探针
	go startProbe()
	<-ctx.Done()
	log.Println("🛑 agent-runtime stopped")
}

// ------------------- 配置组装 -------------------
func loadRuntimeConfig() (*agent.Config, error) {
	cfg := &agent.Config{}
	cfg.Observability.Port = mustEnv("OTEL_PORT", "8081")
	cfg.Sandbox.Type = mustEnv("SANDBOX_TYPE", "gvisor")
	cfg.Tools.GatewayAddr = mustEnv("TOOL_GATEWAY", "")
	return cfg, nil
}

func mustEnv(k, defVal string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return defVal
}

// ------------------- 探针 -------------------
func startProbe() {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("ok")) })
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("ready")) })
	mux.Handle("/metrics", observability.MustHandler())
	s := &http.Server{Addr: ":8080", Handler: mux}
	log.Println("🔥 agent-runtime probe on :8080")
	_ = s.ListenAndServe()
}
//Personal.AI order the ending
