// pkg/utils/kubernetes.go
package utils

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// KubeClient 返回全局 K8s 客户端实例；优先尝试 In-Cluster → KUBECONFIG
func KubeClient() (*kubernetes.Clientset, error) {
	var cfg *rest.Config
	var err error
	cfg, err = rest.InClusterConfig()
	if err != nil {
		// fallback 到本地 kubeconfig
		rules := clientcmd.NewDefaultClientConfigLoadingRules()
		cfg, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, &clientcmd.ConfigOverrides{}).ClientConfig()
		if err != nil {
			return nil, fmt.Errorf("build k8s config: %w", err)
		}
	}
	return kubernetes.NewForConfig(cfg)
}

// CreateOrUpdate deploys a Pod，若已存在则执行滚动替换
func CreateOrUpdate(ctx context.Context, cs *kubernetes.Clientset, ns string, pod *corev1.Pod) error {
	pods := cs.CoreV1().Pods(ns)
	_, err := pods.Get(ctx, pod.Name, metav1.GetOptions{})
	switch {
	case err == nil: // 已存在 -> 删除后建
		err = pods.Delete(ctx, pod.Name, metav1.DeleteOptions{GracePeriodSeconds: pointerInt64(0)})
		if err != nil {
			return fmt.Errorf("delete pod: %w", err)
		}
		// 简单轮询等待直到消失再创建
		for range 10 {
			_, err := pods.Get(ctx, pod.Name, metav1.GetOptions{})
			if apierrors.IsNotFound(err) {
				break
			}
			time.Sleep(500 * time.Millisecond)
		}
	case !apierrors.IsNotFound(err): // 网络错误
		return fmt.Errorf("get pod: %w", err)
	}
	_, err = pods.Create(ctx, pod, metav1.CreateOptions{})
	return err
}

// DeletePod 按名删除 Pod（同步等待）
func DeletePod(ctx context.Context, cs *kubernetes.Clientset, ns, name string) error {
	pods := cs.CoreV1().Pods(ns)
	err := pods.Delete(ctx, name, metav1.DeleteOptions{GracePeriodSeconds: pointerInt64(0)})
	if apierrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}
	// 等待实际删除
	for range 60 {
		_, err := pods.Get(ctx, name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return nil
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("pod %s still terminating after 60s", name)
}

// GetPodByLabel 简单封装
func GetPodByLabel(ctx context.Context, cs *kubernetes.Clientset, ns, labelSel string) (*corev1.PodList, error) {
	return cs.CoreV1().
		Pods(ns).
		List(ctx, metav1.ListOptions{LabelSelector: labelSel})
}

// LabelBuilder 快速创建键值选择器
func LabelBuilder(k, v string) labels.Selector {
	return labels.SelectorFromSet(labels.Set{k: v})
}

// AddLabels 为对象追加/更新标签（元数据变更）
func AddLabels(obj metav1.Object, kv map[string]string) {
	labels := obj.GetLabels()
	if labels == nil {
		labels = kv
	} else {
		for k, v := range kv {
			labels[k] = v
		}
	}
	obj.SetLabels(labels)
}

// AddAnnotations 统一注解操作
func AddAnnotations(obj metav1.Object, kv map[string]string) {
	ann := obj.GetAnnotations()
	if ann == nil {
		ann = kv
	} else {
		for k, v := range kv {
			ann[k] = v
		}
	}
	obj.SetAnnotations(ann)
}

// WaitPodReady 阻塞直到 Pod 成为 Ready，默认 120s timeout
func WaitPodReady(ctx context.Context, cs *kubernetes.Clientset, ns, name string) error {
	ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()
	watch, err := cs.CoreV1().
		Pods(ns).
		Watch(ctx, metav1.ListOptions{FieldSelector: fmt.Sprintf("metadata.name=%s", name)})
	if err != nil {
		return fmt.Errorf("watch pod: %w", err)
	}
	defer watch.Stop()

	for e := range watch.ResultChan() {
		switch e.Type {
		case "ERROR":
			return fmt.Errorf("watch error")
		case "MODIFIED", "ADDED":
			pod := e.Object.(*corev1.Pod)
			for _, cond := range pod.Status.Conditions {
				if cond.Type == corev1.PodReady && cond.Status == corev1.ConditionTrue {
					return nil
				}
			}
		}
	}
	return fmt.Errorf("timeout waiting pod %s ready", name)
}

// ClusterHealth 快速检查 api server 是否存活
func ClusterHealth(ctx context.Context, cs *kubernetes.Clientset) error {
	_, err := cs.Discovery().RESTClient().Get().AbsPath("/healthz").Timeout(3 * time.Second).DoRaw(ctx)
	return err
}

// Pointer helper
func pointerInt64(i int64) *int64 { return &i }
//Personal.AI order the ending
