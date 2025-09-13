// pkg/controller/agent_controller.go
package controller

import (
	"context"
	"go.uber.org/zap"
	"reflect"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/turtacn/agenticai/internal/constants"
	"github.com/turtacn/agenticai/internal/logger"
	"github.com/turtacn/agenticai/pkg/apis"
)

// AgentReconciler reconciles a Agent object
type AgentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=agenticai.io,resources=agents,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=agenticai.io,resources=agents/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=agenticai.io,resources=agents/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=pods;persistentvolumeclaims;services;secrets;serviceaccounts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="apps",resources=deployments;replicasets,verbs=get;list;watch;create;update;patch;delete

func (r *AgentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logger.WithCtx(ctx).With(
		zap.String("agent", req.NamespacedName.String()),
	)

	// 1. 拉取 CR
	var agent apis.Agent
	if err := r.Get(ctx, req.NamespacedName, &agent); err != nil {
		if errors.IsNotFound(err) {
			// Object deleted, ignore
			return ctrl.Result{}, nil
		}
		log.Error("unable to fetch Agent", zap.Error(err))
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}

	// 2. 校验 spec
	if err := agent.Spec.Validate(); err != nil {
		log.Error("invalid spec observed, skip reconcile", zap.Error(err))
		return ctrl.Result{}, nil
	}

	// 3. 处理默认命名空间
	if agent.Namespace == "" {
		agent.Namespace = constants.DefaultNamespace
	}

	// 4. 根据 spec 生成 Deployment
	deploy, err := r.buildDeployment(&agent)
	if err != nil {
		log.Error("unable to build deployment", zap.Error(err))
		r.markFailed(ctx, &agent, err)
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	// 5. 设置 ownerRef→保证级联删除
	if err := controllerutil.SetControllerReference(&agent, deploy, r.Scheme); err != nil {
		log.Error("unable to set owner ref", zap.Error(err))
		return ctrl.Result{}, err
	}

	// 6. 创建或更新 Deployment
	found := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{
		Name:      deploy.Name,
		Namespace: deploy.Namespace,
	}, found)

	if err != nil && errors.IsNotFound(err) {
		log.Info("creating deployment", zap.String("deployment", deploy.Name))
		if err := r.Create(ctx, deploy); err != nil {
			log.Error("unable to create deployment", zap.Error(err))
			return ctrl.Result{}, err
		}
	} else if err == nil {
		// 更新 replica / image / resources / env ...
		found.Spec = deploy.Spec
		if err := r.Update(ctx, found); err != nil {
			log.Error("unable to update deployment", zap.Error(err))
			return ctrl.Result{}, err
		}
		log.Info("deployment updated", zap.String("deployment", deploy.Name))
	} else {
		return ctrl.Result{}, err
	}

	// 7. 计算副本就绪状态
	sts, err := r.computeStatus(ctx, &agent, deploy)
	if err != nil {
		log.Error("compute status failed", zap.Error(err))
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}
	if !reflect.DeepEqual(sts, agent.Status) {
		agent.Status = *sts
		if err := r.Status().Update(ctx, &agent); err != nil {
			log.Error("unable to update agent status", zap.Error(err))
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// buildDeployment 拼装业务 + 沙箱 sidecars
func (r *AgentReconciler) buildDeployment(agent *apis.Agent) (*appsv1.Deployment, error) {
	podLabels := map[string]string{
		"app.kubernetes.io/name":       constants.ProjectName,
		"app.kubernetes.io/component":  "agent",
		"app.kubernetes.io/instance":   agent.Name,
		constants.LabelIsAgent:         "true",
	}
	for k, v := range agent.Spec.Labels {
		podLabels[k] = v
	}

	replicas := int32(agent.Spec.Replicas)
	if replicas == 0 {
		replicas = 1
	}

	// 构造 container
	mainContainer := corev1.Container{
		Name:            "agent",
		Image:           agent.Spec.ImageRef,
		ImagePullPolicy: corev1.PullIfNotPresent,
		Resources:       agent.Spec.Resources,
		// Env:             envVars(agent.Spec.Env), // TODO: AgentSpec does not have Env
	}
	// if len(agent.Spec.Command) > 0 { // TODO: AgentSpec does not have Command
	// 	mainContainer.Command = agent.Spec.Command
	// }
	// if len(agent.Spec.Args) > 0 { // TODO: AgentSpec does not have Args
	// 	mainContainer.Args = agent.Spec.Args
	// }

	// 拼接 volume mounts & sidecar
	podSpec := corev1.PodSpec{
		Containers: []corev1.Container{mainContainer},
		// sidecar injection 按 agent.Spec.Sandbox.Type
	}

	// label 选择器
	matchLabels := labels.Set{
		"app.kubernetes.io/name":      constants.ProjectName,
		"app.kubernetes.io/component": "agent",
		"app.kubernetes.io/instance":  agent.Name,
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      agent.Name,
			Namespace: agent.Namespace,
			Labels:    podLabels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: matchLabels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: podLabels},
				Spec:       podSpec,
			},
		},
	}, nil
}

// computeStatus 收集底层 deployment 状态到 AgentStatus
func (r *AgentReconciler) computeStatus(ctx context.Context, agent *apis.Agent, delp *appsv1.Deployment) (*apis.AgentStatus, error) {
	deployment := &appsv1.Deployment{}
	if err := r.Get(ctx, types.NamespacedName{
		Name:      delp.Name,
		Namespace: delp.Namespace,
	}, deployment); err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	sts := apis.AgentStatus{
		DesiredReplicas: *deployment.Spec.Replicas,
		ReadyReplicas:   deployment.Status.ReadyReplicas,
		CurrentVersion:  deployment.Annotations["image-version"], // 自定义
		Phase:           apis.AgentRunning,
		Message:         "deployment healthy",
	}
	return &sts, nil
}

func (r *AgentReconciler) markFailed(ctx context.Context, agent *apis.Agent, err error) {
	agent.Status.Phase = apis.AgentFailed
	agent.Status.Message = err.Error()
	_ = r.Status().Update(ctx, agent)
}

func envVars(src []corev1.EnvVar) []corev1.EnvVar {
	if len(src) == 0 {
		return nil
	}
	return src
}

// SetupWithManager wired into controller-manager
func (r *AgentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apis.Agent{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
//Personal.AI order the ending
