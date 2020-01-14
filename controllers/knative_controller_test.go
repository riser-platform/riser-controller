package controllers

import (
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"riser-controller/pkg/util"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/stretchr/testify/assert"
	knserving "knative.dev/serving/pkg/apis/serving/v1"
)

func Test_createStatusFromKnative(t *testing.T) {
	cfg := &knserving.Configuration{
		ObjectMeta: metav1.ObjectMeta{
			Name: "mydep",
			Annotations: map[string]string{
				riserLabel("generation"): "1",
			},
		},
		Status: knserving.ConfigurationStatus{
			ConfigurationStatusFields: knserving.ConfigurationStatusFields{
				LatestReadyRevisionName:   "rev0",
				LatestCreatedRevisionName: "rev1",
			},
		},
	}
	route := &knserving.Route{
		Status: knserving.RouteStatus{
			RouteStatusFields: knserving.RouteStatusFields{
				Traffic: []knserving.TrafficTarget{
					knserving.TrafficTarget{
						RevisionName: "rev0",
						Percent:      util.PtrInt64(90),
					},
					knserving.TrafficTarget{
						RevisionName:   "rev1",
						LatestRevision: util.PtrBool(true),
						Percent:        util.PtrInt64(10),
					},
				},
			},
		},
	}
	revisions := []revisionEtc{
		revisionEtc{
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
		revisionEtc{
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
			Deployment: &appsv1.Deployment{
				Status: appsv1.DeploymentStatus{
					AvailableReplicas: 1,
				},
			},
		},
	}

	result, err := createStatusFromKnative(cfg, route, revisions)

	require.NoError(t, err)
	assert.Equal(t, int64(1), result.ObservedRiserGeneration)
	assert.Equal(t, "rev1", result.LatestCreatedRevisionName)
	assert.Equal(t, "rev0", result.LatestReadyRevisionName)

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

	// Traffic
	require.Len(t, result.Traffic, 2)
	assert.Equal(t, "rev0", result.Traffic[0].RevisionName)
	assert.Equal(t, int64(90), *result.Traffic[0].Percent)
	assert.Nil(t, result.Traffic[0].Latest)
	assert.Equal(t, "rev1", result.Traffic[1].RevisionName)
	assert.Equal(t, int64(10), *result.Traffic[1].Percent)
	assert.True(t, *result.Traffic[1].Latest)
}