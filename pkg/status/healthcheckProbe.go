package status

import (
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
)

const defaultPodReadinessSeconds = 30

type healthcheckProbe struct{}

func (probe *healthcheckProbe) GetProblem(pod *corev1.Pod) *Problem {
	// Scenario: Rollout completes. Old version goes away. New version readiness (health check) fails (e.g. after a few minutes and all old pods are terminated)
	if pod.Status.Phase == corev1.PodRunning {
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if !containerStatus.Ready {
				if containerStatus.State.Running != nil {
					// TODO: Can we be in a running state while not ready for any other reason other than pod readiness?
					// TODO: look up the readiness probe and calculate a reasonable threshold between InitialDelaySeconds, TimeoutSeconds, and FailureThreshold
					if time.Since(containerStatus.State.Running.StartedAt.Time) > defaultPodReadinessSeconds*time.Second {
						return &Problem{Message: fmt.Sprintf("Container %q failing health check", containerStatus.Name)}
					}
				}
			}
		}
	}
	return nil
}
