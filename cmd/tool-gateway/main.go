// cmd/tool-gateway/main.go
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
	"github.com/turtacn/agenticai/pkg/gateway"
	"github.com/turtacn/agenticai/pkg/observability"
)

const ServiceName = "tool-gateway"

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// 1. 加载配置
	cfg, err := loadGwConfig()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	// 2. 初始化可观测性
	shutdownObs, err := observability.Init(ctx, cfg.Observability, ServiceName)
	if err != nil {
		log.Fatalf("observability: %v", err)
	}
	defer shutdownObs()

	// 3. 启动 HTTP 网关
	router := gateway.NewRouter(cfg)
	srv := &http.Server{
		Addr:         cfg.ListenAddr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// 4. 监听 & 热重载文件
	if err := gateway.StartFileWatcher(cfg.ToolsConfigPath, router); err != nil {
		log.Fatalf("watch config: %v", err)
	}

	go func() {
		log.Printf("🚀 tool-gateway listening on %s", cfg.ListenAddr)
		if err := srv.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("gateway server: %v", err)
		}
	}()

	// 5. 探针
	go startProbe(cfg)

	// 6. 优雅关闭
	<-ctx.Done()
	log.Println("💤 shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("graceful shutdown timedout: %v", err)
	}
	log.Println("✅ tool-gateway exited")
}

// ------------------- 配置函数 -------------------
func loadGwConfig() (*gateway.Config, error) {
	cfg := gateway.DefaultConfig()
	cfg.ListenAddr = envWithDefault("TOOLGW_ADDR", ":8082")
	cfg.ToolsConfigPath = envWithDefault("TOOLGW_CONFIG", "/etc/agentic/gw-tools.json")
	cfg.Observability.Port = envWithDefault("OTEL_PORT", "8083")
	return cfg, nil
}

func envWithDefault(k, defVal string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return defVal
}

// ------------------- 探针 -------------------
func startProbe(cfg *gateway.Config) {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("ok")) })
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("ready")) })
	mux.Handle("/metrics", observability.MustHandler())
	s := &http.Server{Addr: ":8084", Handler: mux}
	log.Println("🩺 probe on :8084")
	_ = s.ListenAndServe()
}
//Personal.AI order the ending
