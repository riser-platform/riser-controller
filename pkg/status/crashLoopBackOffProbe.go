package status

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

type crashLoopBackOffProbe struct {
}

func (probe *crashLoopBackOffProbe) GetProblem(pod *corev1.Pod) *Problem {
	if pod.Status.Phase == corev1.PodRunning {
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if !containerStatus.Ready && containerStatus.State.Waiting != nil && containerStatus.State.Waiting.Reason == "CrashLoopBackOff" {
				return &Problem{Message: fmt.Sprintf("%s: %s", containerStatus.State.Waiting.Reason, containerStatus.State.Waiting.Message)}
			}
		}
	}

	return nil
}
