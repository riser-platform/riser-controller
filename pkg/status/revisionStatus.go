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

func GetRevisionStatus(rev *knserving.Revision) RevisionStatus {
	// For a new revision, the status is not observed right away. Assume that it's deploying.
	if rev.Status.ObservedGeneration == 0 {
		return RevisionStatus{Status: model.RevisionStatusWaiting, Reason: "Deploying"}
	}
	for _, cnd := range rev.Status.Conditions {
		if cnd.Type == "Active" && cnd.IsFalse() && cnd.Reason == "TimedOut" {
			return RevisionStatus{Status: model.RevisionStatusUnhealthy, Reason: fmt.Sprintf("%s This is usually due to a failing health check", cnd.Message)}
		}
		if cnd.Type == "Ready" {
			if cnd.IsUnknown() {
				if cnd.Reason == "Deploying" {
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
