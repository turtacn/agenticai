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
