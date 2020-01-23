package controllers

import (
	// TODO: Rename pkg/status
	"context"
	"fmt"
	"riser-controller/pkg/runtime"
	probe "riser-controller/pkg/status"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
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

// TODO: Rename to revisionGraph?
type revisionEtc struct {
	knserving.Revision
	Deployment *appsv1.Deployment
	Pods       *corev1.PodList
}

// This arguably breaks the prescriptive reconcile pattern of one type per reconcile. Similar to a knative Service, we treat
// Configuration+Route as a single status entry in Riser. Options are:
// - Have separate status endpoints in the riser API for config and route status and update the statuses independently
// - Use a knative Service: this is problematic with the gitops pattern because the lifecycles are different for each resource
// - Keep doing what we're doing if there's no practical side effects (<- likely the right answer)
func (r *KNativeReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("knative", req.NamespacedName)

	configuration := &knserving.Configuration{}

	err := r.Get(ctx, req.NamespacedName, configuration)
	if err != nil {
		log.Error(err, "Unable to get knative configuration")
		return ctrl.Result{}, err
	}

	if isRiserApp(configuration.ObjectMeta) {
		revisions, err := r.getRevisions(configuration)
		if err != nil {
			log.Error(err, "Unable to list revisions")
			return ctrl.Result{}, err
		}

		route := &knserving.Route{}
		err = r.Get(ctx, req.NamespacedName, route)
		if err != nil {
			log.Error(err, "Unable to get knative route")
			return ctrl.Result{}, err
		}

		status, err := createStatusFromKnative(configuration, route, revisions)
		if err != nil {
			log.Error(err, "Unable to determine configuration status")
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

func (r *KNativeConfigurationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&knserving.Configuration{}).
		Complete(r)
}

func (r *KNativeRouteReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&knserving.Route{}).
		Complete(r)
}

func (r *KNativeReconciler) getRevisions(kcfg *knserving.Configuration) ([]revisionEtc, error) {
	revisionList := &knserving.RevisionList{}
	// Filtering on app label works but ownerReference is probably the more correct approach.
	// Couldn't quickly find how to do that so sticking with the label filter for now.
	err := r.List(context.Background(), revisionList, client.InNamespace(kcfg.Namespace), riserAppFilter(kcfg.ObjectMeta))
	if err != nil {
		return nil, errors.Wrap(err, "error listing revisions")
	}

	revisions := []revisionEtc{}
	for _, revision := range revisionList.Items {
		deployment, err := r.getDeployment(&revision)
		if err != nil {
			if kerrors.IsNotFound(err) {
				r.Log.Info("Deployment not found for revision", "revision", revision.Name)
			} else {
				return nil, errors.Wrap(err, "error getting deployment for revision")
			}
		}
		pods := &corev1.PodList{}
		if deployment != nil {
			pods, err = r.getPodsForRevision(&revision)
			if err != nil {
				return nil, errors.Wrap(err, "error getting pods for deployment")
			}
		}
		revisions = append(revisions, revisionEtc{
			Revision:   revision,
			Deployment: deployment,
			Pods:       pods,
		})
	}

	return revisions, nil
}

func createStatusFromKnative(kcfg *knserving.Configuration, route *knserving.Route, revisions []revisionEtc) (*model.DeploymentStatusMutable, error) {
	// TODO: check route generation and warn when there's a conflict, or consider not updating status at all
	observedRiserGeneration, err := getRiserGeneration(kcfg.ObjectMeta)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Error getting riser generation for knative service %q", kcfg.Name))
	}

	status := &model.DeploymentStatusMutable{
		ObservedRiserGeneration:   observedRiserGeneration,
		LatestCreatedRevisionName: kcfg.Status.LatestCreatedRevisionName,
		LatestReadyRevisionName:   kcfg.Status.LatestReadyRevisionName,
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

		rolloutStatus := probe.GetRevisionStatus(&revision.Revision)
		problems := probe.GetPodProblems(revision.Pods)
		status.Revisions[idx] = model.DeploymentRevisionStatus{
			Name:                revision.Name,
			AvailableReplicas:   getAvailableReplicasFromDeployment(revision.Deployment),
			DockerImage:         dockerImage,
			RiserGeneration:     revisionGen,
			RolloutStatus:       rolloutStatus.Status,
			RolloutStatusReason: rolloutStatus.Reason,
			Problems:            problems.Items(),
		}
	}

	status.Traffic = make([]model.DeploymentTrafficStatus, len(route.Status.Traffic))
	for idx, traffic := range route.Status.Traffic {
		status.Traffic[idx] = model.DeploymentTrafficStatus{
			RevisionName: traffic.RevisionName,
			Percent:      traffic.Percent,
			Tag:          traffic.Tag,
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

func (r *KNativeReconciler) getPodsForRevision(revision *knserving.Revision) (*corev1.PodList, error) {
	pods := &corev1.PodList{}
	labels := client.MatchingLabels{
		"serving.knative.dev/revision": revision.Name,
	}
	err := r.List(context.Background(), pods, client.InNamespace(revision.Namespace), labels)
	if err != nil {
		return nil, errors.Wrap(err, "error listing pods")
	}

	return pods, nil
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

func (r *KNativeReconciler) getDeployment(revision *knserving.Revision) (*appsv1.Deployment, error) {
	deployment := &appsv1.Deployment{}
	err := r.Get(context.Background(), types.NamespacedName{Name: fmt.Sprintf("%s-deployment", revision.Name), Namespace: revision.Namespace}, deployment)
	if err != nil {
		return nil, err
	}

	return deployment, nil
}
