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

package main

import (
	"flag"
	"os"
	"riser-controller/pkg/ping"
	riserruntime "riser-controller/pkg/runtime"
	"riser-controller/pkg/sealedsecret"
	"time"

	"github.com/riser-platform/riser/sdk"

	corev1 "k8s.io/api/core/v1"

	"riser-controller/controllers"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	knserving "knative.dev/serving/pkg/apis/serving/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports
)

// All env vars are prefixed with RISER_
const envPrefix = "RISER"

// DotEnv file typically used For local development
const dotEnvFile = ".env"

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	err := appsv1.AddToScheme(scheme)
	exitIfError(err, "appsv1")
	err = corev1.AddToScheme(scheme)
	exitIfError(err, "corev1")
	err = knserving.AddToScheme(scheme)
	exitIfError(err, "knserving")
	// +kubebuilder:scaffold:scheme
}

func main() {
	// TODO: Even those these are standard switches for a kubebuilder controller, should consider moving to env vars (RuntimeConfiguration)
	var metricsAddr string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	ctrl.SetLogger(zap.Logger(true))

	err := loadDotEnv()
	exitIfError(err, "Error loading .env file")

	var rc riserruntime.Config
	err = envconfig.Process(envPrefix, &rc)
	exitIfError(err, "Error loading environment variables")

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		LeaderElection:     enableLeaderElection,
	})
	exitIfError(err, "unable to start manager")

	riserClient, err := sdk.NewClient(rc.ServerURL, rc.ServerApikey)
	exitIfError(err, "Unable to initialize riser client")

	serverPingDuration := time.Second * time.Duration(rc.ServerPingSeconds)
	ping.StartNewPinger(riserClient, ctrl.Log.WithName("pinger"), rc.Stage, serverPingDuration)

	sealedSecretRefreshDuration, err := time.ParseDuration(rc.SealedsecretCertRefreshDuration)
	exitIfError(err, "Unable to parse sealed secret cert refresh duration")

	if rc.SealedSecretEnabled {
		err = sealedsecret.StartCertRefresher(
			ctrl.GetConfigOrDie(),
			riserClient,
			rc.Stage,
			rc.SealedsecretControllerName,
			rc.SealedsecretNamespace,
			sealedSecretRefreshDuration,
			ctrl.Log.WithName("sealedsecret").WithName("refresher"),
		)
		exitIfError(err, "Unable to start sealed secret cert refresher")
	}

	err = (&controllers.KNativeConfigurationReconciler{
		KNativeReconciler: controllers.KNativeReconciler{
			Client:      mgr.GetClient(),
			Log:         ctrl.Log.WithName("controllers").WithName("KNativeConfiguration"),
			Config:      rc,
			RiserClient: riserClient,
		},
	}).SetupWithManager(mgr)
	exitIfError(err, "unable to create controller", "controller", "KNativeConfiguration")

	err = (&controllers.KNativeRouteReconciler{
		KNativeReconciler: controllers.KNativeReconciler{
			Client:      mgr.GetClient(),
			Log:         ctrl.Log.WithName("controllers").WithName("KNativeConfiguration"),
			Config:      rc,
			RiserClient: riserClient,
		},
	}).SetupWithManager(mgr)
	exitIfError(err, "unable to create controller", "controller", "KNativeConfiguration")

	err = (&controllers.KNativeDomainReconciler{
		Client:      mgr.GetClient(),
		Log:         ctrl.Log.WithName("controllers").WithName("KNativeDomain"),
		Config:      rc,
		RiserClient: riserClient,
	}).SetupWithManager(mgr)
	exitIfError(err, "unable to create controller", "controller", "KNativeDomain")

	setupLog.Info("starting manager")
	err = mgr.Start(ctrl.SetupSignalHandler())
	exitIfError(err, "problem starting manager")
}

func loadDotEnv() error {
	_, err := os.Stat(dotEnvFile)
	if !os.IsNotExist(err) {
		return godotenv.Load(dotEnvFile)
	}

	return nil
}

func exitIfError(err error, message string, keysAndValues ...interface{}) {
	if err != nil {
		setupLog.Error(err, message, keysAndValues...)
		os.Exit(1)
	}
}
