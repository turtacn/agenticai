// internal/constants/constants.go
package constants

import (
	"github.com/uber-go/zap"
	"time"
)

// Generic consts
const (
	ProjectName = "agenticai"
	Version     = "0.1.0-dev"
)

// API / Networking
const (
	DefaultHTTPPort      = 8080
	DefaultMetricsPort   = 9090
	DefaultGRPCPort      = 8090
	DefaultTimeout       = 30 * time.Second
	DefaultRetryBackoff  = 500 * time.Millisecond
	DefaultMaxRetry      = 3
	RequestIDHeader      = "X-Request-Id"
	TraceparentHeader    = "traceparent"
)

// Kubernetes
const (
	ControllerWorkerThreads = 4
	DefaultNamespace        = "agenticai-system"
	AgentCRDName            = "agents.agenticai.io"
	TaskCRDName             = "tasks.agenticai.io"
)

// Security
const (
	TrustDomain             = "agenticai.io"
	SVIDTTLMinutes          = 60
	SVIDRefreshThresholdPct = 50
	DefaultSecurityPolicy   = "default"
)

// Observability
const (
	MetricsNamespace = "agenticai"
	MetricsSubsystem = "core"
	SystemNameLabel  = "agenticai.io/component"
)

// Sandbox
const (
	DefaultSandboxCPU    = "500m"
	DefaultSandboxMemory = "512Mi"
	DefaultSandboxDisk   = "1Gi"
	LabelIsAgent         = "agenticai.io/is-agent"
)

// Storage
const (
	DefaultVectorCollectionPrefix = "agenticai_"
	DefaultMaxObjectSizeBytes     = 100 * 1024 * 1024 // 100 MiB
	DefaultPresignedURLExpiration = 1 * time.Hour
)

// Logging & Debugging
const (
	DefaultLogLevel      = zap.InfoLevel
	ConsoleEncoding      = "console"
	OutputPathsJSON      = `["stdout"]`
	DisableStacktraceKey = "disable"
	//Personal.AI order the ending
)
