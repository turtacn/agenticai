// pkg/controller/task_controller.go
package controller

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"reflect"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	ctrl "sigs.k8s.io/controller-runtime"
	"k8s.io/utils/pointer"

	agenticaiov1 "github.com/turtacn/agenticai/pkg/apis/agenticai.io/v1"
	"github.com/turtacn/agenticai/internal/logger"
)

// TaskReconciler reconciles a Task object
type TaskReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=agenticai.io,resources=tasks,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=agenticai.io,resources=tasks/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=agenticai.io,resources=tasks/finalizers,verbs=update
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=pods;persistentvolumeclaims,verbs=get;list;watch
//+kubebuilder:rbac:groups=agenticai.io,resources=agents,verbs=get;list;watch

func (r *TaskReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logCtx := logger.WithCtx(ctx).With(zap.String("task", req.NamespacedName.String()))
	log := logCtx.Sugar()

	var task agenticaiov1.Task
	if err := r.Client.Get(ctx, req.NamespacedName, &task); err != nil {
		if apierrs.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error("fetch task error", zap.Error(err))
		return ctrl.Result{Requeue: true, RequeueAfter: 5 * time.Second}, nil
	}

	// 预检查 spec
	if err := task.Spec.Validate(); err != nil {
		log.Errorf("invalid spec: %v", err)
		r.updateStatus(ctx, &task, agenticaiov1.TaskFailed, fmt.Sprintf("invalid spec: %v", err), 0)
		return ctrl.Result{}, nil
	}

	// 依赖检查
	ready, err := r.checkDependencies(ctx, &task)
	if err != nil {
		log.Errorf("dependency error: %v", err)
		return ctrl.Result{}, err
	}
	if !ready {
		log.Infof("dependencies not ready, backoff 10s")
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	createdJob, err := r.ensureJob(ctx, &task)
	if err != nil {
		log.Errorf("ensure job error: %v", err)
		return ctrl.Result{RequeueAfter: 20 * time.Second}, nil
	}

	// 读取 Job 的 Pod 状态来更新
	podResult, err := r.loadRunningPodStatus(ctx, createdJob)
	if err != nil {
		log.Errorf("load pod status error: %v", err)
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}

	// 更新 Task.Status
	oldStatus := task.Status.DeepCopy()
	task.Status.Phase = podResult.Phase
	task.Status.Message = podResult.Message
	task.Status.Progress = podResult.Progress
	task.Status.TaskResult = podResult.Result
	if !reflect.DeepEqual(oldStatus, &task.Status) {
		if err := r.Status().Update(ctx, &task); err != nil {
			// 冲突时重入
			return ctrl.Result{Requeue: true, RequeueAfter: 2 * time.Second}, nil
		}
	}

	return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
}

// ensureJob：根据 Task.Spec 创建 Batch Job
func (r *TaskReconciler) ensureJob(ctx context.Context, task *agenticaiov1.Task) (*batchv1.Job, error) {
	jobName := fmt.Sprintf("%s-job", task.Name)
	ns := task.Namespace
	if ns == "" {
		ns = metav1.NamespaceDefault
	}

	jobSpec := batchv1.JobSpec{
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: task.Spec.Labels,
			},
			Spec: corev1.PodSpec{
				RestartPolicy: corev1.RestartPolicyNever,
				Containers: []corev1.Container{
					{
						Name:      "task-runner",
						Image:     task.Spec.ImageRef,
						Command:   task.Spec.Command,
						Args:      task.Spec.Args,
						Resources: task.Spec.Resources,
						Env:       task.Spec.Env,
					},
				},
			},
		},
		Parallelism:           pointer.Int32(1),
		Completions:           pointer.Int32(1),
		ActiveDeadlineSeconds: pointer.Int64(int64(task.Spec.Timeout.Duration.Seconds())),
		BackoffLimit:          pointer.Int32(task.Spec.RetryPolicy.Limit),
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: ns,
		},
		Spec: jobSpec,
	}

	// ownerRef 确保任务删除时 Job 级联
	_ = controllerutil.SetControllerReference(task, job, r.Scheme)

	found := &batchv1.Job{}
	err := r.Get(ctx, types.NamespacedName{Name: jobName, Namespace: ns}, found)
	switch {
	case err == nil:
		return found, nil
	case apierrs.IsNotFound(err):
		if err := r.Create(ctx, job); err != nil {
			return nil, err
		}
		return job, nil
	default:
		return nil, err
	}
}

