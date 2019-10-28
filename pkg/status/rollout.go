/*
Some portions of this code may fall under the following license.

Copyright 2016 The Kubernetes Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

References:

https://github.com/kubernetes/kubernetes/blob/5a7b978c7401efad5672bcc876bcf5d3dfa71bd0/pkg/kubectl/rollout_status.go
https://github.com/kubernetes/kubernetes/blob/5a7b978c7401efad5672bcc876bcf5d3dfa71bd0/pkg/kubectl/util/deployment/deployment.go

*/

package status

import (
	"fmt"

	"github.com/riser-platform/riser-server/api/v1/model"

	appsv1 "k8s.io/api/apps/v1"
)

const (
	// TimedOutReason is added in a deployment when its newest replica set fails to show any progress
	// within the given deadline (progressDeadlineSeconds).
	TimedOutReason = "ProgressDeadlineExceeded"
)

type RolloutStatus struct {
	Status string
	Reason string
}

// GetRolloutStatus returns a message describing deployment status, and a status indicating the state of the rollout.
// Currently does not support checking the desired revision
func GetRolloutStatus(deployment *appsv1.Deployment) RolloutStatus {
	if deployment.Generation <= deployment.Status.ObservedGeneration {
		cond := GetDeploymentCondition(deployment.Status, appsv1.DeploymentProgressing)
		if cond != nil && cond.Reason == TimedOutReason {
			return RolloutStatus{
				Status: model.RolloutStatusFailed,
				Reason: "Deployment timed out",
			}
		}
		if deployment.Spec.Replicas != nil && deployment.Status.UpdatedReplicas < *deployment.Spec.Replicas {
			return RolloutStatus{
				Status: model.RolloutStatusInProgress,
				Reason: fmt.Sprintf("%d/%d new replicas have been updated", deployment.Status.UpdatedReplicas, *deployment.Spec.Replicas),
			}
		}
		if deployment.Status.Replicas > deployment.Status.UpdatedReplicas {
			return RolloutStatus{
				Status: model.RolloutStatusInProgress,
				Reason: fmt.Sprintf("%d old replicas are pending termination", deployment.Status.Replicas-deployment.Status.UpdatedReplicas),
			}
		}
		if deployment.Status.AvailableReplicas < deployment.Status.UpdatedReplicas {
			return RolloutStatus{
				Status: model.RolloutStatusInProgress,
				Reason: fmt.Sprintf("%d/%d updated replicas are available", deployment.Status.AvailableReplicas, deployment.Status.UpdatedReplicas),
			}
		}
		return RolloutStatus{
			Status: model.RolloutStatusComplete,
			Reason: "Successfully rolled out",
		}
	}
	return RolloutStatus{
		Status: model.RolloutStatusInProgress,
		Reason: fmt.Sprintf("Waiting for deployment details"),
	}
}

// GetDeploymentCondition returns the condition with the provided type.
func GetDeploymentCondition(status appsv1.DeploymentStatus, condType appsv1.DeploymentConditionType) *appsv1.DeploymentCondition {
	for i := range status.Conditions {
		c := status.Conditions[i]
		if c.Type == condType {
			return &c
		}
	}
	return nil
}
