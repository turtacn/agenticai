package v1

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Task is the Schema for the tasks API
type Task struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TaskSpec   `json:"spec,omitempty"`
	Status TaskStatus `json:"status,omitempty"`
}

// TaskSpec defines the desired state of Task
type TaskSpec struct {
	// 任务内容
	Description string   `json:"description,omitempty"`
	Artifacts   []string `json:"artifacts,omitempty"` // 输入参数或文件

	// 执行配置
	ImageRef    string `json:"imageRef"`
	Command     []string `json:"command,omitempty"`
	Args        []string `json:"args,omitempty"`
	Tools       []string `json:"tools,omitempty"` // 预定义已注册工具的 ID
	ContextID   string   `json:"contextId,omitempty"`

	// 资源
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// 策略
	Priority       int32           `json:"priority"` // 数值越大优先级越高
	Timeout        metav1.Duration `json:"timeout,omitempty"`
	RetryPolicy    RetryPolicy     `json:"retryPolicy,omitempty"`
	MaxConcurrency int32           `json:"maxConcurrency,omitempty"`

	// 依赖
	Dependencies []Dependency `json:"dependencies,omitempty"`

	// 环境变量
	Env []corev1.EnvVar `json:"env,omitempty"`

	// 标签体系
	Labels map[string]string `json:"labels,omitempty"`
}

// Validate checks the TaskSpec for correctness.
func (s *TaskSpec) Validate() error {
	if s.ImageRef == "" {
		return fmt.Errorf("imageRef is required")
	}
	return nil
}

// TaskStatus defines the observed state of Task
type TaskStatus struct {
	Phase      TaskPhase        `json:"phase"`
	Message    string           `json:"message,omitempty"`
	StartTime  *metav1.Time     `json:"startTime,omitempty"`
	EndTime    *metav1.Time     `json:"endTime,omitempty"`
	Progress   int32            `json:"progress"`
	NodeName   string           `json:"nodeName,omitempty"`
	CPUUsed    resource.Quantity `json:"cpuUsed,omitempty"`
	MemoryUsed resource.Quantity `json:"memoryUsed,omitempty"`
	GPUUsed    int64            `json:"gpuUsed,omitempty"`
	TaskResult *TaskResult      `json:"taskResult,omitempty"`
	Conditions []TaskCondition  `json:"conditions,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TaskList contains a list of Task
type TaskList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Task `json:"items"`
}

// TaskPhase defines the lifecycle phase of a Task.
type TaskPhase string

const (
	TaskPending     TaskPhase = "Pending"
	TaskScheduled   TaskPhase = "Scheduled"
	TaskRunning     TaskPhase = "Running"
	TaskCompleted   TaskPhase = "Completed"
	TaskFailed      TaskPhase = "Failed"
	TaskCancelled   TaskPhase = "Cancelled"
)

// TaskCondition aligns with K8s conditions.
type TaskCondition struct {
	Type               string             `json:"type"`
	Status             metav1.ConditionStatus `json:"status"`
	LastTransitionTime metav1.Time        `json:"lastTransitionTime"`
	Reason             string             `json:"reason"`
	Message            string             `json:"message"`
}

// TaskResult holds the output of a task.
type TaskResult struct {
	ExitCode int32  `json:"exitCode"`
	Output   string `json:"output,omitempty"`
	Artifact string `json:"artifact,omitempty"`
}

// RetryPolicy defines the retry strategy.
type RetryPolicy struct {
	Limit   int32           `json:"limit"`
	Backoff metav1.Duration `json:"backoff,omitempty"`
}

// Dependency defines a dependency on another task.
type Dependency struct {
	TaskID string    `json:"taskId"`
	State  TaskPhase `json:"state"`
}

func init() {
	SchemeBuilder.Register(&Task{}, &TaskList{})
}
