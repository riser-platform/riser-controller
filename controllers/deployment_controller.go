/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"riser-controller/pkg/runtime"
	"riser-controller/pkg/status"
	"strconv"

	"github.com/pkg/errors"

	corev1 "k8s.io/api/core/v1"

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/riser-platform/riser-server/api/v1/model"
	"github.com/riser-platform/riser/sdk"
	appsv1 "k8s.io/api/apps/v1"
)

// DeploymentReconciler reconciles a Deployment object
type DeploymentReconciler struct {
	client.Client
	Log         logr.Logger
	Config      runtime.Config
	RiserClient *sdk.Client
}

// Note: For some reason the generator is not working. Must manually update config/rbac/role.yaml
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get
// +kubebuilder:rbac:groups=apps,resources=pods,verbs=get;list
func (r *DeploymentReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("deployment", req.NamespacedName)

	deployment := &appsv1.Deployment{}

	err := r.Get(ctx, req.NamespacedName, deployment)
	if err != nil {
		log.Error(err, "Unable to fetch deployment")
		return ctrl.Result{}, err
	}

	if isRiserApp(deployment.ObjectMeta) {
		pods, err := r.getPodsForDeployment(deployment)
		if err != nil {
			log.Error(err, "Unable to get pods for deployment")
			return ctrl.Result{}, err
		}

		rolloutStatus := status.GetRolloutStatus(deployment)
		problems := status.GetPodProblems(pods)

		revision, _ := strconv.ParseInt(deployment.Annotations["deployment.kubernetes.io/revision"], 10, 64)
		riserGeneration, _ := strconv.ParseInt(deployment.Annotations[riserLabel("generation")], 10, 64)
		status := &model.DeploymentStatusMutable{
			ObservedRiserGeneration: riserGeneration,
			RolloutStatus:           rolloutStatus.Status,
			RolloutStatusReason:     rolloutStatus.Reason,
			RolloutRevision:         revision,
			DockerImage:             getAppDockerImage(deployment),
			Problems:                problems.Items(),
		}

		err = r.RiserClient.Deployments.SaveStatus(deployment.Labels[riserLabel("deployment")], deployment.Labels[riserLabel("stage")], status)
		if err != nil {
			log.Error(err, "Unable to update status")
			return ctrl.Result{Requeue: true}, err
		} else {
			log.Info("Updated status", "riserGeneration", riserGeneration)
		}

	}

	return ctrl.Result{}, nil
}

func (r *DeploymentReconciler) getPodsForDeployment(deployment *appsv1.Deployment) (*corev1.PodList, error) {
	pods := &corev1.PodList{}
	err := r.List(context.Background(), pods, client.InNamespace(deployment.Namespace), riserAppFilter(deployment.ObjectMeta))
	if err != nil {
		return nil, errors.Wrap(err, "error listing pods")
	}

	return pods, nil
}

func getAppDockerImage(deployment *appsv1.Deployment) string {
	for _, container := range deployment.Spec.Template.Spec.Containers {
		if container.Name == deployment.Name {
			return container.Image
		}
	}
	return "Unable to find app container"
}

func (r *DeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Deployment{}).
		Complete(r)
}
