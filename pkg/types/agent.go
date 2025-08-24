// pkg/types/agent.go
package types

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/turtacn/agenticai/internal/constants"
)

// AgentSpec 描述创建或更新 Agent 的期望状态
type AgentSpec struct {
	// 元数据
	Version  string            `json:"version,omitempty"`
	ImageRef string            `json:"imageRef,omitempty"`
	Replicas int32             `json:"replicas,omitempty"`
	Labels   map[string]string `json:"labels,omitempty"`

	// 资源需求
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// GPU 调度
	GPU struct {
		Count   int32  `json:"count,omitempty"`   // GPU 数量
		Device  string `json:"device,omitempty"`  // device type/selector
		MIGMode bool   `json:"migMode,omitempty"` // MIG 切分
	} `json:"gpu,omitempty"`

	// 沙箱配置
	Sandbox SandboxPolicy `json:"sandbox,omitempty"`

	// 安全策略
	Security struct {
		WorkloadIdentity    string                 `json:"workloadIdentity,omitempty"`
		ServiceAccountName  string                 `json:"serviceAccountName,omitEmpty"`
		PodSecurityContext  *corev1.PodSecurityContext `json:"podSecurityContext,omitempty"`
		SecurityContext    *corev1.SecurityContext     `json:"securityContext,omitempty"`
		AllowedCapabilities []corev1.Capability     `json:"allowedCapabilities,omitempty"`
	} `json:"security,omitempty"`

	// 任务队列
	TaskSelector *metav1.LabelSelector `json:"taskSelector,omitempty"`

	// 节点亲和
	NodeSelector map[string]string        `json:"nodeSelector,omitempty"`
	Tolerations  []corev1.Toleration      `json:"tolerations,omitempty"`
	Affinity     *corev1.Affinity         `json:"affinity,omitempty"`
}

// AgentStatus 描述当前实际运行状态
type AgentStatus struct {
	Phase   AgentPhase `json:"phase"`
	Message string     `json:"message,omitempty"`

	// Replica status
	DesiredReplicas int32 `json:"desiredReplicas"`
	ReadyReplicas   int32 `json:"readyReplicas"`

	// 运行信息
	StartTime      *metav1.Time     `json:"startTime,omitempty"`
	CompletionTime *metav1.Time     `json:"completionTime,omitempty"`
	Conditions     []AgentCondition `json:"conditions,omitempty"`

	// GPU 资源实际分配
	AssignedGPUs []string `json:"assignedGPUs,omitempty"`

	// 版本校验
	CurrentVersion string `json:"currentVersion,omitempty"`
}

// AgentPhase 定义可能的生命周期状态
type AgentPhase string

const (
	AgentPending   AgentPhase = "Pending"
	AgentCreating  AgentPhase = "Creating"
	AgentRunning   AgentPhase = "Running"
	AgentStopping  AgentPhase = "Stopping"
	AgentStopped   AgentPhase = "Stopped"
	AgentFailed    AgentPhase = "Failed"
	AgentSucceeded AgentPhase = "Succeeded")
)

// AgentCondition 扩展 CRD status.conditions
type AgentCondition struct {
	Type               string             `json:"type"`
	Status             corev1.ConditionStatus `json:"status"`
	LastUpdateTime     metav1.Time        `json:"lastUpdateTime"`
	LastTransitionTime metav1.Time        `json:"lastTransitionTime"`
	Reason             string             `json:"reason"`
	Message            string             `json:"message"`
}

// SandboxPolicy 指定沙箱级别及补充参数
type SandboxPolicy struct {
	Type    string            `json:"type"` // gvisor / kata / firecracker
	Options map[string]string `json:"options,omitempty"`
}

// Validate 为输入校验钩子
func (s *AgentSpec) Validate() error {
	if s.ImageRef == "" {
		return fmt.Errorf("imageRef required")
	}
	if s.Resources.Requests == nil {
		s.Resources.Requests = corev1.ResourceList{}
	}
	if s.Resources.Limits == nil {
		s.Resources.Limits = corev1.ResourceList{}
	}
	// 补充默认值
	if s.Sandbox.Type == "" {
		s.Sandbox.Type = "gvisor"
	}
	if s.Resources.Requests.Cpu().IsZero() {
		defCpu := resource.MustParse(constants.DefaultSandboxCPU)
		s.Resources.Requests[corev1.ResourceCPU] = defCpu
	}
	if s.Resources.Requests.Memory().IsZero() {
		defMem := resource.MustParse(constants.DefaultSandboxMemory)
		s.Resources.Requests[corev1.ResourceMemory] = defMem
	}
	return nil
}
//Personal.AI order the ending
