package controllers

import (
	"context"
	"fmt"
	"riser-controller/pkg/runtime"
	"time"

	appsv1 "k8s.io/api/apps/v1"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/riser-platform/riser/sdk"
)

type PodReconciler struct {
	client.Client
	Log         logr.Logger
	Config      runtime.Config
	RiserClient *sdk.Client
}

// HACK: There's got to be a better way!
// THere are scenarios where a pod becomes unhealthy but does not trigger a Deployment reconcile, in which case we do not update the pod status
// with riser. This forces a deployment reconcile by mutating the deployment.
func (r *PodReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("pod", req.NamespacedName)

	pod := &corev1.Pod{}

	err := r.Get(ctx, req.NamespacedName, pod)
	if err != nil {
		log.Error(err, "Unable to fetch pod")
		return ctrl.Result{}, err
	}

	_, isRiserApp := pod.Labels[RiserAppLabel]
	if isRiserApp {
		log.Info("Pod reconcile", "pod", pod.Name, "namespace", pod.Namespace)
		deployments := &appsv1.DeploymentList{}
		labels := map[string]string{
			"deployment": pod.Labels["deployment"],
		}
		err := r.List(context.Background(), deployments, client.InNamespace(pod.Namespace), client.MatchingLabels(labels))
		if err != nil {
			log.Error(err, "Error fetching deployment for pod")
			return ctrl.Result{}, err
		}

		if len(deployments.Items) != 1 {
			log.Error(nil, fmt.Sprintf("Unexpected number of deployments for pod %s. Found %d expected 1", pod.Name, len(deployments.Items)))
			return ctrl.Result{}, err
		}

		deployment := deployments.Items[0]
		deployment.Annotations["riser-trigger-hack"] = fmt.Sprint(time.Now().Unix())
		err = r.Update(ctx, &deployment)
		if err != nil {
			log.Error(err, "Error updating deployment", "deployment", deployment.Name)
		}
		log.Info("Deployment for pod found", "deployment", deployment.Name, "pod", pod.Name, "namespace", pod.Namespace)
	}

	return ctrl.Result{}, nil
}

func (r *PodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{}).
		// If a pod is deleted or created there's no need to trigger as the deployment reconcilation will be triggerd
		WithEventFilter(predicate.Funcs{
			CreateFunc: func(event.CreateEvent) bool {
				return false
			},
			DeleteFunc: func(event.DeleteEvent) bool {
				return false
			},
		}).
		Complete(r)
}
