// pkg/types/security.go
package types

import (
	"time"
)

// WorkloadIdentity 对应 SPIFFE spiffe://trust-domain/workload-name
// 字段映射：https://github.com/spiffe/spiffe/blob/main/standards/SPIFFE-ID.md
type WorkloadIdentity struct {
	SVID      string    `json:"svid"`               // DER encoded SVID string
	PublicKey string    `json:"public_key"`         // PEM or DER
	Issuer    string    `json:"issuer"`             // 即 CA SVID
	NotAfter  time.Time `json:"not_after"`
	NotBefore time.Time `json:"not_before"`
	Bundle    string    `json:"bundle"`             // trust-bundle PEM
}

// AuthResult 鉴权后统一结果
type AuthResult struct {
	Subject      string   `json:"subject"`      // SPIFFE ID / CN
	ExtraClaims  map[string]interface{} `json:"extra_claims,omitempty"`
	AllowedRoles []string `json:"allowed_roles"`
}

// SecurityPolicy 可序列化策略对象，与 Gatekeeper/Rego 兼容
type SecurityPolicy struct {
	ApiVersion string       `json:"apiVersion"`
	Kind       string       `json:"kind"`
	Metadata   PolicyMeta   `json:"metadata"`
	Spec       PolicySpec   `json:"spec"`
}

type PolicyMeta struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

// PolicySpec 兼容 Kubernetes 约束框架
type PolicySpec struct {
	Match       PolicyMatch       `json:"match"`
	Constraints PolicyConstraints `json:"constraints"`
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
	Role   string   `json:"role"`
	Verbs  []string `json:"verbs"`  // "create","delete"...
	Resources []string `json:"resources"` // "Agent","Task",...
}

// AuditEvent 所有控制层入口统一事件格式
type AuditEvent struct {
	EventTime time.Time `json:"event_time"`
	Actor     string    `json:"actor"`     // SPIFFE ID / User / Service Account
	Resource  string    `json:"resource"`  // Agent/Task ID
	Action    string    `json:"action"`    // CREATE/READ/UPDATE/DELETE
	Outcome   string    `json:"outcome"`   // success / denied / error
	Meta      map[string]interface{} `json:"meta,omitempty"`
	IP        string   `json:"ip,omitempty"`   // 客户端 / 节点
	SessionID string   `json:"session_id,omitempty"`
}

// ComplianceReport 定期扫描后自动生成
type ComplianceReport struct {
	Policy    string `json:"policy"`
	Checksum  string `json:"checksum"`
	Outcome   string `json:"outcome"` // pass/fail
	Details   map[string]interface{} `json:"details"`
	Generated time.Time `json:"generated"`
}

// SecretPolicy secret store backend
type SecretPolicy struct {
	Backend string `json:"backend"` // vault / k8s
	Auth    map[string]interface{} `json:"auth"`
	Paths   []string `json:"paths"` // 限制前缀
}
//Personal.AI order the ending
