package status

import (
	"fmt"

	"github.com/riser-platform/riser-server/api/v1/model"
	knserving "knative.dev/serving/pkg/apis/serving/v1"
)

type RevisionStatus struct {
	Status string
	Reason string
}

// GetRevisionStatus attempts to simplify the range of statuses returned by a revision. This will probably need to be completely rethought at some point
// Once this gets tested in the wild, will want to consider proposing upstream improvements to knative serving.
func GetRevisionStatus(rev *knserving.Revision) RevisionStatus {
	// For a new revision, the status is not observed right away. Assume that it's deploying.
	if rev.Status.ObservedGeneration == 0 {
		return RevisionStatus{Status: model.RevisionStatusWaiting, Reason: "Deploying"}
	}
	for _, cnd := range rev.Status.Conditions {
		if cnd.Type == "ContainerHealthy" && cnd.IsFalse() {
			return RevisionStatus{Status: model.RevisionStatusUnhealthy, Reason: cnd.Message}
		}
		if cnd.Type == "Active" && cnd.IsFalse() && cnd.Reason == "TimedOut" {
			return RevisionStatus{Status: model.RevisionStatusUnhealthy, Reason: fmt.Sprintf("%s This is usually due to a failing health check", cnd.Message)}
		}
		if cnd.Type == "Ready" {
			if cnd.IsUnknown() {
				if cnd.Reason == "Deploying" || cnd.Reason == "ResolvingDigests" {
					return RevisionStatus{Status: model.RevisionStatusWaiting, Reason: "Deploying"}
				}
				return RevisionStatus{Status: model.RevisionStatusUnknown, Reason: cnd.Message}
			}
			if cnd.IsTrue() {
				return RevisionStatus{Status: model.RevisionStatusReady}
			} else {
				return RevisionStatus{Status: model.RevisionStatusUnhealthy, Reason: cnd.Message}
			}
		}

	}

	return RevisionStatus{Status: model.RevisionStatusUnknown}
}
