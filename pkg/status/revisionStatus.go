package status

import (
	"github.com/riser-platform/riser-server/api/v1/model"
	knserving "knative.dev/serving/pkg/apis/serving/v1"
)

/*

Ready Reason Deploying -> Deploying
Ready

Revision
1) Check ResourcesAvailable Reason Deploying -> Deploying
2) Check Ready Reason Deploying -> Deploying
Revision Pods
1) Check Ready
Note: Could check PodScheduled if there's limited resources?
*/

// TODO: What about Pod Problems at the revision level?
// TODO: Test
func GetRevisionStatus(rev *knserving.Revision) RolloutStatus {
	if len(rev.Status.Conditions) == 0 {
		return RolloutStatus{Status: model.RolloutStatusUnknown}
	}

	for _, cnd := range rev.Status.Conditions {
		if cnd.Type == "Ready" {
			if cnd.IsUnknown() && cnd.Reason != "Deploying" {
				return RolloutStatus{Status: model.RolloutStatusUnknown, Reason: cnd.Message}
			}
			if cnd.IsFalse() {
				return RolloutStatus{Status: model.RolloutStatusFailed, Reason: cnd.Message}
			}
		}
	}

	return RolloutStatus{Status: model.RolloutStatusComplete}
}
