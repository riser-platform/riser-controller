package controllers

import (
	"context"
	"fmt"
	"riser-controller/pkg/runtime"
	"time"

	"k8s.io/apimachinery/pkg/types"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/riser-platform/riser/sdk"

	knserving "knative.dev/serving/pkg/apis/serving/v1"
)

type PodReconciler struct {
	client.Client
	Log         logr.Logger
	Config      runtime.Config
	RiserClient *sdk.Client
}

// HACK: There's got to be a better way!
// THere are scenarios where a pod becomes unhealthy but does not trigger a knative service update, in which case we do not update the pod status
// with riser. This forces a reconcile by mutating the knative service.
func (r *PodReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("pod", req.NamespacedName)

	pod := &corev1.Pod{}

	err := r.Get(ctx, req.NamespacedName, pod)
	if err != nil {
		log.Error(err, "Unable to fetch pod")
		return ctrl.Result{}, err
	}

	if isRiserApp(pod.ObjectMeta) {
		log.Info("Pod reconcile", "pod", pod.Name, "namespace", pod.Namespace)
		service := &knserving.Service{}
		err := r.Get(context.Background(), types.NamespacedName{Name: pod.Labels["serving.knative.dev/service"], Namespace: req.Namespace}, service)
		if err != nil {
			log.Error(err, "Error fetching service for pod")
			return ctrl.Result{}, err
		}

		service.Annotations[riserLabel("controller-observed")] = fmt.Sprint(time.Now().Unix())
		err = r.Update(ctx, service)
		if err != nil {
			log.Error(err, "Error updating service", "service", service.Name)
		}
	}

	return ctrl.Result{}, nil
}

func (r *PodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{}).
		// If a pod is deleted or created there's no need to trigger as the deployment reconciliation will be triggerred
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
