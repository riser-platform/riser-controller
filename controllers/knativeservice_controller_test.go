package controllers

import (
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/stretchr/testify/assert"
	knserving "knative.dev/serving/pkg/apis/serving/v1"
)

func Test_createStatusFromKnativeSvc(t *testing.T) {
	ksvc := &knserving.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "mydep",
			Annotations: map[string]string{
				riserLabel("generation"): "1",
			},
		},
		Status: knserving.ServiceStatus{
			ConfigurationStatusFields: knserving.ConfigurationStatusFields{
				LatestReadyRevisionName: "rev1",
			},
		},
	}
	revisions := []revisionDeployment{
		revisionDeployment{
			Revision: knserving.Revision{
				ObjectMeta: metav1.ObjectMeta{
					Name: "rev0",
					Labels: map[string]string{
						riserLabel("deployment"): "mydep",
					},
					Annotations: map[string]string{
						riserLabel("generation"): "0",
					},
				},
				Spec: knserving.RevisionSpec{
					PodSpec: corev1.PodSpec{
						Containers: []corev1.Container{
							corev1.Container{
								Name:  "mydep",
								Image: "my/image:0.0.1",
							},
							corev1.Container{Name: "istio-proxy"},
						},
					},
				},
			},
		},
		revisionDeployment{
			Revision: knserving.Revision{
				ObjectMeta: metav1.ObjectMeta{
					Name: "rev1",
					Labels: map[string]string{
						riserLabel("deployment"): "mydep",
					},
					Annotations: map[string]string{
						riserLabel("generation"): "1",
					},
				},
				Spec: knserving.RevisionSpec{
					PodSpec: corev1.PodSpec{
						Containers: []corev1.Container{
							corev1.Container{Name: "istio-proxy"},
							corev1.Container{
								Name:  "mydep",
								Image: "my/image:0.0.2",
							},
						},
					},
				},
			},
			Deployment: appsv1.Deployment{
				Status: appsv1.DeploymentStatus{
					AvailableReplicas: 1,
				},
			},
		},
	}

	result, err := createStatusFromKnativeSvc(ksvc, revisions)

	require.NoError(t, err)
	assert.Equal(t, int64(1), result.ObservedRiserGeneration)
	assert.Equal(t, "rev1", result.LatestReadyRevisionName)
	// Revisions
	require.Len(t, result.Revisions, 2)
	assert.Equal(t, "rev0", result.Revisions[0].Name)
	assert.Equal(t, int32(0), result.Revisions[0].AvailableReplicas)
	assert.Equal(t, "my/image:0.0.1", result.Revisions[0].DockerImage)
	assert.Equal(t, int64(0), result.Revisions[0].RiserGeneration)
	assert.Equal(t, "rev1", result.Revisions[1].Name)
	assert.Equal(t, int32(1), result.Revisions[1].AvailableReplicas)
	assert.Equal(t, "my/image:0.0.2", result.Revisions[1].DockerImage)
	assert.Equal(t, int64(1), result.Revisions[1].RiserGeneration)
	// TODO: Rollout status
}
