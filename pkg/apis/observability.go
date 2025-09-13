// pkg/apis/observability.go
package apis

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// Telemetry CR 名称
const (
	TelemetryKind = "Telemetry"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Telemetry struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ObservabilitySpec `json:"spec"`
}

// ---------- 顶层可观测配置 ----------
// +k8s:deepcopy-gen=true
type ObservabilitySpec struct {
	Metrics                MetricsConfig `json:"metrics,omitempty"`
	Tracing                TracingConfig `json:"tracing,omitempty"`
	Logging                LoggingConfig `json:"logging,omitempty"`
	Alerts                 []AlertRule   `json:"alerts,omitempty"`
	ServiceLevelIndicators []SLI         `json:"slis,omitempty"`
	ServiceLevelObjectives []SLO         `json:"slos,omitempty"`
}

// ---------- Metrics 配置 ----------
type MetricsConfig struct {
	ScrapeInterval metav1.Duration   `json:"scrapeInterval,omitempty"` // default 30s
	Path           string            `json:"path,omitempty"`           // /metrics
	Port           int32             `json:"port,omitempty"`           // 9090
	SystemLabels   map[string]string `json:"systemLabels,omitempty"`   // 固定label
	Rules          []RecordingRule   `json:"rules,omitempty"`          // 自定义表达式
}

type RecordingRule struct {
	Name       string            `json:"name,string"`
	Expression string            `json:"expression,string"`
	Labels     map[string]string `json:"labels,omitempty"`
}

// ---------- Tracing 配置 ----------
type TracingConfig struct {
	Backend  string            `json:"backend"`   // otlp / jaeger / zipkin / otlpgrpc
	Endpoint string            `json:"endpoint"`  // <host>:<port>
	Headers  map[string]string `json:"headers,omitempty"`
	Sampling SamplingConfig    `json:"sampling,omitempty"`
}

type SamplingConfig struct {
	Ratio       float64 `json:"ratio,omitempty"`    // 0-1, default 0.1
	ParentBased bool    `json:"parentBased,omitempty"`
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
	Name        string            `json:"name"`
	Expr        string            `json:"expr"`
	For         metav1.Duration   `json:"for,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

// ---------- Service Level Indicator ----------
type SLI struct {
	Name     string `json:"name"`
	Expr     string `json:"expr"`     // PromQL
	Resource string `json:"resource"` // e.g. "Task" or "Agent"
}

// ---------- Service Level Objective ----------
type SLO struct {
	Name        string          `json:"name"`
	Objective   float64         `json:"objective"` // e.g. 0.999
	Budget      float64         `json:"budget"`    // seconds of error budget
	Window      metav1.Duration `json:"window"`    // e.g. "30d"
	SLIRef      string          `json:"sliRef"`    // maps SLI.Name
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type TelemetryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Telemetry `json:"items"`
}

func init() {
	localSchemeBuilder.Register(addTelemetryTypes)
}

func addTelemetryTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&Telemetry{},
		&TelemetryList{},
	)
	return nil
}
//Personal.AI order the ending
