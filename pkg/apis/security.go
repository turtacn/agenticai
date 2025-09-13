// pkg/apis/security.go
package apis

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SecurityPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec PolicySpec `json:"spec"`

	// 可选 status 后期扩展
	// Status SecurityPolicyStatus `json:"status,omitempty"`
}

// PolicySpec 兼容 Kubernetes 约束框架
// +k8s:deepcopy-gen=true
type PolicySpec struct {
	Match       PolicyMatch       `json:"match"`
	Constraints PolicyConstraints `json:"constraints"`
	Rules       []RBACRule        `json:"rules,omitempty"`
}

type PolicyMatch struct {
	Kind        string            `json:"kind"`        // "Agent" | "Task" ...
	Labels      map[string]string `json:"labels,omitempty"`
	Namespaces  []string          `json:"namespaces,omitempty"`
	Expressions []string          `json:"expressions,omitempty"`
}
type PolicyConstraints struct {
	// 例：禁止高危系统调用
	SysCallRestriction []string `json:"sysCallRestriction,omitempty"`
	// 例：只读根文件系统
	ReadOnlyRootFS bool `json:"readOnlyRootFS"`
	// 例：禁用 privileged
	AllowPrivileged bool `json:"allowPrivileged"`
}

// RBACRule 用于内存策略树
type RBACRule struct {
	Role      string   `json:"role"`
	Verbs     []string `json:"verbs"`  // "create","delete"...
	Resources []string `json:"resources"` // "Agent","Task",...
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
