// cmd/agent-runtime/main.go
package main

import (
	"log"
)

const ServiceName = "agent-runtime"

func main() {
	// TODO: This main function is broken and needs to be fixed.
	// Commenting out for now to allow compilation.
	log.Println("agent-runtime is starting...")
}

/*
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
*/
//Personal.AI order the ending
