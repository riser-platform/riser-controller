package controllers

import (
	// TODO: Rename pkg/status
	"context"
	"fmt"
	"riser-controller/pkg/runtime"
	probe "riser-controller/pkg/status"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/pkg/errors"

	"github.com/riser-platform/riser-server/api/v1/model"

	"github.com/go-logr/logr"
	"github.com/riser-platform/riser/sdk"
	knserving "knative.dev/serving/pkg/apis/serving/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
)

type KNativeServiceReconciler struct {
	client.Client
	Log         logr.Logger
	Config      runtime.Config
	RiserClient *sdk.Client
}

type revisionDeployment struct {
	knserving.Revision
	Deployment *appsv1.Deployment
}

func (r *KNativeServiceReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("service", req.NamespacedName)

	service := &knserving.Service{}

	err := r.Get(ctx, req.NamespacedName, service)
	if err != nil {
		log.Error(err, "Unable to get KNative service")
		return ctrl.Result{}, err
	}
	if isRiserApp(service.ObjectMeta) {
		revisions, err := r.getRevisions(service)
		if err != nil {
			log.Error(err, "Unable to list revisions")
			return ctrl.Result{}, err
		}
		status, err := createStatusFromKnativeSvc(service, revisions)
		if err != nil {
			return ctrl.Result{}, err
		}

		err = r.RiserClient.Deployments.SaveStatus(req.Name, r.Config.Stage, status)
		if err == nil {
			log.Info("Updated deployment status", "riserGeneration", status.ObservedRiserGeneration)
		} else {
			log.Error(err, "Error saving status")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *KNativeServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&knserving.Service{}).
		Complete(r)
}

func (r *KNativeServiceReconciler) getRevisions(ksvc *knserving.Service) ([]revisionDeployment, error) {
	revisions := &knserving.RevisionList{}
	// Filtering on app label works but ownerReference is probably the more correct approach.
	// Couldn't quickly find how to do that so sticking with the label filter for now.
	err := r.List(context.Background(), revisions, client.InNamespace(ksvc.Namespace), riserAppFilter(ksvc.ObjectMeta))
	if err != nil {
		return nil, errors.Wrap(err, "error listing revisions")
	}

	revisionDeployments := []revisionDeployment{}
	for _, revision := range revisions.Items {
		deployment, err := r.getDeployment(&revision)
		if err != nil {
			if kerrors.IsNotFound(err) {
				r.Log.Info("Deployment not found for revision", "revision", revision.Name)
			} else {
				return nil, errors.Wrap(err, "error getting deployment for revision")
			}
		}
		revisionDeployments = append(revisionDeployments, revisionDeployment{
			Revision:   revision,
			Deployment: deployment,
		})
	}

	return revisionDeployments, nil
}

func createStatusFromKnativeSvc(ksvc *knserving.Service, revisions []revisionDeployment) (*model.DeploymentStatusMutable, error) {
	observedRiserGeneration, err := getRiserGeneration(ksvc.ObjectMeta)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Error getting riser generation for knative service %q", ksvc.Name))
	}

	status := &model.DeploymentStatusMutable{
		ObservedRiserGeneration:   observedRiserGeneration,
		LatestCreatedRevisionName: ksvc.Status.LatestCreatedRevisionName,
		LatestReadyRevisionName:   ksvc.Status.LatestReadyRevisionName,
	}

	status.Revisions = make([]model.DeploymentRevisionStatus, len(revisions))
	for idx, revision := range revisions {
		dockerImage, err := getAppDockerImageFromKnativeRevision(&revision.Revision)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to get docker image")
		}
		revisionGen, err := getRiserGeneration(revision.ObjectMeta)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("Error getting riser generation for revision %q", revision.ObjectMeta.Name))
		}

		rolloutStatus := probe.GetRevisionStatus(&revision.Revision, revision.Deployment)
		status.Revisions[idx] = model.DeploymentRevisionStatus{
			Name:                revision.Name,
			AvailableReplicas:   getAvailableReplicasFromDeployment(revision.Deployment),
			DockerImage:         dockerImage,
			RiserGeneration:     revisionGen,
			RolloutStatus:       rolloutStatus.Status,
			RolloutStatusReason: rolloutStatus.Reason,
		}
	}

	status.Traffic = make([]model.DeploymentTrafficStatus, len(ksvc.Status.Traffic))
	for idx, traffic := range ksvc.Status.Traffic {
		status.Traffic[idx] = model.DeploymentTrafficStatus{
			RevisionName: traffic.RevisionName,
			Percent:      traffic.Percent,
			Latest:       traffic.LatestRevision,
		}
	}
	return status, nil
}

func getAvailableReplicasFromDeployment(deployment *appsv1.Deployment) int32 {
	if deployment == nil {
		return 0
	}

	return deployment.Status.AvailableReplicas
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

func (r *KNativeServiceReconciler) getDeployment(revision *knserving.Revision) (*appsv1.Deployment, error) {
	deployment := &appsv1.Deployment{}
	err := r.Get(context.Background(), types.NamespacedName{Name: fmt.Sprintf("%s-deployment", revision.Name), Namespace: revision.Namespace}, deployment)
	if err != nil {
		return nil, err
	}

	return deployment, nil
}
