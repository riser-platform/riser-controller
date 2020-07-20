package controllers

import (
	"context"
	"fmt"
	"net/http"
	"riser-controller/pkg/runtime"
	"riser-controller/pkg/status"

	appsv1 "k8s.io/api/apps/v1"

	"github.com/pkg/errors"

	"github.com/riser-platform/riser-server/api/v1/model"

	"github.com/go-logr/logr"
	"github.com/riser-platform/riser-server/pkg/sdk"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	knserving "knative.dev/serving/pkg/apis/serving/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type KNativeConfigurationReconciler struct {
	KNativeReconciler
}

type KNativeRouteReconciler struct {
	KNativeReconciler
}

type KNativeReconciler struct {
	client.Client
	Log         logr.Logger
	Config      runtime.Config
	RiserClient *sdk.Client
}

// SetupWithManager functions for each type that we want to reconcile
// This arguably breaks the prescriptive reconcile pattern of one type per reconcile. Similar to a knative Service, we treat
// Configuration+Route as a single status entry in Riser. Options are:
// - Have separate status endpoints in the riser API for config and route status and update the statuses independently
// - Use a knative Service: this is problematic with the gitops pattern because the lifecycles are different for each resource
// - Keep doing what we're doing if there's no practical side effects (<- likely the right answer)
func (r *KNativeConfigurationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&knserving.Configuration{}).
		WithEventFilter(createUpdateRiserFilter()).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}

func (r *KNativeRouteReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&knserving.Route{}).
		WithEventFilter(createUpdateRiserFilter()).
		Complete(r)
}

func (r *KNativeReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("knative", req.NamespacedName)

	configuration := &knserving.Configuration{}

	err := r.Get(ctx, req.NamespacedName, configuration)
	if err != nil {
		if kerrors.IsNotFound(err) {
			log.Info("Configuration not found")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Unable to get knative configuration")
		return ctrl.Result{}, err
	}

	revisions, err := r.getRevisions(configuration)
	if err != nil {
		if !kerrors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
	}

	route := &knserving.Route{}
	err = r.Get(ctx, req.NamespacedName, route)
	if err != nil {
		if kerrors.IsNotFound(err) {
			log.Info("Route not found")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Unable to get knative route")
		return ctrl.Result{}, err
	}

	status, err := createStatusFromKnative(configuration, route, revisions)
	if err != nil {
		log.Error(err, "Unable to determine configuration status")
		return ctrl.Result{}, err
	}

	statusCode, err := r.RiserClient.Deployments.SaveStatus(req.Name, req.Namespace, r.Config.Environment, status)
	return r.handleDeploymentsSaveStatusResult(log, status.ObservedRiserRevision, statusCode, err)
}

func (r *KNativeReconciler) handleDeploymentsSaveStatusResult(log logr.Logger, observedRiserRevision int64, statusCode int, err error) (ctrl.Result, error) {
	if err == nil {
		log.Info("Saved deployment status", "observedRiserRevision", observedRiserRevision)
	} else {
		if statusCode == http.StatusConflict {
			log.Error(err, "Error saving deployment status: conflict", "observedRiserRevision", observedRiserRevision)
			return ctrl.Result{}, nil
		}
		log.Error(err, "Error saving deployment status", "observedRiserRevision", observedRiserRevision)
		return ctrl.Result{Requeue: true}, err
	}
	return ctrl.Result{}, nil
}

func (r *KNativeReconciler) getRevisions(kcfg *knserving.Configuration) ([]knserving.Revision, error) {
	revisionList := &knserving.RevisionList{}
	// Filtering on app label works but ownerReference is probably the more correct approach.
	// Couldn't quickly find how to do that so sticking with the label filter for now.
	err := r.List(context.Background(), revisionList, client.InNamespace(kcfg.Namespace), riserAppFilter(kcfg.ObjectMeta))
	if err != nil {
		return nil, errors.Wrap(err, "error listing revisions")
	}

	return revisionList.Items, nil
}

func createStatusFromKnative(kcfg *knserving.Configuration, route *knserving.Route, revisions []knserving.Revision) (*model.DeploymentStatusMutable, error) {
	// TODO: check route revision and warn when there's a conflict, or consider not updating status at all
	observedRiserRevision, err := getRiserRevision(kcfg.ObjectMeta)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Error getting riser revision for knative configuration %q", kcfg.Name))
	}

	riserStatus := &model.DeploymentStatusMutable{
		ObservedRiserRevision:     observedRiserRevision,
		LatestCreatedRevisionName: kcfg.Status.LatestCreatedRevisionName,
		LatestReadyRevisionName:   kcfg.Status.LatestReadyRevisionName,
	}

	riserStatus.Revisions = make([]model.DeploymentRevisionStatus, len(revisions))
	for idx, revision := range revisions {
		dockerImage, err := getAppDockerImageFromKnativeRevision(&revision)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to get docker image")
		}
		revisionGen, err := getRiserRevision(revision.ObjectMeta)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("Error getting riser revision for revision %q", revision.ObjectMeta.Name))
		}

		revisionStatus := status.GetRevisionStatus(&revision)

		riserStatus.Revisions[idx] = model.DeploymentRevisionStatus{
			Name:                 revision.Name,
			DockerImage:          dockerImage,
			RiserRevision:        revisionGen,
			RevisionStatus:       revisionStatus.Status,
			RevisionStatusReason: revisionStatus.Reason,
		}
	}

	riserStatus.Traffic = make([]model.DeploymentTrafficStatus, len(route.Status.Traffic))
	for idx, traffic := range route.Status.Traffic {
		riserStatus.Traffic[idx] = model.DeploymentTrafficStatus{
			RevisionName: traffic.RevisionName,
			Percent:      traffic.Percent,
			Tag:          traffic.Tag,
		}
	}
	return riserStatus, nil
}

func getAppDockerImageFromKnativeRevision(revision *knserving.Revision) (string, error) {
	riserDeployment := revision.Labels[riserLabel("deployment")]
	for _, container := range revision.Spec.Containers {
		if container.Name == riserDeployment {
			return container.Image, nil
		}
	}
	return "", fmt.Errorf("Unable to find a container matching the deployment %q", riserDeployment)
}

func createUpdateRiserFilter() predicate.Funcs {
	return predicate.Funcs{
		CreateFunc: func(evt event.CreateEvent) bool {
			return isRiserApp(evt.Meta)
		},
		UpdateFunc: func(evt event.UpdateEvent) bool {
			return isRiserApp(evt.MetaNew)
		},
	}
}
