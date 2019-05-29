package status

import (
	"fmt"

	"github.com/riser-platform/riser-server/api/v1/model"

	corev1 "k8s.io/api/core/v1"
)

type PodProblem struct {
	Message string
}

type ProblemList struct {
	problemMap map[string]model.DeploymentStatusProblem
}

type PodProblemProbe interface {
	GetProblem(pod *corev1.Pod) *PodProblem
}

func (list *ProblemList) Items() []model.DeploymentStatusProblem {
	items := []model.DeploymentStatusProblem{}
	for _, item := range list.problemMap {
		items = append(items, item)
	}

	return items
}

func (list *ProblemList) AddProblem(message string) {
	if list.problemMap == nil {
		list.problemMap = map[string]model.DeploymentStatusProblem{}
	}
	if problem, found := list.problemMap[message]; found {
		problem.Count = problem.Count + 1
		list.problemMap[message] = problem
	} else {
		list.problemMap[message] = model.DeploymentStatusProblem{Count: 1, Message: message}
	}
}

func GetPodProblems(pods *corev1.PodList) *ProblemList {
	return getPodProblems(pods,
		&crashLoopBackOffProbe{},
		&healthcheckProbe{},
		&waitingProbe{},
	)
}

func getPodProblems(pods *corev1.PodList, probes ...PodProblemProbe) *ProblemList {
	podProblems := ProblemList{}
	if pods != nil {
		for _, pod := range pods.Items {
			fmt.Printf("%+v", pod.GetOwnerReferences())
			for _, probe := range probes {
				problem := probe.GetProblem(&pod)
				if problem != nil {
					podProblems.AddProblem(problem.Message)
				}
			}
		}
	}
	return &podProblems
}
