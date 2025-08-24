// pkg/apis/security.go
package apis

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	types "github.com/turtacn/agenticai/pkg/types"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SecurityPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec types.SecurityPolicy `json:"spec"`

	// 可选 status 后期扩展
	// Status SecurityPolicyStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SecurityPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []SecurityPolicy `json:"items"`
}

func init() {
	localSchemeBuilder.Register(addSecurityPolicyTypes)
}

func addSecurityPolicyTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&SecurityPolicy{},
		&SecurityPolicyList{},
	)
	return nil
}
//Personal.AI order the ending
