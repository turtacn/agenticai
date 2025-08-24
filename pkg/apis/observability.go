// pkg/apis/observability.go
package apis

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	types "github.com/turtacn/agenticai/pkg/types"
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

	Spec types.ObservabilitySpec `json:"spec"`
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
