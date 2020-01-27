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

type RevisionStatus struct {
	Status string
	Reason string
}

// TODO: Add tests
// TODO: This is nowhere near exhaustive. However there are multiple issues related to Revision status so there's no point in going too deep right now
// https://github.com/knative/serving/issues/6265
// https://github.com/knative/serving/issues/6346
// https://github.com/knative/serving/issues/6489
func GetRevisionStatus(rev *knserving.Revision) RevisionStatus {
	for _, cnd := range rev.Status.Conditions {
		if cnd.Type == "Ready" {
			if cnd.IsUnknown() {
				if cnd.Reason == "Deploying" {
					return RevisionStatus{Status: model.RevisionStatusWaiting, Reason: cnd.Message}
				}
				return RevisionStatus{Status: model.RevisionStatusUnknown, Reason: cnd.Message}
			}
			if cnd.IsTrue() {
				return RevisionStatus{Status: model.RevisionStatusReady}
			} else {
				return RevisionStatus{Status: model.RevisionStatusUnhealthy, Reason: cnd.Message}
			}
		}
		// TODO: Check ResourcesAvailable condition?
	}

	return RevisionStatus{Status: model.RevisionStatusUnknown}
}
