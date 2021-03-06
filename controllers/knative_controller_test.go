package controllers

import (
	"errors"
	"net/http"
	"riser-controller/pkg/util"
	"testing"

	"github.com/stretchr/testify/require"
	corea1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/stretchr/testify/assert"
	knserving "knative.dev/serving/pkg/apis/serving/v1"
)

func Test_createStatusFromKnative(t *testing.T) {
	cfg := &knserving.Configuration{
		ObjectMeta: metav1.ObjectMeta{
			Name: "mydep",
			Annotations: map[string]string{
				riserLabel("revision"): "1",
			},
		},
		Status: knserving.ConfigurationStatus{
			ConfigurationStatusFields: knserving.ConfigurationStatusFields{
				LatestReadyRevisionName:   "rev0",
				LatestCreatedRevisionName: "rev1",
			},
		},
	}
	route := &knserving.Route{
		Status: knserving.RouteStatus{
			RouteStatusFields: knserving.RouteStatusFields{
				Traffic: []knserving.TrafficTarget{
					{
						RevisionName: "rev0",
						Percent:      util.PtrInt64(90),
						Tag:          "r0",
					},
					{
						RevisionName: "rev1",
						Percent:      util.PtrInt64(10),
						Tag:          "r1",
					},
				},
			},
		},
	}
	revisions := []knserving.Revision{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "rev0",
				Labels: map[string]string{
					riserLabel("deployment"): "mydep",
				},
				Annotations: map[string]string{
					riserLabel("revision"): "0",
				},
			},
			Spec: knserving.RevisionSpec{
				PodSpec: corea1.PodSpec{
					Containers: []corea1.Container{
						{
							Name:  "mydep",
							Image: "my/image:0.0.1",
						},
						{Name: "istio-proxy"},
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "rev1",
				Labels: map[string]string{
					riserLabel("deployment"): "mydep",
				},
				Annotations: map[string]string{
					riserLabel("revision"): "1",
				},
			},
			Spec: knserving.RevisionSpec{
				PodSpec: corea1.PodSpec{
					Containers: []corea1.Container{
						{Name: "istio-proxy"},
						{
							Name:  "mydep",
							Image: "my/image:0.0.2",
						},
					},
				},
			},
		},
	}

	result, err := createStatusFromKnative(cfg, route, revisions)

	require.NoError(t, err)
	assert.Equal(t, int64(1), result.ObservedRiserRevision)
	assert.Equal(t, "rev1", result.LatestCreatedRevisionName)
	assert.Equal(t, "rev0", result.LatestReadyRevisionName)

	// Revisions
	require.Len(t, result.Revisions, 2)
	assert.Equal(t, "rev0", result.Revisions[0].Name)
	assert.Equal(t, "my/image:0.0.1", result.Revisions[0].DockerImage)
	assert.Equal(t, int64(0), result.Revisions[0].RiserRevision)
	assert.Equal(t, "rev1", result.Revisions[1].Name)
	assert.Equal(t, "my/image:0.0.2", result.Revisions[1].DockerImage)
	assert.Equal(t, int64(1), result.Revisions[1].RiserRevision)

	// Traffic
	require.Len(t, result.Traffic, 2)
	assert.Equal(t, "rev0", result.Traffic[0].RevisionName)
	assert.Equal(t, int64(90), *result.Traffic[0].Percent)
	assert.Equal(t, "r0", result.Traffic[0].Tag)
	assert.Equal(t, "rev1", result.Traffic[1].RevisionName)
	assert.Equal(t, int64(10), *result.Traffic[1].Percent)
	assert.Equal(t, "r1", result.Traffic[1].Tag)
}

func Test_handleDeploymentsSaveStatusResult(t *testing.T) {
	reconciler := &KNativeReconciler{}
	logger := &FakeLogger{
		FakeInfoLogger: FakeInfoLogger{
			InfoFn: func(msg string, keysAndValues ...interface{}) {
				assert.EqualValues(t, keysAndValues[0], []interface{}{"observedRiserRevision", int64(1)})
				assert.Equal(t, "Saved deployment status", msg)
			},
		},
	}

	result, err := reconciler.handleDeploymentsSaveStatusResult(logger, 1, http.StatusOK, nil)

	assert.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, result)
	assert.Equal(t, 1, logger.FakeInfoLogger.InfoCallCount)
}

func Test_handleDeploymentsSaveStatusResult_RequeueOnError(t *testing.T) {
	reconciler := &KNativeReconciler{}
	saveErr := errors.New("failed")
	logger := &FakeLogger{
		ErrorFn: func(err error, msg string, keysAndValues ...interface{}) {
			assert.EqualValues(t, keysAndValues[0], []interface{}{"observedRiserRevision", int64(1)})
			assert.Equal(t, saveErr, err)
			assert.Equal(t, "Error saving deployment status", msg)
		},
	}

	result, err := reconciler.handleDeploymentsSaveStatusResult(logger, 1, http.StatusInternalServerError, saveErr)

	assert.Equal(t, err, saveErr)
	assert.Equal(t, ctrl.Result{Requeue: true}, result)
	assert.Equal(t, 1, logger.ErrorCallCount)
}

func Test_handleDeploymentsSaveStatusResult_DoesNotRequeueOnConflict(t *testing.T) {
	reconciler := &KNativeReconciler{}
	saveErr := errors.New("failed")
	logger := &FakeLogger{
		ErrorFn: func(err error, msg string, keysAndValues ...interface{}) {
			assert.EqualValues(t, keysAndValues[0], []interface{}{"observedRiserRevision", int64(1)})
			assert.Equal(t, saveErr, err)
			assert.Equal(t, "Error saving deployment status: conflict", msg)
		},
	}

	result, err := reconciler.handleDeploymentsSaveStatusResult(logger, 1, http.StatusConflict, saveErr)

	assert.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, result)
	assert.Equal(t, 1, logger.ErrorCallCount)
}
