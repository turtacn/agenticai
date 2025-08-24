module github.com/turtacn/agenticai

go 1.20

require (
	// K8s ecosystem
	k8s.io/apimachinery v0.29.0
	k8s.io/client-go    v0.29.0
	k8s.io/api          v0.29.0
	k8s.io/component-base v0.29.0
	sigs.k8s.io/controller-runtime v0.17.0

	// CLI
	github.com/spf13/cobra v1.8.0
	github.com/spf13/viper v1.18.2

	// HTTP & gRPC
	github.com/gin-gonic/gin v1.9.1
	google.golang.org/grpc v1.62.0
	google.golang.org/protobuf v1.33.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.19.1

	// OpenTelemetry
	go.opentelemetry.io/otel v1.25.0
	go.opentelemetry.io/otel/trace v1.25.0
	go.opentelemetry.io/otel/metric v1.25.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.50.0
	go.opentelemetry.io/otel/exporters/prometheus v0.47.0

	// Prometheus
	github.com/prometheus/client_golang v1.19.0
	github.com/prometheus/common v0.51.1

	// Logging
	go.uber.org/zap v1.27.0

	// JSON / YAML
	github.com/json-iterator/go v1.1.12
	github.com/ghodss/yaml v1.0.0
	gopkg.in/yaml.v3 v3.0.1

	// Misc utils
	github.com/google/uuid v1.6.0
	github.com/hashicorp/go-retryablehttp v0.7.5
	github.com/fsnotify/fsnotify v1.7.0

	// SPIFFE
	github.com/spiffe/go-spiffe/v2 v2.2.0

)
//Personal.AI order the ending
