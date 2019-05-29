package status

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

func Test_crashLoopBackOffProbe_HasProblem(t *testing.T) {
	probe := crashLoopBackOffProbe{}
	pod := &corev1.Pod{
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			ContainerStatuses: []corev1.ContainerStatus{
				corev1.ContainerStatus{
					Ready: false,
					State: corev1.ContainerState{
						Waiting: &corev1.ContainerStateWaiting{
							Reason:  "CrashLoopBackOff",
							Message: "crash reason",
						},
					},
				},
			},
		},
	}

	result := probe.GetProblem(pod)

	assert.Equal(t, "CrashLoopBackOff: crash reason", result.Message)
}

func Test_crashLoopBackoffProbe_NoProblem(t *testing.T) {
	probe := crashLoopBackOffProbe{}
	pod := &corev1.Pod{}

	result := probe.GetProblem(pod)

	assert.Nil(t, result)
}