// checkDependencies：遍历 Task.Dependencies 检查前置是否完成
func (r *TaskReconciler) checkDependencies(ctx context.Context, task *agenticaiov1.Task) (bool, error) {
	for _, dep := range task.Spec.Dependencies {
		var depTask agenticaiov1.Task
		if err := r.Get(ctx, types.NamespacedName{Name: dep.TaskID, Namespace: task.Namespace}, &depTask); err != nil {
			return false, err
		}
		if depTask.Status.Phase != dep.State {
			return false, nil
		}
	}
	return true, nil
}

type podStatusResult struct {
	Phase    agenticaiov1.TaskPhase
	Message  string
	Progress int32
	Result   *agenticaiov1.TaskResult // 非 nil 代表终止
}

// loadRunningPodStatus：读 Job Pod 实时状态
func (r *TaskReconciler) loadRunningPodStatus(ctx context.Context, job *batchv1.Job) (*podStatusResult, error) {
	podList := &corev1.PodList{}
	selector, _ := metav1.LabelSelectorAsSelector(job.Spec.Selector)
	if err := r.List(ctx, podList, client.InNamespace(job.Namespace), client.MatchingLabelsSelector{Selector: selector}); err != nil {
		return nil, err
	}
	if len(podList.Items) == 0 {
		return &podStatusResult{Phase: agenticaiov1.TaskPending, Message: "waiting for pod schedule"}, nil
	}
	pod := &podList.Items[0]
	switch pod.Status.Phase {
	case corev1.PodPending:
		return &podStatusResult{Phase: agenticaiov1.TaskPending, Message: pod.Status.Message}, nil
	case corev1.PodRunning:
		return &podStatusResult{Phase: agenticaiov1.TaskRunning, Progress: percentFromAnnotations(pod.Annotations)}, nil
	case corev1.PodSucceeded:
		exitCode := int32(0)
		if len(pod.Status.ContainerStatuses) > 0 {
			if state := pod.Status.ContainerStatuses[0].State.Terminated; state != nil {
				exitCode = state.ExitCode
			}
		}
		return &podStatusResult{
			Phase:   agenticaiov1.TaskCompleted,
			Message: "finished",
			Result:  &agenticaiov1.TaskResult{ExitCode: exitCode, Output: pod.Annotations["output"], Artifact: pod.Annotations["artifact"]},
		}, nil
	case corev1.PodFailed:
		msg := "pod failed"
		if len(pod.Status.ContainerStatuses) > 0 {
			if state := pod.Status.ContainerStatuses[0].State.Terminated; state != nil {
				msg = state.Message
			}
		}
		return &podStatusResult{
			Phase:   agenticaiov1.TaskFailed,
			Message: msg,
			Result:  &agenticaiov1.TaskResult{ExitCode: 1},
		}, nil
	}
	return &podStatusResult{Phase: agenticaiov1.TaskPending, Message: "unknown"}, nil
}

// percentFromAnnotations 示例提取 Pod 进度
func percentFromAnnotations(ann map[string]string) int32 {
	if v, ok := ann["agenticai.io/progress"]; ok {
		var p int32
		fmt.Sscanf(v, "%d", &p)
		return p
	}
	return 0
}

// updateStatus 快捷工具
func (r *TaskReconciler) updateStatus(ctx context.Context, task *agenticaiov1.Task, phase agenticaiov1.TaskPhase, msg string, progress int32) {
	task.Status.Phase = phase
	task.Status.Message = msg
	task.Status.Progress = progress
	_ = r.Status().Update(ctx, task)
}

// SetupWithManager attach watch
func (r *TaskReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&agenticaiov1.Task{}).
		Owns(&batchv1.Job{}).
		Complete(r)
}
//Personal.AI order the ending
