package status

import (
	"testing"

	"github.com/riser-platform/riser-server/api/v1/model"
	"github.com/stretchr/testify/assert"

	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	knserving "knative.dev/serving/pkg/apis/serving/v1"
)

func Test_getRevisionStatus(t *testing.T) {
	tt := []struct {
		name     string
		rev      *knserving.Revision
		expected RevisionStatus
	}{
		{
			name: "status is not observed",
			rev: &knserving.Revision{
				Status: knserving.RevisionStatus{
					Status: duckv1.Status{
						ObservedGeneration: 0,
					},
				},
			},
			expected: RevisionStatus{
				Status: model.RevisionStatusWaiting,
				Reason: "Deploying",
			},
		},
		{
			name: "deployment has timed out",
			rev: &knserving.Revision{
				Status: knserving.RevisionStatus{
					Status: duckv1.Status{
						ObservedGeneration: 1,
						Conditions: duckv1.Conditions{
							apis.Condition{
								Type:    "Active",
								Status:  "False",
								Reason:  "TimedOut",
								Message: "Timed out message.",
							},
						},
					},
				},
			},
			expected: RevisionStatus{
				Status: model.RevisionStatusUnhealthy,
				Reason: "Timed out message. This is usually due to a failing health check",
			},
		},
		{
			name: "ready state is unknown and deploying",
			rev: &knserving.Revision{
				Status: knserving.RevisionStatus{
					Status: duckv1.Status{
						ObservedGeneration: 1,
						Conditions: duckv1.Conditions{
							apis.Condition{
								Type:   "Active",
								Status: "True",
							},
							apis.Condition{
								Type:   "Ready",
								Status: "Unknown",
								Reason: "Deploying",
							},
						},
					},
				},
			},
			expected: RevisionStatus{
				Status: model.RevisionStatusWaiting,
				Reason: "Deploying",
			},
		},
		{
			name: "ready state is unknown and resolving digests",
			rev: &knserving.Revision{
				Status: knserving.RevisionStatus{
					Status: duckv1.Status{
						ObservedGeneration: 1,
						Conditions: duckv1.Conditions{
							apis.Condition{
								Type:   "Active",
								Status: "True",
							},
							apis.Condition{
								Type:   "Ready",
								Status: "Unknown",
								Reason: "ResolvingDigests",
							},
						},
					},
				},
			},
			expected: RevisionStatus{
				Status: model.RevisionStatusWaiting,
				Reason: "Deploying",
			},
		},
		{
			name: "ready state is unknown",
			rev: &knserving.Revision{
				Status: knserving.RevisionStatus{
					Status: duckv1.Status{
						ObservedGeneration: 1,
						Conditions: duckv1.Conditions{
							apis.Condition{
								Type:    "Ready",
								Status:  "Unknown",
								Message: "msg",
							},
						},
					},
				},
			},
			expected: RevisionStatus{
				Status: model.RevisionStatusUnknown,
				Reason: "msg",
			},
		},
		{
			name: "container is not healthy",
			rev: &knserving.Revision{
				Status: knserving.RevisionStatus{
					Status: duckv1.Status{
						ObservedGeneration: 1,
						Conditions: duckv1.Conditions{
							apis.Condition{
								Type:    "ContainerHealthy",
								Status:  "False",
								Message: "Failed to resolve image to digest...",
							},
						},
					},
				},
			},
			expected: RevisionStatus{
				Status: model.RevisionStatusUnhealthy,
				Reason: "Failed to resolve image to digest...",
			},
		},
		{
			name: "ready is true",
			rev: &knserving.Revision{
				Status: knserving.RevisionStatus{
					Status: duckv1.Status{
						ObservedGeneration: 1,
						Conditions: duckv1.Conditions{
							apis.Condition{
								Type:   "Ready",
								Status: "True",
							},
						},
					},
				},
			},
			expected: RevisionStatus{
				Status: model.RevisionStatusReady,
			},
		},
		{
			name: "ready is false",
			rev: &knserving.Revision{
				Status: knserving.RevisionStatus{
					Status: duckv1.Status{
						ObservedGeneration: 1,
						Conditions: duckv1.Conditions{
							apis.Condition{
								Type:    "Ready",
								Status:  "False",
								Message: "msg",
							},
						},
					},
				},
			},
			expected: RevisionStatus{
				Status: model.RevisionStatusUnhealthy,
				Reason: "msg",
			},
		},
	}

	for _, test := range tt {
		result := GetRevisionStatus(test.rev)
		assert.Equal(t, test.expected, result, "when %s", test.name)
	}
}
