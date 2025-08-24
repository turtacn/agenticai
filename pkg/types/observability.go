// pkg/types/observability.go
package types

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ---------- 顶层可观测配置 ----------
type ObservabilitySpec struct {
	Metrics       MetricsConfig       `json:"metrics,omitempty"`
	Tracing       TracingConfig       `json:"tracing,omitempty"`
	Logging       LoggingConfig       `json:"logging,omitempty"`
	Alerts        []AlertRule         `json:"alerts,omitempty"`
	ServiceLevelIndicators []SLI   `json:"slis,omitempty"`
	serviceLevelObjectives []SLO   `json:"slos,omitempty"`
}

// ---------- Metrics 配置 ----------
type MetricsConfig struct {
	ScrapeInterval metav1.Duration          `json:"scrapeInterval,omitempty"` // default 30s
	Path           string                   `json:"path,omitempty"`           // /metrics
	Port           int32                    `json:"port,omitempty"`           // 9090
	SystemLabels   map[string]string        `json:"systemLabels,omitempty"`   // 固定label
	Rules          []RecordingRule          `json:"rules,omitempty"`          // 自定义表达式
}

type RecordingRule struct {
	Name       string `json:"name,string"`
	Expression string `json:"expression,string"`
	Labels     map[string]string `json:"labels,omitempty"`
}

// ---------- Tracing 配置 ----------
type TracingConfig struct {
	Backend   string            `json:"backend"`   // otlp / jaeger / zipkin / otlpgrpc
	Endpoint  string            `json:"endpoint"`  // <host>:<port>
	Headers   map[string]string `json:"headers,omitempty"`
	Sampling  SamplingConfig    `json:"sampling,omitempty"`
}

type SamplingConfig struct {
	Ratio    float64 `json:"ratio,omitempty"`    // 0-1, default 0.1
	ParentBased bool `json:"parentBased,omitempty"`
}

// ---------- Logging 配置 ----------
type LoggingConfig struct {
	Level    string `json:"level,omitempty"`    // info/debug/error
	Output   string `json:"output,omitempty"`   // stdout/stdout/dev
	Format   string `json:"format,omitempty"`   // json/console
	MaxSize  int64  `json:"maxSize,omitempty"`  // bytes 阈值
	MaxFiles int32  `json:"maxFiles,omitempty"` // rotate 保留
}

// ---------- AlertRule（兼容 PrometheusRule） ----------
type AlertRule struct {
	Name        string `json:"name"`
	Expr        string `json:"expr"`
	For         metav1.Duration `json:"for,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

// ---------- Service Level Indicator ----------
type SLI struct {
	Name     string `json:"name"`
	Expr     string `json:"expr"`       // PromQL
	Resource string `json:"resource"`   // e.g. "Task" or "Agent"
}

// ---------- Service Level Objective ----------
type SLO struct {
	Name        string `json:"name"`
	Objective   float64 `json:"objective"` // e.g. 0.999
	Budget      float64 `json:"budget"`    // seconds of error budget
	Window      metav1.Duration `json:"window"` // e.g. "30d"
	SLIRef      string `json:"sliRef"`       // maps SLI.Name
}

// ---------- 实时事件 ----------
type ObservableEvent struct {
	Kind      ObservableEventKind `json:"kind"` // metric/span/log/event
	Name      string              `json:"name"` // metric_key or span_name
	Value     float64             `json:"value,omitempty"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
	Timestamp time.Time           `json:"timestamp"`
}

type ObservableEventKind string

const (
	EventMetric = ObservableEventKind("metric")
	EventSpan   = ObservableEventKind("span")
	EventLog    = ObservableEventKind("log")
	EventAlert  = ObservableEventKind("alert")
)

// ---------- 运行时暴露结构 ----------
type TelemetrySnapshot struct {
	AgentID string                  `json:"agentId"`
	TaskID  string                  `json:"taskId,omitempty"` // empty for controller
	Events  []ObservableEvent        `json:"events"`
}

// ---------- 告警状态 ----------
type AlertStatus struct {
	Name      string                 `json:"name"`
	State     string                 `json:"state"` // firing/resolved/pending
	Value     float64                `json:"value"`
	Tags      map[string]string      `json:"tags"`
	LastFire  metav1.Time            `json:"lastFire,omitempty"`
	NextFire  metav1.Time            `json:"nextFire,omitempty"`
}

// Personal.AI order the ending
