package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/turtacn/agenticai/pkg/apis"
)

func TestBuildDeployment(t *testing.T) {
	s := runtime.NewScheme()
	scheme.AddToScheme(s)
	apis.AddToScheme(s)

	agent := &apis.Agent{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-agent",
			Namespace: "default",
		},
		Spec: apis.AgentSpec{
			ImageRef: "test-image:latest",
			Replicas: 2,
			Labels: map[string]string{
				"custom-label": "custom-value",
			},
		},
	}

	reconciler := &AgentReconciler{
		Scheme: s,
	}

	deployment, err := reconciler.buildDeployment(agent)
	assert.NoError(t, err)
	assert.NotNil(t, deployment)

	assert.Equal(t, "test-agent", deployment.Name)
	assert.Equal(t, "default", deployment.Namespace)
	assert.Equal(t, int32(2), *deployment.Spec.Replicas)
	assert.Equal(t, "test-image:latest", deployment.Spec.Template.Spec.Containers[0].Image)
	assert.Equal(t, "custom-value", deployment.Spec.Template.Labels["custom-label"])
}
