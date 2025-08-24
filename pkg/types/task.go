// pkg/types/task.go
package types

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TaskSpec 定义任务提交时的期望
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
	Priority      int32          `json:"priority"`      // 数值越大优先级越高
	Timeout       metav1.Duration `json:"timeout,omitempty"`
	RetryPolicy   RetryPolicy     `json:"retryPolicy,omitempty"`
	MaxConcurrency int32          `json:"maxConcurrency,omitempty"`

	// 依赖
	Dependencies []Dependency `json:"dependencies,omitempty"`

	// 环境变量
	Env []corev1.EnvVar `json:"env,omitempty"`

	// 标签体系
	Labels map[string]string `json:"labels,omitempty"`
}

// Task 顶层对象（等价于 CR）
type Task struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TaskSpec   `json:"spec"`
	Status TaskStatus `json:"status"`
}

// TaskStatus 实时状态
type TaskStatus struct {
	Phase      TaskPhase        `json:"phase"`
	Message    string           `json:"message,omitempty"`
	StartTime  *metav1.Time     `json:"startTime,omitempty"`
	EndTime    *metav1.Time     `json:"endTime,omitempty"`

	// 进度 [0-100]
	Progress int32 `json:"progress"`

	// 节点
	NodeName string `json:"nodeName,omitempty"`

	// 资源
	CPUUsed    resource.Quantity `json:"cpuUsed,omitempty"`
	MemoryUsed resource.Quantity `json:"memoryUsed,omitempty"`
	GPUUsed    int64            `json:"gpuUsed,omitempty"`

	// 结果
	TaskResult *TaskResult `json:"taskResult,omitempty"`

	// 条件
	Conditions []TaskCondition `json:"conditions,omitempty"`
}

// TaskPhase 完整生命周期
type TaskPhase string

const (
	TaskPending     TaskPhase = "Pending"
	TaskScheduled   TaskPhase = "Scheduled"
	TaskRunning     TaskPhase = "Running"
	TaskCompleted   TaskPhase = "Completed"
	TaskFailed      TaskPhase = "Failed"
	TaskCancelled   TaskPhase = "Cancelled"
)

// TaskCondition 与 K8s 对齐
type TaskCondition struct {
	Type               string             `json:"type"`
	Status             metav1.ConditionStatus `json:"status"`
	LastTransitionTime metav1.Time        `json:"lastTransitionTime"`
	Reason             string             `json:"reason"`
	Message            string             `json:"message"`
}

// TaskResult 存放任务输出
type TaskResult struct {
	ExitCode int32  `json:"exitCode"`
	Output   string `json:"output,omitempty"`
	Artifact string `json:"artifact,omitempty"` // 对象存储 URI
}

// RetryPolicy 重试策略
type RetryPolicy struct {
	Limit    int32 `json:"limit"`
	Backoff  metav1.Duration `json:"backoff,omitempty"`
}

// Dependency 任务依赖
type Dependency struct {
	TaskID string `json:"taskId"`
	State  TaskPhase `json:"state"` // 依赖的任务必须达到的 Phase
}
//Personal.AI order the ending
