package status

import (
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

func Test_healthcheckProbe_HasProblem(t *testing.T) {
	probe := healthcheckProbe{}
	startedAtAgo := (-defaultPodReadinessSeconds - 1) * time.Second
	pod := &corev1.Pod{
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			ContainerStatuses: []corev1.ContainerStatus{
				corev1.ContainerStatus{
					Name:  "myapp",
					Ready: false,
					State: corev1.ContainerState{
						Running: &corev1.ContainerStateRunning{
							StartedAt: metav1.Time{
								Time: time.Now().Add(startedAtAgo),
							},
						},
					},
				},
			},
		},
	}

	result := probe.GetProblem(pod)

	assert.Equal(t, `Container "myapp" failing health check`, result.Message)
}

func Test_healthcheckProbe_WhenRecentlyDeployed_NoProblem(t *testing.T) {
	probe := healthcheckProbe{}
	pod := &corev1.Pod{
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			ContainerStatuses: []corev1.ContainerStatus{
				corev1.ContainerStatus{
					Name:  "myapp",
					Ready: false,
					State: corev1.ContainerState{
						Running: &corev1.ContainerStateRunning{
							StartedAt: metav1.Time{
								Time: time.Now(),
							},
						},
					},
				},
			},
		},
	}

	result := probe.GetProblem(pod)

	assert.Nil(t, result)
}

func Test_healthcheckProbe_NoProblem(t *testing.T) {
	probe := healthcheckProbe{}
	pod := &corev1.Pod{}

	result := probe.GetProblem(pod)

	assert.Nil(t, result)
}
