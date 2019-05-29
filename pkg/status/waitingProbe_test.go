package status

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

func Test_waitingProbe(t *testing.T) {
	probe := waitingProbe{}
	pod := &corev1.Pod{
		Status: corev1.PodStatus{
			Phase: corev1.PodPending,
			ContainerStatuses: []corev1.ContainerStatus{
				corev1.ContainerStatus{
					Name:  "myapp",
					Ready: false,
					State: corev1.ContainerState{
						Waiting: &corev1.ContainerStateWaiting{
							Reason:  "ImagePullBackoff",
							Message: "Bad image",
						},
					},
				},
			},
		},
	}

	result := probe.GetProblem(pod)

	assert.Equal(t, `Container "myapp" is waiting: ImagePullBackoff (Bad image)`, result.Message)
}

func Test_waitingProbe_WhenPodInitializing_NoProblem(t *testing.T) {
	probe := waitingProbe{}
	pod := &corev1.Pod{
		Status: corev1.PodStatus{
			Phase: corev1.PodPending,
			ContainerStatuses: []corev1.ContainerStatus{
				corev1.ContainerStatus{
					Name:  "myapp",
					Ready: false,
					State: corev1.ContainerState{
						Waiting: &corev1.ContainerStateWaiting{
							Reason: "PodInitializing",
						},
					},
				},
			},
		},
	}

	result := probe.GetProblem(pod)

	assert.Nil(t, result)
}
