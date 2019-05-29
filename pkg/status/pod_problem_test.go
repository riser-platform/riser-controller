package status

import (
	"sort"
	"strings"
	"testing"

	"github.com/riser-platform/riser-server/api/v1/model"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_ProblemList_AddProblem(t *testing.T) {
	problems := ProblemList{}
	problems.AddProblem("test")
	problems.AddProblem("test")
	problems.AddProblem("test2")

	result := problems.Items()

	expected := []model.DeploymentStatusProblem{
		model.DeploymentStatusProblem{Count: 2, Message: "test"},
		model.DeploymentStatusProblem{Count: 1, Message: "test2"},
	}

	assert.ElementsMatch(t, expected, result)
}

func Test_getPodProblems(t *testing.T) {
	pod1 := corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod1"}}
	pod2 := corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod2"}}
	pods := &corev1.PodList{}
	pods.Items = []corev1.Pod{pod1, pod2}
	probe1 := &fakeProbe{
		GetProblemFn: func(podArg *corev1.Pod) *PodProblem {
			if podArg.Name == "pod1" {
				return &PodProblem{Message: "problem1"}
			}
			return nil
		},
	}
	probe2 := &fakeProbe{
		GetProblemFn: func(podArg *corev1.Pod) *PodProblem {
			return &PodProblem{Message: "problem2"}
		},
	}

	result := getPodProblems(pods, probe1, probe2)

	assert.Len(t, result.Items(), 2)
	items := result.Items()
	sort.Slice(items, func(i, j int) bool {
		return strings.Compare(items[i].Message, items[j].Message) < 0
	})
	assert.Equal(t, "problem1", items[0].Message)
	assert.Equal(t, 1, items[0].Count)
	assert.Equal(t, "problem2", items[1].Message)
	assert.Equal(t, 2, items[1].Count)
}

type fakeProbe struct {
	GetProblemFn func(*corev1.Pod) *PodProblem
}

func (probe *fakeProbe) GetProblem(pod *corev1.Pod) *PodProblem {
	return probe.GetProblemFn(pod)
}
