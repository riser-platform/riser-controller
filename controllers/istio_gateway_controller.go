package controllers

import (
	"fmt"
	"riser-controller/pkg/runtime"

	"github.com/pkg/errors"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"k8s.io/client-go/dynamic"

	"github.com/go-logr/logr"
	"github.com/riser-platform/riser-server/api/v1/model"
	"github.com/riser-platform/riser/sdk"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
)

const AnnotationRiserGateway = "riser-gateway"
const AnnotationRiserGatewayHost = "riser-gatewayHost"

// IstioGatewayReconciler reconciles a Istio Gateways
type IstioGatewayReconciler struct {
	Log           logr.Logger
	Config        runtime.Config
	RiserClient   *sdk.Client
	DynamicClient dynamic.Interface
}

// +kubebuilder:rbac:groups=networking.istio.io,resources=gateways,verbs=get;list;watch

// TODO: How do we use controller manager and avoid all this plumbing for dynamic types?
// TODO: Refactor and test
func (r *IstioGatewayReconciler) SetupAndStart() error {
	// TODO: Since we only care about labels, this kind of sucks that we have to
	gvr := schema.GroupVersionResource{
		Group:    "networking.istio.io",
		Version:  "v1alpha3",
		Resource: "gateways",
	}
	watcher, err := r.DynamicClient.Resource(gvr).Watch(metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "Error seting up watcher")
	}

	go r.startWatcher(watcher)

	return nil
}

func (r *IstioGatewayReconciler) startWatcher(watcher watch.Interface) {
	ch := watcher.ResultChan()
	for event := range ch {
		if event.Type == watch.Added || event.Type == watch.Modified {
			item := event.Object.(*unstructured.Unstructured)
			annotations := item.GetAnnotations()
			if annotations[AnnotationRiserGateway] == "public-default" && annotations[AnnotationRiserGatewayHost] != "" {
				r.Log.Info(fmt.Sprintf("Found istio gateway %q. Setting as the stage public default gateway with host %q.", item.GetName(), annotations[AnnotationRiserGatewayHost]))
				r.setStageConfig(annotations[AnnotationRiserGatewayHost])
			}
		}
	}
}

func (r *IstioGatewayReconciler) setStageConfig(publicGatewayHost string) {
	err := r.RiserClient.Stages.SetConfig(r.Config.Stage, &model.StageConfig{
		PublicGatewayHost: publicGatewayHost,
	})
	if err != nil {
		r.Log.Error(err, "Error setting stage config, retrying...")
		// Hack: Need to figure out how to use the reconcile pattern as we do with structured resources so we can requeue correctly
		// TODO: Remove recursive call
		r.setStageConfig(publicGatewayHost)
	}
}
