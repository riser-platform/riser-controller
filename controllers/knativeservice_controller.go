package controllers

import (
	"context"
	"riser-controller/pkg/runtime"

	"github.com/go-logr/logr"
	"github.com/riser-platform/riser/sdk"
	knserving "knative.dev/serving/pkg/apis/serving/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type KNativeServiceReconciler struct {
	client.Client
	Log         logr.Logger
	Config      runtime.Config
	RiserClient *sdk.Client
}

func (r *KNativeServiceReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("service", req.NamespacedName)

	service := &knserving.Service{}

	err := r.Get(ctx, req.NamespacedName, service)
	if err != nil {
		log.Error(err, "Unable to fetch KNative service")
		return ctrl.Result{}, err
	}

	if isRiserApp(service.ObjectMeta) {
		log.Info("IsRiserApp!")
	}

	return ctrl.Result{}, nil
}

func (r *KNativeServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&knserving.Service{}).
		Complete(r)
}
