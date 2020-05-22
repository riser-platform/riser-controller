package controllers

import (
	"context"
	"fmt"
	"riser-controller/pkg/runtime"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/go-logr/logr"
	"github.com/riser-platform/riser-server/api/v1/model"
	"github.com/riser-platform/riser-server/pkg/sdk"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const DomainConfigName = "config-domain"
const KNativeServingNamespace = "knative-serving"

type KNativeDomainReconciler struct {
	client.Client
	Log         logr.Logger
	Config      runtime.Config
	RiserClient *sdk.Client
}

func (r *KNativeDomainReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("knative domain config", req.NamespacedName)

	cm := &corev1.ConfigMap{}

	err := r.Get(ctx, req.NamespacedName, cm)
	if err != nil {
		log.Error(err, "Unable to get configmap")
		return ctrl.Result{}, err
	}

	for key := range cm.Data {
		// Select the first key that does not start with an underscore (e.g. "_example").
		// This is an intentional Riser limitation for now (single domain per environment/cluster). In the future we should read the selector
		// with a riser specific label
		if key[:1] != "_" {
			log.Info(fmt.Sprintf("Found custom domain %q. Updating environment config...", key))
			err = r.RiserClient.Environments.SetConfig(r.Config.Environment, &model.EnvironmentConfig{
				PublicGatewayHost: key,
			})
			if err != nil {
				return ctrl.Result{Requeue: true}, err
			}
		}
	}

	return ctrl.Result{}, nil
}

func (r *KNativeDomainReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.ConfigMap{}).
		WithEventFilter(predicate.Funcs{
			CreateFunc: func(evt event.CreateEvent) bool {
				return filterDomainConfigMap(evt.Meta)
			},
			UpdateFunc: func(evt event.UpdateEvent) bool {
				return filterDomainConfigMap(evt.MetaNew)
			},
		}).
		Complete(r)
}

func filterDomainConfigMap(meta metav1.Object) bool {
	return meta.GetNamespace() == KNativeServingNamespace && meta.GetName() == DomainConfigName
}
