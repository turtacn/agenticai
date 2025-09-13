package controller

import (
	"context"
	"testing"

	agenticaiov1 "github.com/turtacn/agenticai/pkg/apis/agenticai.io/v1"
	"github.com/stretchr/testify/assert"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestCheckDependencies(t *testing.T) {
	s := runtime.NewScheme()
	scheme.AddToScheme(s)
	agenticaiov1.AddToScheme(s)

	// Test case 1: No dependencies
	task1 := &agenticaiov1.Task{
		ObjectMeta: metav1.ObjectMeta{Name: "task1", Namespace: "default"},
		Spec:       agenticaiov1.TaskSpec{},
	}
	reconciler := &TaskReconciler{
		Client: fake.NewClientBuilder().WithScheme(s).WithObjects(task1).Build(),
		Scheme: s,
	}
	ready, err := reconciler.checkDependencies(context.Background(), task1)
	assert.NoError(t, err)
	assert.True(t, ready)

	// Test case 2: Dependencies met
	depTask := &agenticaiov1.Task{
		ObjectMeta: metav1.ObjectMeta{Name: "dep1", Namespace: "default"},
		Status:     agenticaiov1.TaskStatus{Phase: agenticaiov1.TaskCompleted},
	}
	task2 := &agenticaiov1.Task{
		ObjectMeta: metav1.ObjectMeta{Name: "task2", Namespace: "default"},
		Spec: agenticaiov1.TaskSpec{
			Dependencies: []agenticaiov1.Dependency{
				{TaskID: "dep1", State: agenticaiov1.TaskCompleted},
			},
		},
	}
	reconciler = &TaskReconciler{
		Client: fake.NewClientBuilder().WithScheme(s).WithObjects(depTask, task2).Build(),
		Scheme: s,
	}
	ready, err = reconciler.checkDependencies(context.Background(), task2)
	assert.NoError(t, err)
	assert.True(t, ready)

	// Test case 3: Dependencies not met
	depTask.Status.Phase = agenticaiov1.TaskRunning
	task3 := &agenticaiov1.Task{
		ObjectMeta: metav1.ObjectMeta{Name: "task3", Namespace: "default"},
		Spec: agenticaiov1.TaskSpec{
			Dependencies: []agenticaiov1.Dependency{
				{TaskID: "dep1", State: agenticaiov1.TaskCompleted},
			},
		},
	}
	reconciler = &TaskReconciler{
		Client: fake.NewClientBuilder().WithScheme(s).WithObjects(depTask, task3).Build(),
		Scheme: s,
	}
	ready, err = reconciler.checkDependencies(context.Background(), task3)
	assert.NoError(t, err)
	assert.False(t, ready)
}

func TestLoadRunningPodStatus(t *testing.T) {
	s := runtime.NewScheme()
	scheme.AddToScheme(s)
	agenticaiov1.AddToScheme(s)

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-job",
			Namespace: "default",
		},
		Spec: batchv1.JobSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"job-name": "test-job"},
			},
		},
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
			Labels:    map[string]string{"job-name": "test-job"},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodPending,
		},
	}

	// Test case 1: Pod Pending
	reconciler := &TaskReconciler{
		Client: fake.NewClientBuilder().WithScheme(s).WithObjects(job, pod).Build(),
		Scheme: s,
	}
	result, err := reconciler.loadRunningPodStatus(context.Background(), job)
	assert.NoError(t, err)
	assert.Equal(t, agenticaiov1.TaskPending, result.Phase)

	// Test case 2: Pod Running
	pod.Status.Phase = corev1.PodRunning
	reconciler = &TaskReconciler{
		Client: fake.NewClientBuilder().WithScheme(s).WithObjects(job, pod).Build(),
		Scheme: s,
	}
	result, err = reconciler.loadRunningPodStatus(context.Background(), job)
	assert.NoError(t, err)
	assert.Equal(t, agenticaiov1.TaskRunning, result.Phase)

	// Test case 3: Pod Succeeded
	pod.Status.Phase = corev1.PodSucceeded
	reconciler = &TaskReconciler{
		Client: fake.NewClientBuilder().WithScheme(s).WithObjects(job, pod).Build(),
		Scheme: s,
	}
	result, err = reconciler.loadRunningPodStatus(context.Background(), job)
	assert.NoError(t, err)
	assert.Equal(t, agenticaiov1.TaskCompleted, result.Phase)
	assert.NotNil(t, result.Result)
	assert.Equal(t, int32(0), result.Result.ExitCode)

	// Test case 4: Pod Failed
	pod.Status.Phase = corev1.PodFailed
	reconciler = &TaskReconciler{
		Client: fake.NewClientBuilder().WithScheme(s).WithObjects(job, pod).Build(),
		Scheme: s,
	}
	result, err = reconciler.loadRunningPodStatus(context.Background(), job)
	assert.NoError(t, err)
	assert.Equal(t, agenticaiov1.TaskFailed, result.Phase)
	assert.NotNil(t, result.Result)
	assert.Equal(t, int32(1), result.Result.ExitCode)
}
