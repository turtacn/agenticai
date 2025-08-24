// cmd/controller/main.go
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/turtacn/agenticai/internal/config"
	"github.com/turtacn/agenticai/pkg/controller"
	"github.com/turtacn/agenticai/pkg/observability"
)

const ServiceName = "agentic-controller"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// 1. 初始化配置
	cfg, err := config.Load()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Fatalf("load config: %v", err)
	}
	cfg.ServiceName = ServiceName

	// 2. 初始化可观测性
	shutdownObs, err := observability.Init(ctx, cfg.Observability, ServiceName)
	if err != nil {
		log.Fatalf("observability: %v", err)
	}
	defer shutdownObs()

	// 3. 启动控制器
	mgr, err := controller.NewManager(cfg)
	if err != nil {
		log.Fatalf("controller: %v", err)
	}
	go func() {
		if err := mgr.Start(ctx); err != nil {
			log.Fatalf("controller start: %v", err)
		}
	}()

	// 4. 管理探针 & 优雅退出
	startHealth()
	select {
	case <-ctx.Done():
		log.Println("⏳ received shutdown signal")
	}
	log.Println("🛑 controller stopped")

	// 5. 宽限期优雅关闭
	ctx2, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := mgr.Shutdown(ctx2); err != nil {
		log.Fatalf("graceful shutdown: %v", err)
	}
	log.Println("✅ controller exited")
}

var server *http.Server

// startHealth 启动探针 HTTP 服务
func startHealth() {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready"))
	})
	metrics, _ := observability.Handler()
	mux.Handle("/metrics", metrics)

	server = &http.Server{
		Addr:    fmt.Sprintf(":%s", os.Getenv("PORT")),
		Handler: mux,
	}
	if server.Addr == ":PORT" {
		server.Addr = ":8080"
	}
	go func() {
		log.Printf("🚀 probe server listening on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("probe: %v", err)
		}
	}()
}
//Personal.AI order the ending
