// cmd/tool-gateway/main.go
package main

import (
	"log"
)

const ServiceName = "tool-gateway"

func main() {
	// TODO: This main function is broken and needs to be fixed.
	// Commenting out for now to allow compilation.
	log.Println("tool-gateway is starting...")
}

/*
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
*/
//Personal.AI order the ending
