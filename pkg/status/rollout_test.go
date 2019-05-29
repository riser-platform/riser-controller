package status

import (
	"riser-controller/pkg/util"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"testing"

	"github.com/riser-platform/riser-server/api/v1/model"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
)

var rolloutTests = []struct {
	name       string
	deployment *appsv1.Deployment
	expected   RolloutStatus
}{
	{
		name: "Returns failed when progress times out",
		deployment: &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name: "myDeployment",
			},
			Status: appsv1.DeploymentStatus{
				Conditions: []appsv1.DeploymentCondition{
					{
						Type:   appsv1.DeploymentProgressing,
						Reason: TimedOutReason,
					},
				},
			},
		},
		expected: RolloutStatus{
			Status: model.RolloutStatusFailed,
			Reason: "Deployment timed out",
		},
	},
	{
		name: "Returns InProgress when updating new replicas",
		deployment: &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name: "myDeployment",
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: util.Int32Ptr(3),
			},
			Status: appsv1.DeploymentStatus{
				UpdatedReplicas: 1,
			},
		},
		expected: RolloutStatus{
			Status: model.RolloutStatusInProgress,
			Reason: "Waiting for rollout to finish: 1 out of 3 new replicas have been updated...",
		},
	},
	{
		name: "Returns InProgress when terminating old replicas",
		deployment: &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name: "myDeployment",
			},
			Status: appsv1.DeploymentStatus{
				Replicas:        3,
				UpdatedReplicas: 2,
			},
		},
		expected: RolloutStatus{
			Status: model.RolloutStatusInProgress,
			Reason: "Waiting for rollout to finish: 1 old replicas are pending termination...",
		},
	},
	{
		name: "Returns InProgress when updating existing replicas",
		deployment: &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name: "myDeployment",
			},
			Status: appsv1.DeploymentStatus{
				AvailableReplicas: 2,
				UpdatedReplicas:   3,
			},
		},
		expected: RolloutStatus{
			Status: model.RolloutStatusInProgress,
			Reason: "Waiting for rollout to finish: 2 of 3 updated replicas are available...",
		},
	},
	{
		name: "Returns Rollout Complete",
		deployment: &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name: "myDeployment",
			},
			Status: appsv1.DeploymentStatus{
				AvailableReplicas: 2,
				UpdatedReplicas:   2,
			},
		},
		expected: RolloutStatus{
			Status: model.RolloutStatusComplete,
			Reason: "Successfully rolled out",
		},
	},
	{
		name: "Returns InProgress when status has not yet been updated",
		deployment: &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "myDeployment",
				Generation: 2,
			},
			Status: appsv1.DeploymentStatus{
				ObservedGeneration: 1,
			},
		},
		expected: RolloutStatus{
			Status: model.RolloutStatusInProgress,
			Reason: "Waiting for deployment details...",
		},
	},
}

func Test_GetRolloutStatus(t *testing.T) {
	for _, tt := range rolloutTests {
		result := GetRolloutStatus(tt.deployment)

		assert.Equal(t, tt.expected, result, tt.name)
	}
}
