// pkg/types/observability.go
package types

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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
