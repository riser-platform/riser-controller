package status

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

type waitingProbe struct{}

/*
	This may go away in favor of more specific probes in the future. For example, if waiting is due to an ImagePullBackoff,
	the message is not very helpful. Instead we could search for most recent ErrImage event to provide debugging details.
*/
func (probe *waitingProbe) GetProblem(pod *corev1.Pod) *PodProblem {
	if pod.Status.Phase == corev1.PodPending {
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if !containerStatus.Ready && containerStatus.State.Waiting != nil && containerStatus.State.Waiting.Reason != "PodInitializing" {
				return &PodProblem{
					Message: fmt.Sprintf("Container %q is waiting: %s (%s)", containerStatus.Name, containerStatus.State.Waiting.Reason, containerStatus.State.Waiting.Message),
				}
			}
		}
	}

	return nil
}
