// pkg/apis/agent.go
package apis

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	types "github.com/turtacn/agenticai/pkg/types"
)

// GroupVersion / Kind
const (
	GroupName = "agenticai.io"
	Version   = "v1"
	Kind      = "Agent"
)

// SchemeGroupVersion 用于 RESTMapper
var SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: Version}

// Agent CRD 顶层结构
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Agent struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   types.AgentSpec   `json:"spec"`
	Status types.AgentStatus `json:"status"`
}

// AgentList 列表
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type AgentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Agent `json:"items"`
}

// Scheme 构造器，Register 到全局 scheme
func init() {
	localSchemeBuilder.Register(addKnownTypes)
}

// localSchemeBuilder
var localSchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)

// addKnownTypes
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&Agent{},
		&AgentList{},
	)
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
//Personal.AI order the ending
